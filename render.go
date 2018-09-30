package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	texturePath  = flag.String("t", "texture.png", "texture file")
	renderRadius = flag.Int("r", 6, "render radius")
)

func loadImage(fname string) (*image.RGBA, image.Rectangle, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, image.Rectangle{}, err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)

	bounds := rgba.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r := rgba.At(x, y).(color.RGBA)
			if r.A == 0 {
				r.R = 255
				r.G = 0
				r.B = 255
			}
			rgba.Set(x, y, r)
		}
	}

	return rgba, img.Bounds(), nil
}

type BlockRender struct {
	shader  *glhf.Shader
	texture *glhf.Texture

	facePool *sync.Pool

	sigch     chan bool
	meshcache sync.Map //map[Vec3]*Mesh

	stat Stat
	text *Text

	item *Mesh
}

func NewBlockRender() (*BlockRender, error) {
	var (
		err error
	)
	img, rect, err := loadImage(*texturePath)
	if err != nil {
		return nil, err
	}

	r := &BlockRender{
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
		r.text.UpdateTexture()
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
func makeBlock(vertices []float32, w *Block, id Vec3) []float32 {
	//pos := game.camera.Pos()
	show := [...]bool{
		game.world.Block(id.Left()).IsTransparent() || w.Life != 100,
		game.world.Block(id.Right()).IsTransparent() || w.Life != 100,
		game.world.Block(id.Up()).IsTransparent() || w.Life != 100,
		game.world.Block(id.Down()).IsTransparent() && game.world.Block(id.Down()) != nil, //&& id.Y != 0
		game.world.Block(id.Front()).IsTransparent() || w.Life != 100,
		game.world.Block(id.Back()).IsTransparent() || w.Life != 100,
	}
	vertices = makeData(w, vertices, show, id)
	return vertices
}

func (r *BlockRender) makeChunkMesh(c *Chunk, onmainthread bool) *Mesh {
	start := time.Now()
	defer func() {
		log.Printf("make chunk spend %fs %v", float64(time.Since(start))/float64(time.Second), c.Id())
	}()
	facedata := r.facePool.Get().([]float32)
	defer r.facePool.Put(facedata[:0])

	c.RangeBlocks(func(id Vec3, w *Block) {
		facedata = makeBlock(facedata, w, id)
	})
	n := len(facedata) / (r.shader.VertexFormat().Size() / 4)
	log.Printf("chunk faces: %v %d %fs", c.Id(), n/6, float64(time.Since(start))/float64(time.Second), len(facedata))
	var mesh *Mesh
	mesh = NewMesh(r.shader, facedata, onmainthread)
	mesh.Id = c.Id()
	return mesh
}

// call on mainthread
func (r *BlockRender) UpdateItem(bt *BlockType) {
	vertices := r.facePool.Get().([]float32)
	defer r.facePool.Put(vertices[:0])

	show := [...]bool{true, true, true, true, true, true}
	pos := Vec3{0, 0, 0}
	w := NewBlock(bt.Type)

	vertices = makeData(w, vertices, show, pos)

	item := NewMesh(r.shader, vertices, true)
	if r.item != nil {
		r.item.Release()
	}
	r.item = item
}

func frustumPlanes(mat *mgl32.Mat4) []mgl32.Vec4 {
	c1, c2, c3, c4 := mat.Rows()
	return []mgl32.Vec4{
		c4.Add(c1),          // left
		c4.Sub(c1),          // right
		c4.Sub(c2),          // top
		c4.Add(c2),          // bottom
		c4.Mul(0.1).Add(c3), // front
		c4.Mul(320).Sub(c3), // back
	}
}

func isChunkVisiable(planes []mgl32.Vec4, id Vec3) bool {
	p := mgl32.Vec3{float32(id.X * ChunkWidth), 0, float32(id.Z * ChunkWidth)}
	const m = ChunkWidth
	const max = 1024000

	points := []mgl32.Vec3{
		mgl32.Vec3{p.X(), p.Y() - max, p.Z()},
		mgl32.Vec3{p.X() + m, p.Y() - max, p.Z()},
		mgl32.Vec3{p.X() + m, p.Y() - max, p.Z() + m},
		mgl32.Vec3{p.X(), p.Y() - max, p.Z() + m},

		mgl32.Vec3{p.X(), p.Y() + max, p.Z()},
		mgl32.Vec3{p.X() + m, p.Y() + max, p.Z()},
		mgl32.Vec3{p.X() + m, p.Y() + max, p.Z() + m},
		mgl32.Vec3{p.X(), p.Y() + max, p.Z() + m},
	}
	for _, plane := range planes {
		var in, out int
		for _, point := range points {
			if plane.Dot(point.Vec4(1)) < 0 {
				out++
			} else {
				in++
			}
			if in != 0 && out != 0 {
				break
			}
		}
		if in == 0 {
			return false
		}
	}
	return true
}

// camera
func (r *BlockRender) get3dmat() mgl32.Mat4 {
	n := float32(*renderRadius * ChunkWidth)
	width, height := game.win.GetSize()
	mat := mgl32.Perspective(radian(45), float32(width)/float32(height), 0.01, n)
	mat = mat.Mul4(game.player.Matrix())
	return mat
}

func (r *BlockRender) get2dmat() mgl32.Mat4 {
	n := float32(*renderRadius * ChunkWidth)
	mat := mgl32.Ortho(-n, n, -n, n, -1, n)
	mat = mat.Mul4(game.player.Matrix())
	return mat
}

func (r *BlockRender) sortChunks(chunks []Vec3) []Vec3 {
	cid := NearBlock(game.player.Pos()).Chunkid()
	x, z := cid.X, cid.Z
	mat := r.get3dmat()
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

func (r *BlockRender) updateMeshCache() {
	block := NearBlock(game.player.Pos())
	chunk := block.Chunkid()
	x, z := chunk.X, chunk.Z
	n := *renderRadius
	needed := make(map[Vec3]bool)

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
		mesh, ok := r.meshcache.Load(id)
		// 不在cache里面的需要重新构建
		if !ok {
			added = append(added, id)
		} else {
			if mesh.(*Mesh).Dirty {
				log.Printf("update cache %v", id)
				added = append(added, id)
				removed = append(removed, id)
			}
		}
	}
	// 单次并发构造的chunk个数
	const batchBuildChunk = 4
	r.sortChunks(added)
	if len(added) > batchBuildChunk {
		added = added[:batchBuildChunk]
	}

	var removedMesh []*Mesh
	for _, id := range removed {
		log.Printf("remove cache %v", id)
		mesh, _ := r.meshcache.Load(id)
		r.meshcache.Delete(id)
		removedMesh = append(removedMesh, mesh.(*Mesh))
	}

	start := time.Now()
	newChunks := game.world.Chunks(added)
	group := sync.WaitGroup{}
	for _, c := range newChunks {
		group.Add(1)
		go func(c *Chunk) {
			defer group.Done()
			log.Printf("add cache %v", c.Id())
			r.meshcache.Store(c.Id(), r.makeChunkMesh(c, false))
		}(c)
	}
	group.Wait()
	if len(newChunks) > 0 {
		log.Printf("make chunks spend %fs %d", float64(time.Since(start))/float64(time.Second), len(newChunks))
	}

	mainthread.CallNonBlock(func() {
		for _, mesh := range removedMesh {
			mesh.Release()
		}
	})

}

// called on mainthread
func (r *BlockRender) forceChunks(ids []Vec3) {
	var removedMesh []*Mesh
	chunks := game.world.Chunks(ids)
	for _, chunk := range chunks {
		id := chunk.Id()
		imesh, ok := r.meshcache.Load(id)
		var mesh *Mesh
		if ok {
			mesh = imesh.(*Mesh)
		}
		if ok && !mesh.Dirty {
			continue
		}
		r.meshcache.Store(id, r.makeChunkMesh(chunk, true))
		if ok {
			removedMesh = append(removedMesh, mesh)
		}
	}
	mainthread.CallNonBlock(func() {
		for _, mesh := range removedMesh {
			mesh.Release()
		}
	})
}

func (r *BlockRender) forcePlayerChunks() {
	bid := NearBlock(game.player.Pos())
	cid := bid.Chunkid()
	var ids []Vec3
	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			id := Vec3{cid.X + dx, 0, cid.Z + dz}
			ids = append(ids, id)
		}
	}
	r.forceChunks(ids)
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
	mesh.(*Mesh).Dirty = true
}

func (r *BlockRender) UpdateLoop() {
	for {
		select {
		case <-r.sigch:
		}
		r.updateMeshCache()
	}
}

func (r *BlockRender) drawChunks() {
	r.forcePlayerChunks()
	r.checkChunks()
	mat := r.get3dmat()

	r.shader.SetUniformAttr(0, mat)
	r.shader.SetUniformAttr(1, game.player.Pos())
	r.shader.SetUniformAttr(2, float32(*renderRadius)*ChunkWidth)

	planes := frustumPlanes(&mat)
	r.stat = Stat{}
	r.meshcache.Range(func(k, v interface{}) bool {
		id, mesh := k.(Vec3), v.(*Mesh)
		r.stat.CacheChunks++
		if isChunkVisiable(planes, id) {
			r.stat.RendingChunks++
			r.stat.Faces += mesh.Faces()
			mesh.Draw()
		}
		return true
	})
}

func (r *BlockRender) drawItem() {
	if r.item == nil {
		return
	}
	width, height := game.win.GetSize()
	ratio := float32(width) / float32(height)
	projection := mgl32.Ortho2D(0, 15, 0, 15/ratio)
	model := mgl32.Translate3D(1, 1, 0)
	model = model.Mul4(mgl32.HomogRotate3DX(radian(10)))
	model = model.Mul4(mgl32.HomogRotate3DY(radian(45)))
	mat := projection.Mul4(model)
	r.shader.SetUniformAttr(0, mat)
	r.shader.SetUniformAttr(1, mgl32.Vec3{0, 0, 0})
	r.shader.SetUniformAttr(2, float32(*renderRadius)*ChunkWidth)
	r.item.Draw()
}

type UnicodePage struct {
	rgba *image.RGBA
	rect image.Rectangle
}

func (u *UnicodePage) Draw(r *image.RGBA, p image.Point, index int) {
	ip := image.Point{index % 16 * 16, index / 16 * 16}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			c := u.rgba.At(ip.X+x, ip.Y+y).(color.RGBA)
			r.Set(p.X+x, p.Y+y, c)
		}
	}

}

type Text struct {
	texture *glhf.Texture
	shader  *glhf.Shader
	face    *Mesh
	texts   []string
	rgba    *image.RGBA
	pages   map[int]*UnicodePage
}

func (t *Text) LoadPages() {
	t.pages = make(map[int]*UnicodePage)
	for i := 0; i < 256; i++ {
		img, rect, err := loadImage(fmt.Sprintf("font/unicode_page_%.2x.png", i))
		if err != nil {
			//panic(err)
			continue
		}
		t.pages[i] = &UnicodePage{rgba: img, rect: rect}
	}
}

func (t *Text) UpdateTexture() {
	// 80 one line  256/16 one text
	rect := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{16 * 60, 60 * 16}}
	if t.rgba == nil {
		t.LoadPages()
		t.rgba = image.NewRGBA(rect)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				r := t.rgba.At(x, y).(color.RGBA)
				r.R = 255
				r.G = 0
				r.B = 255
				t.rgba.Set(x, y, r)
			}
		}
	}
	s := "你好啊 hello world"
	for i, w, j := 0, 0, 0; i < len(s); i += w {
		runeValue, width := utf8.DecodeRuneInString(s[i:])
		tidx := int(runeValue % 256)
		pidx := int(runeValue / 256)
		page := t.pages[pidx]
		log.Printf("update textrue %d", j)
		if runeValue != rune(' ') {
			page.Draw(t.rgba, image.Point{j * 16, 10}, tidx)
		}
		w = width
		j += 1
	}

	t.texture = glhf.NewTexture(rect.Dx(), rect.Dy(), false, t.rgba.Pix)
}

func (t *Text) Draw() {
	texture := t.texture
	texture.Begin()
	defer texture.End()
	cubeWeight := float32(1)
	cubeHeight := float32(1)
	if t.face == nil {
		x := float32(0)
		y := float32(0)
		z := float32(0)
		//f := MakeFaceTexture(1)
		f := [6][2]float32{
			{0, 0},
			{1, 0},
			{1, 1},
			{1, 1},
			{0, 1},
			{0, 0},
		}
		vertices := []float32{
			x, y, z, f[0][0], f[0][1], 0, 0, 1,
			x + cubeWeight, y, z, f[1][0], f[1][1], 0, 0, 1,
			x + cubeWeight, y + cubeHeight, z, f[2][0], f[2][1], 0, 0, 1,
			x + cubeWeight, y + cubeHeight, z, f[3][0], f[3][1], 0, 0, 1,
			x, y + cubeHeight, z, f[4][0], f[4][1], 0, 0, 1,
			x, y, z, f[5][0], f[5][1], 0, 0, 1,
		}
		t.face = NewMesh(t.shader, vertices, true)
	}

	projection := mgl32.Ortho2D(0, 1, 0, 1)
	model := mgl32.Translate3D(0, 0, 0)
	mat := projection.Mul4(model)
	t.shader.SetUniformAttr(0, mat)
	t.shader.SetUniformAttr(1, mgl32.Vec3{0, 0, 0})
	t.shader.SetUniformAttr(2, float32(*renderRadius)*ChunkWidth)
	//r.item.Draw()
	t.face.Draw()
}

func (r *BlockRender) drawText() {
	r.text.Draw()
}

func (r *BlockRender) Draw() {
	r.shader.Begin()
	r.texture.Begin()

	r.drawChunks()
	r.drawItem()

	r.shader.End()
	r.texture.End()

	r.shader.Begin()
	r.drawText()
	r.shader.End()
}

type Stat struct {
	Faces         int
	CacheChunks   int
	RendingChunks int
}

func (r *BlockRender) Stat() Stat {
	return r.stat
}

type Lines struct {
	vao, vbo uint32
	shader   *glhf.Shader
	nvertex  int
}

func NewLines(shader *glhf.Shader, data []float32) *Lines {
	l := new(Lines)
	l.shader = shader
	l.nvertex = len(data) / (shader.VertexFormat().Size() / 4)
	gl.GenVertexArrays(1, &l.vao)
	gl.GenBuffers(1, &l.vbo)
	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)

	offset := 0
	for _, attr := range shader.VertexFormat() {
		loc := gl.GetAttribLocation(shader.ID(), gl.Str(attr.Name+"\x00"))
		var size int32
		switch attr.Type {
		case glhf.Float:
			size = 1
		case glhf.Vec2:
			size = 2
		case glhf.Vec3:
			size = 3
		case glhf.Vec4:
			size = 4
		}
		gl.VertexAttribPointer(
			uint32(loc),
			size,
			gl.FLOAT,
			false,
			int32(shader.VertexFormat().Size()),
			gl.PtrOffset(offset),
		)
		gl.EnableVertexAttribArray(uint32(loc))
		offset += attr.Type.Size()
	}
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	return l
}

func (l *Lines) Draw(mat mgl32.Mat4) {
	if l.vao != 0 {
		l.shader.SetUniformAttr(0, mat)
		gl.BindVertexArray(l.vao)
		gl.DrawArrays(gl.LINES, 0, int32(l.nvertex))
		gl.BindVertexArray(0)
	}
}

func (l *Lines) Release() {
	if l.vao != 0 {
		gl.DeleteVertexArrays(1, &l.vao)
		gl.DeleteBuffers(1, &l.vbo)
		l.vao = 0
		l.vbo = 0
	}
}

type LineRender struct {
	shader    *glhf.Shader
	cross     *Lines
	wireFrame *Lines
	lastBlock Vec3
}

func NewLineRender() (*LineRender, error) {
	r := &LineRender{}
	var err error
	mainthread.Call(func() {
		r.shader, err = glhf.NewShader(glhf.AttrFormat{
			glhf.Attr{Name: "pos", Type: glhf.Vec3},
		}, glhf.AttrFormat{
			glhf.Attr{Name: "matrix", Type: glhf.Mat4},
		}, lineVertexSource, lineFragmentSource)

		if err != nil {
			return
		}
		r.cross = makeCross(r.shader)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *LineRender) drawCross() {
	width, height := game.win.GetFramebufferSize()
	project := mgl32.Ortho2D(0, float32(width), float32(height), 0)
	model := mgl32.Translate3D(float32(width/2), float32(height/2), 0)
	model = model.Mul4(mgl32.Scale3D(float32(height/30), float32(height/30), 0))
	r.cross.Draw(project.Mul4(model))
}

func (r *LineRender) drawWireFrame(mat mgl32.Mat4) {
	var vertices []float32
	block, _ := game.world.HitTest(game.player.Pos(), game.player.Front())
	if block == nil {
		return
	}

	mat = mat.Mul4(mgl32.Translate3D(float32(block.X), float32(block.Y), float32(block.Z)))
	mat = mat.Mul4(mgl32.Scale3D(1.06, 1.06, 1.06))
	if *block == r.lastBlock {
		r.wireFrame.Draw(mat)
		return
	}

	id := *block
	show := [...]bool{
		game.world.Block(id.Left()).IsTransparent(),
		game.world.Block(id.Right()).IsTransparent(),
		game.world.Block(id.Up()).IsTransparent(),
		game.world.Block(id.Down()).IsTransparent(),
		game.world.Block(id.Front()).IsTransparent(),
		game.world.Block(id.Back()).IsTransparent(),
	}
	vertices = makeWireFrameData(vertices, show)
	if len(vertices) == 0 {
		return
	}
	r.lastBlock = *block
	if r.wireFrame != nil {
		r.wireFrame.Release()
	}

	r.wireFrame = NewLines(r.shader, vertices)
	r.wireFrame.Draw(mat)
}

func (r *LineRender) Draw() {
	width, height := game.win.GetSize()
	projection := mgl32.Perspective(radian(45), float32(width)/float32(height), 0.01, ChunkWidth*float32(*renderRadius))
	camera := game.player.Matrix()
	mat := projection.Mul4(camera)

	r.shader.Begin()
	r.drawCross()
	r.drawWireFrame(mat)
	r.shader.End()
}

func makeCross(shader *glhf.Shader) *Lines {
	return NewLines(shader, []float32{
		-0.5, 0, 0, 0.5, 0, 0,
		0, -0.5, 0, 0, 0.5, 0,
	})
}
