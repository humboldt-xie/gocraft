package render

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	lru "github.com/hashicorp/golang-lru"
	"github.com/humboldt-xie/tinycraft/world"
)

type MuCache struct {
	mu  sync.Mutex
	lru *lru.Cache
}
type Evicted func(key interface{}, value interface{})

func NewMuCache(entry int, onEvicted Evicted) *MuCache {
	glru, _ := lru.NewWithEvict(entry, onEvicted)
	return &MuCache{lru: glru}
}

func (c *MuCache) Add(key interface{}, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Add(key, value)
}

func (c *MuCache) Get(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Get(key)
}

type BlockRender struct {
	world   *world.World
	player  *world.Player
	win     *glfw.Window
	shader  *glhf.Shader
	texture *glhf.Texture

	facePool *sync.Pool

	sigch chan Vec3
	//meshcache sync.Map //map[Vec3]*ChunkMesh
	meshcache *MuCache

	stat Stat
	text *Text

	item *Mesh
}

func NewBlockRender(win *glfw.Window, world *world.World, player *world.Player) (*BlockRender, error) {
	var err error
	img, rect, err := LoadImage(*texturePath)
	if err != nil {
		return nil, err
	}

	//img = dst

	r := &BlockRender{
		world:  world,
		player: player,
		win:    win,
		sigch:  make(chan Vec3, 8),
	}

	n := *RenderRadius * 2
	r.meshcache = NewMuCache(n*n*4, r.OnEvicted)

	mainthread.Call(func() {
		r.shader, err = glhf.NewShader(glhf.AttrFormat{
			glhf.Attr{Name: "pos", Type: glhf.Vec3},
			glhf.Attr{Name: "tex", Type: glhf.Vec2},
			glhf.Attr{Name: "normal", Type: glhf.Vec3},
		}, glhf.AttrFormat{
			glhf.Attr{Name: "matrix", Type: glhf.Mat4},
			glhf.Attr{Name: "camera", Type: glhf.Vec3},
			glhf.Attr{Name: "fogdis", Type: glhf.Float},
		}, blockVertexSource, blockFragmentSource)
		r.text = NewText(r.shader) //&Text{shader: r.shader}
		r.text.Update("欢迎光临")
		if err != nil {
			return
		}
		r.texture = glhf.NewTexture(rect.Dx(), rect.Dy(), false, img.Pix)
	})
	if err != nil {
		return nil, err
	}
	r.facePool = &sync.Pool{
		New: func() interface{} {
			size := 500000 //r.shader.VertexFormat().Size() * 4 * 6 * 6
			log.Printf("new face buffer %d", size)
			return make([]float32, 0, size)
		},
	}
	go r.UpdateLoop(r.player)

	return r, nil
}

func ShowFaces(world *world.World, id Vec3) FaceFilter {
	return FaceFilter{
		Left:  world.Block(id.Left()).IsTransparent(),
		Right: world.Block(id.Right()).IsTransparent(),
		Up:    world.Block(id.Up()).IsTransparent(),
		Down:  world.Block(id.Down()).IsTransparent() && world.Block(id.Down()) != nil, //&& id.Y != 0
		Front: world.Block(id.Front()).IsTransparent(),
		Back:  world.Block(id.Back()).IsTransparent(),
	}
}
func makeBlock(world *world.World, vertices []float32, w *Block, id Vec3) []float32 {
	show := ShowFaces(world, id)
	vertices = makeData(w, vertices, show, id)
	return vertices
}

func (r *BlockRender) UpdateText(s string) {
	r.text.Update(s)
}

func (r *BlockRender) makeChunkMesh(c *world.Chunk, onmainthread bool) *Mesh {
	start := time.Now()
	makeDataSpend := 0.0
	defer func() {
		log.Printf("make chunk spend %.2fs make data spend: %.2fs %v", float64(time.Since(start))/float64(time.Second), makeDataSpend, c.Id())
	}()
	facedata := r.facePool.Get().([]float32)
	defer r.facePool.Put(facedata[:0])
	merge := make(chan []float32, 1024)
	maker := make(chan *Block, 1024)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for faces := range merge {
			facedata = append(facedata, faces...)
		}
	}()

	wgMaker := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wgMaker.Add(1)
		go func() {
			defer wgMaker.Done()
			for b := range maker {
				temp := []float32{}
				temp = makeBlock(r.world, temp, b, b.ID)
				merge <- temp
			}

		}()
	}
	c.RangeBlocks(func(id Vec3, w *Block) {
		/*temp := []float32{}
		temp = makeBlock(r.world, temp, w, id)
		merge <- temp*/
		maker <- w
	})
	close(maker)
	wgMaker.Wait()
	close(merge)
	wg.Wait()
	makeDataSpend = float64(time.Since(start)) / float64(time.Second)
	n := len(facedata) / (r.shader.VertexFormat().Size() / 4)
	log.Printf("chunk faces: %v %d %fs %d", c.Id(), n/6, float64(time.Since(start))/float64(time.Second), len(facedata))
	var mesh *Mesh
	mesh = NewMesh(r.shader, facedata, onmainthread)
	mesh.Id = c.Id()
	return mesh
}

// call on mainthread
func (r *BlockRender) UpdateItem(bt *world.BlockType) {
	vertices := r.facePool.Get().([]float32)
	defer r.facePool.Put(vertices[:0])

	show := FaceFilter{true, true, true, true, true, true}
	pos := Vec3{0, 0, 0}
	w := world.NewBlock(bt.Type)

	vertices = makeData(w, vertices, show, pos)

	item := NewMesh(r.shader, vertices, true)
	if r.item != nil {
		r.item.Release()
	}
	r.item = item
}

// camera
func (r *BlockRender) Get3dmat(player *world.Player) mgl32.Mat4 {
	n := float32(*RenderRadius * world.ChunkWidth)
	width, height := r.win.GetSize()
	mat := mgl32.Perspective(radian(45), float32(width)/float32(height), 0.01, n)
	mat = mat.Mul4(player.Matrix())
	return mat
}

func (r *BlockRender) get2dmat(player *world.Player) mgl32.Mat4 {
	n := float32(*RenderRadius * world.ChunkWidth)
	mat := mgl32.Ortho(-n, n, -n, n, -1, n)
	mat = mat.Mul4(player.Matrix())
	return mat
}

func (r *BlockRender) sortChunks(player *world.Player, chunks []Vec3) []Vec3 {
	cid := world.NearBlock(player.Pos()).Chunkid()
	x, z := cid.X, cid.Z
	mat := r.Get3dmat(player)
	planes := frustumPlanes(&mat)

	sort.Slice(chunks, func(i, j int) bool {
		v1 := isChunkVisiable(planes, chunks[i])
		v2 := isChunkVisiable(planes, chunks[j])
		if v1 && !v2 {
			return true
		}
		if v2 && !v1 {
			return false
		}
		d1 := (chunks[i].X-x)*(chunks[i].X-x) + (chunks[i].Z-z)*(chunks[i].Z-z)
		d2 := (chunks[j].X-x)*(chunks[j].X-x) + (chunks[j].Z-z)*(chunks[j].Z-z)
		return d1 < d2
	})
	return chunks
}

func (r *BlockRender) OnEvicted(key interface{}, value interface{}) {
	log.Printf("onEvicted %v", key)
	value.(*ChunkMesh).Close()
}

func (r *BlockRender) updateMeshCache(player *world.Player, id Vec3) {
	log.Printf("updateMeshCache %v", id)
	if _, ok := r.meshcache.Get(id); !ok {
		r.meshcache.Add(id, NewChunkMesh(r.world, r, id))
	}
}

func (r *BlockRender) forcePlayerChunks(player *world.Player) {
	bid := world.NearBlock(player.Pos())
	cid := bid.Chunkid()
	//var ids []Vec3
	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			id := Vec3{cid.X + dx, 0, cid.Z + dz}
			r.world.Chunk(id)
		}
	}
}

func (r *BlockRender) checkChunk(id Vec3) {
	// nonblock signal
	select {
	case r.sigch <- id:
	default:
	}
}

func (r *BlockRender) DirtyBlock(id Vec3) {
	cid := id.Chunkid()
	r.DirtyChunk(cid)
	neighbors := []Vec3{id.Left(), id.Right(), id.Front(), id.Back()}
	for _, neighbor := range neighbors {
		chunkid := neighbor.Chunkid()
		if chunkid != cid {
			r.DirtyChunk(chunkid)
		}
	}
}

func (r *BlockRender) DirtyChunk(id Vec3) {
	mesh, ok := r.meshcache.Get(id)
	if !ok {
		return
	}
	mesh.(*ChunkMesh).DirtyChunk()
}

func (r *BlockRender) UpdateLoop(player *world.Player) {
	onUpdate := sync.Map{}
	for {
		id := <-r.sigch
		if _, ok := onUpdate.Load(id); !ok {
			onUpdate.Store(id, true)
			go func(id Vec3) {
				r.updateMeshCache(player, id)
				defer onUpdate.Delete(id)
			}(id)
		}
	}
}

func (r *BlockRender) drawChunks(player *world.Player) {
	r.forcePlayerChunks(player)
	//r.checkChunks()
	mat := r.Get3dmat(player)

	r.shader.SetUniformAttr(0, mat)
	r.shader.SetUniformAttr(1, player.Pos())
	r.shader.SetUniformAttr(2, float32(*RenderRadius)*world.ChunkWidth)

	r.stat = Stat{}
	planes := frustumPlanes(&mat)

	block := world.NearBlock(player.Pos())
	chunk := block.Chunkid()
	x, z := chunk.X, chunk.Z
	n := *RenderRadius
	r.stat.CacheChunks = r.meshcache.lru.Len()
	//var info = fmt.Sprintf("pos (%v) chunk (%v)\n", block, chunk)
	needMakeMesh := []Vec3{}
	for dx := -n + 1; dx < n; dx++ {
		for dz := -n + 1; dz < n; dz++ {
			id := Vec3{x + dx, 0, z + dz}
			if dx*dx+dz*dz > n*n {
				continue
			}
			if !isChunkVisiable(planes, id) {
				continue
			}
			//info += fmt.Sprintf("(%d,%d)", id.X, id.Z)
			if v, ok := r.meshcache.Get(id); ok {
				chunk := r.world.Chunk(id)
				cmesh := v.(*ChunkMesh)
				mesh := cmesh.mesh
				if chunk.V() != cmesh.version {
					cmesh.checkChunk()
				}
				//info += fmt.Sprintf("e[%d]\t", mesh.Faces())
				r.stat.RendingChunks++
				r.stat.Faces += mesh.Faces()
				mesh.Draw()
			} else {
				//info += fmt.Sprintf("n\t")
				needMakeMesh = append(needMakeMesh, id)
			}
		}
		//info += "\n"
	}
	//log.Printf("%s %v", info,needAdd)
	r.sortChunks(player, needMakeMesh)
	for _, id := range needMakeMesh {
		r.checkChunk(id)
	}
	//r.stat.Info = info
	//r.text.UpdateTexture(info)
}

func (r *BlockRender) drawItem() {
	if r.item == nil {
		return
	}
	width, height := r.win.GetSize()
	ratio := float32(width) / float32(height)
	projection := mgl32.Ortho2D(0, 15, 0, 15/ratio)
	model := mgl32.Translate3D(1, 1, 0)
	model = model.Mul4(mgl32.HomogRotate3DX(radian(10)))
	model = model.Mul4(mgl32.HomogRotate3DY(radian(45)))
	mat := projection.Mul4(model)
	r.shader.SetUniformAttr(0, mat)
	r.shader.SetUniformAttr(1, mgl32.Vec3{0, 0, 0})
	r.shader.SetUniformAttr(2, float32(*RenderRadius)*world.ChunkWidth)
	r.item.Draw()
}
