package render

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"unicode/utf8"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/humboldt-xie/tinycraft/world"
)

var (
	texturePath  = flag.String("t", "texture.png", "texture file")
	RenderRadius = flag.Int("r", 6, "render radius")
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
	p := mgl32.Vec3{float32(id.X * world.ChunkWidth), 0, float32(id.Z * world.ChunkWidth)}
	const m = world.ChunkWidth
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

func (t *Text) UpdateTexture(s string) {
	// 80 one line  256/16 one text
	rect := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{16 * 80, 80 * 16}}
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
	//s := "你好啊 hello world\n hello"
	var runeValue rune
	for i, width, line, j := 0, 0, 0, 0; i < len(s); i += width {
		runeValue, width = utf8.DecodeRuneInString(s[i:])
		if runeValue == '\n' {
			line += 1
			j = 0
			continue
		}
		tidx := int(runeValue % 256)
		pidx := int(runeValue / 256)
		page := t.pages[pidx]
		if runeValue != rune(' ') {
			page.Draw(t.rgba, image.Point{j * 16, 10 + 16*line}, tidx)
		}
		j += 1
	}
	if t.texture != nil {
		t.texture.SetPixels(0, 0, rect.Dx(), rect.Dy(), t.rgba.Pix)
	} else {
		t.texture = glhf.NewTexture(rect.Dx(), rect.Dy(), false, t.rgba.Pix)
	}
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
	t.shader.SetUniformAttr(2, float32(*RenderRadius)*world.ChunkWidth)
	//r.item.Draw()
	t.face.Draw()
}

func (r *BlockRender) drawText() {
	r.text.Draw()
}

func (r *BlockRender) Draw(player *world.Player) {
	r.shader.Begin()
	r.texture.Begin()

	r.drawChunks(player)
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
	Info          string
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
	win       *glfw.Window
	world     *world.World
	shader    *glhf.Shader
	cross     *Lines
	wireFrame *Lines
	lastBlock Vec3
}

func NewLineRender(win *glfw.Window, world *world.World) (*LineRender, error) {
	r := &LineRender{win: win, world: world}
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
	width, height := r.win.GetFramebufferSize()
	project := mgl32.Ortho2D(0, float32(width), float32(height), 0)
	model := mgl32.Translate3D(float32(width/2), float32(height/2), 0)
	model = model.Mul4(mgl32.Scale3D(float32(height/30), float32(height/30), 0))
	r.cross.Draw(project.Mul4(model))
}

func (r *LineRender) drawWireFrame(player *world.Player, mat mgl32.Mat4) {
	if r.world == nil {
		panic("world is nil")
	}
	var vertices []float32
	block, _ := r.world.HitTest(player.Pos(), player.Front())
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
		true || r.world.Block(id.Left()).IsTransparent(),
		true || r.world.Block(id.Right()).IsTransparent(),
		true || r.world.Block(id.Up()).IsTransparent(),
		true || r.world.Block(id.Down()).IsTransparent(),
		true || r.world.Block(id.Front()).IsTransparent(),
		true || r.world.Block(id.Back()).IsTransparent(),
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

func (r *LineRender) Draw(player *world.Player) {
	width, height := r.win.GetSize()
	projection := mgl32.Perspective(radian(45), float32(width)/float32(height), 0.01, world.ChunkWidth*float32(*RenderRadius))
	camera := player.Matrix()
	mat := projection.Mul4(camera)

	r.shader.Begin()
	r.drawCross()
	r.drawWireFrame(player, mat)
	r.shader.End()
}

func makeCross(shader *glhf.Shader) *Lines {
	return NewLines(shader, []float32{
		-0.5, 0, 0, 0.5, 0, 0,
		0, -0.5, 0, 0, 0.5, 0,
	})
}
