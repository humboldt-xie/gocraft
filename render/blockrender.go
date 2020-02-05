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
	"github.com/humboldt-xie/tinycraft/world"
)

type BlockRender struct {
	world   *world.World
	win     *glfw.Window
	shader  *glhf.Shader
	texture *glhf.Texture

	facePool *sync.Pool

	sigch     chan bool
	meshcache sync.Map //map[Vec3]*ChunkMesh

	stat Stat
	text *Text

	item *Mesh
}

func NewBlockRender(win *glfw.Window, world *world.World) (*BlockRender, error) {
	var (
		err error
	)
	img, rect, err := loadImage(*texturePath)
	if err != nil {
		return nil, err
	}

	r := &BlockRender{
		world: world,
		win:   win,
		sigch: make(chan bool, 4),
	}

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
		r.text = &Text{shader: r.shader}
		r.text.UpdateTexture("")
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

	return r, nil
}

func ShowFaces(world *world.World, id Vec3) [6]bool {
	return [...]bool{
		world.Block(id.Left()).IsTransparent(),
		world.Block(id.Right()).IsTransparent(),
		world.Block(id.Up()).IsTransparent(),
		world.Block(id.Down()).IsTransparent() && world.Block(id.Down()) != nil, //&& id.Y != 0
		world.Block(id.Front()).IsTransparent(),
		world.Block(id.Back()).IsTransparent(),
	}
}
func makeBlock(world *world.World, vertices []float32, w *Block, id Vec3) []float32 {
	//pos := game.camera.Pos()
	/*show = [...]bool{
		true, true, true, true, true, true,
	}*/
	show := ShowFaces(world, id)
	vertices = makeData(w, vertices, show, id)
	return vertices
}

func (r *BlockRender) makeChunkMesh(c *world.Chunk, onmainthread bool) *Mesh {
	start := time.Now()
	defer func() {
		log.Printf("make chunk spend %fs %v", float64(time.Since(start))/float64(time.Second), c.Id())
	}()
	facedata := r.facePool.Get().([]float32)
	defer r.facePool.Put(facedata[:0])

	c.RangeBlocks(func(id Vec3, w *Block) {
		facedata = makeBlock(r.world, facedata, w, id)
	})
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

	show := [...]bool{true, true, true, true, true, true}
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
func (r *BlockRender) get3dmat(player *world.Player) mgl32.Mat4 {
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
	mat := r.get3dmat(player)
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

func (r *BlockRender) updateMeshCache(player *world.Player) {
	block := world.NearBlock(player.Pos())
	chunk := block.Chunkid()
	x, z := chunk.X, chunk.Z
	n := *RenderRadius
	needed := make(map[Vec3]bool)
	//log.Printf("updateMeshCache %v %d", block, n)

	for dx := -n; dx < n; dx++ {
		for dz := -n; dz < n; dz++ {
			id := Vec3{x + dx, 0, z + dz}
			if dx*dx+dz*dz > n*n {
				continue
			}
			needed[id] = true
		}
	}

	var added, removed []Vec3
	r.meshcache.Range(func(k, v interface{}) bool {
		id := k.(Vec3)
		if !needed[id] {
			removed = append(removed, id)
		}
		return true
	})

	for id := range needed {
		_, ok := r.meshcache.Load(id)
		// 不在cache里面的需要重新构建
		if !ok {
			added = append(added, id)
		}
	}
	// 单次并发构造的chunk个数
	const batchBuildChunk = 16
	r.sortChunks(player, added)
	if len(added) > batchBuildChunk {
		added = added[:batchBuildChunk]
	}

	for _, id := range removed {
		log.Printf("remove cache %v", id)
		mesh, _ := r.meshcache.Load(id)
		r.meshcache.Delete(id)
		mesh.(*ChunkMesh).Close()
	}

	start := time.Now()
	//newChunks := game.world.Chunks(added)
	group := sync.WaitGroup{}
	for _, id := range added {
		r.world.Chunk(id)
		group.Add(1)
		go func(id Vec3) {
			defer group.Done()
			log.Printf("add cache %v", id)
			r.meshcache.Store(id, NewChunkMesh(r.world, r, id))
		}(id)
	}
	group.Wait()
	if len(added) > 0 {
		log.Printf("make chunks spend %fs %d", float64(time.Since(start))/float64(time.Second), len(added))
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

func (r *BlockRender) checkChunks() {
	// nonblock signal
	select {
	case r.sigch <- true:
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
	mesh, ok := r.meshcache.Load(id)
	if !ok {
		return
	}
	mesh.(*ChunkMesh).DirtyChunk()
}

func (r *BlockRender) UpdateLoop(player *world.Player) {
	for {
		select {
		case <-r.sigch:
		}
		r.updateMeshCache(player)
	}
}

func (r *BlockRender) drawChunks(player *world.Player) {
	r.forcePlayerChunks(player)
	r.checkChunks()
	mat := r.get3dmat(player)

	r.shader.SetUniformAttr(0, mat)
	r.shader.SetUniformAttr(1, player.Pos())
	r.shader.SetUniformAttr(2, float32(*RenderRadius)*world.ChunkWidth)

	r.stat = Stat{}
	planes := frustumPlanes(&mat)

	block := world.NearBlock(player.Pos())
	chunk := block.Chunkid()
	x, z := chunk.X, chunk.Z
	n := *RenderRadius
	//var info = fmt.Sprintf("pos (%v) chunk (%v)\n", block, chunk)
	for dx := -n + 1; dx < n; dx++ {
		for dz := -n + 1; dz < n; dz++ {
			id := Vec3{x + dx, 0, z + dz}
			if dx*dx+dz*dz > n*n {
				continue
			}
			if v, ok := r.meshcache.Load(id); ok {
				chunk := r.world.Chunk(id)
				//info += fmt.Sprintf("(%d,%d)", id.X, id.Z)
				cmesh := v.(*ChunkMesh)
				mesh := cmesh.mesh
				if chunk.V() != cmesh.version {
					cmesh.checkChunk()
				}
				r.stat.CacheChunks++
				if isChunkVisiable(planes, id) {
					//info += fmt.Sprintf("e[%d]\t", mesh.Faces())
					r.stat.RendingChunks++
					r.stat.Faces += mesh.Faces()
					mesh.Draw()
				}
			}
		}
		//info += "\n"
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
