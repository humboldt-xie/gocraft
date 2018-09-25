package main

import (
	"log"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/icexin/gocraft-server/proto"
)

type PlayerMovement int

const (
	MoveForward PlayerMovement = iota
	MoveBackward
	MoveLeft
	MoveRight
)

type Position struct {
	mgl32.Vec3
	Rx, Ry float32
	T      float64
}

/*type PositionState struct {
	s1, s2 Position
	t1, t2 float64
}*/

/*type playerState struct {
	PlayerState
	time float64
}*/

type Player struct {
	pre    Position
	pos    Position
	up     mgl32.Vec3
	right  mgl32.Vec3
	front  mgl32.Vec3
	wfront mgl32.Vec3
	flying bool
	Sens   float32

	shader *glhf.Shader
	mesh   *Mesh
}

func (c *Player) State() Position {
	return c.pos
}

func (c *Player) Move(dir PlayerMovement, delta float32) {
	if c.flying {
		delta = 5 * delta
	}
	switch dir {
	case MoveForward:
		if c.flying {
			c.pos.Vec3 = c.pos.Add(c.front.Mul(delta))
		} else {
			c.pos.Vec3 = c.pos.Add(c.wfront.Mul(delta))
		}
	case MoveBackward:
		if c.flying {
			c.pos.Vec3 = c.pos.Sub(c.front.Mul(delta))
		} else {
			c.pos.Vec3 = c.pos.Sub(c.wfront.Mul(delta))
		}
	case MoveLeft:
		c.pos.Vec3 = c.pos.Sub(c.right.Mul(delta))
	case MoveRight:
		c.pos.Vec3 = c.pos.Add(c.right.Mul(delta))
	}
	c.pos.T = glfw.GetTime()
}

func (c *Player) ChangeAngle(dx, dy float32) {
	if mgl32.Abs(dx) > 200 || mgl32.Abs(dy) > 200 {
		return
	}
	c.pos.Rx += dx * c.Sens
	c.pos.Ry += dy * c.Sens
	if c.pos.Ry > 89 {
		c.pos.Ry = 89
	}
	if c.pos.Ry < -89 {
		c.pos.Ry = -89
	}
	c.updateAngles()
}

func NewPlayer(pos mgl32.Vec3) *Player {
	//b := NewBlock(64)
	//log.Printf("add new player %d", id)
	//cubeData := makeCubeData([]float32{}, b, [...]bool{true, true, true, true, true, true}, Vec3{0, 0, 0})
	//var mesh *Mesh
	//mesh = NewMesh(r.shader, cubeData, false)
	p := &Player{
		//shader: r.shader,
		//mesh:   mesh,
		front:  mgl32.Vec3{0, 0, -1},
		Sens:   0.14,
		flying: false,
	}
	//r.players[id] = p
	p.pos = Position{Vec3: pos, T: glfw.GetTime(), Rx: -90, Ry: 0}
	p.pre = p.pos
	p.updateAngles()
	return p
}

func (c *Player) Matrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.pos.Vec3, c.pos.Add(c.front), c.up)
}
func (c *Player) updateAngles() {
	front := mgl32.Vec3{
		cos(radian(c.pos.Ry)) * cos(radian(c.pos.Rx)),
		sin(radian(c.pos.Ry)),
		cos(radian(c.pos.Ry)) * sin(radian(c.pos.Rx)),
	}
	c.front = front.Normalize()
	c.right = c.front.Cross(mgl32.Vec3{0, 1, 0}).Normalize()
	c.up = c.right.Cross(c.front).Normalize()
	c.wfront = mgl32.Vec3{0, 1, 0}.Cross(c.right).Normalize()
}

// 线性插值计算玩家位置
func (p *Player) computeMat() mgl32.Mat4 {
	t1 := p.pos.T - p.pre.T
	t2 := glfw.GetTime() - p.pos.T
	t := min(float32(t2/t1), 1)

	x := mix(p.pos.X(), p.pre.X(), t)
	y := mix(p.pos.Y(), p.pre.Y(), t)
	z := mix(p.pos.Z(), p.pre.Z(), t)
	rx := mix(p.pos.Rx, p.pre.Rx, t)
	ry := mix(p.pos.Ry, p.pre.Ry, t)

	front := mgl32.Vec3{
		cos(radian(ry)) * cos(radian(rx)),
		sin(radian(ry)),
		cos(radian(ry)) * sin(radian(rx)),
	}.Normalize()
	right := front.Cross(mgl32.Vec3{0, 1, 0})
	up := right.Cross(front).Normalize()
	pos := mgl32.Vec3{x, y, z}
	return mgl32.LookAtV(pos, pos.Add(front), up).Inv()
}

func (p *Player) UpdateState(s Position) {
	p.pre, p.pos = p.pos, s
}

func (p *Player) Draw(mat mgl32.Mat4) {
	mat = mat.Mul4(p.computeMat())

	p.shader.SetUniformAttr(0, mat)
	p.mesh.Draw()
}
func (c *Player) Restore(state Position) {
	c.pos = state //= mgl32.Vec3{state.X, state.Y, state.Z,RX:state.RX}
	c.updateAngles()
}

func (c *Player) SetPos(pos mgl32.Vec3) {
	c.pos.Vec3 = pos
	c.updateAngles()
}

func (c *Player) Pos() mgl32.Vec3 {
	return c.pos.Vec3
}

func (c *Player) Front() mgl32.Vec3 {
	return c.front
}

func (c *Player) FlipFlying() {
	c.flying = !c.flying
}

func (c *Player) Flying() bool {
	return c.flying
}

func (p *Player) Release() {
	p.mesh.Release()
}

type PlayerRender struct {
	shader  *glhf.Shader
	texture *glhf.Texture
	players map[int32]*Player
}

func NewPlayerRender() (*PlayerRender, error) {
	var (
		err error
	)
	img, rect, err := loadImage(*texturePath)
	if err != nil {
		return nil, err
	}

	r := &PlayerRender{
		players: make(map[int32]*Player),
	}
	mainthread.Call(func() {
		r.shader, err = glhf.NewShader(glhf.AttrFormat{
			glhf.Attr{Name: "pos", Type: glhf.Vec3},
			glhf.Attr{Name: "tex", Type: glhf.Vec2},
			glhf.Attr{Name: "normal", Type: glhf.Vec3},
		}, glhf.AttrFormat{
			glhf.Attr{Name: "matrix", Type: glhf.Mat4},
		}, playerVertexSource, playerFragmentSource)

		if err != nil {
			return
		}
		r.texture = glhf.NewTexture(rect.Dx(), rect.Dy(), false, img)

	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *PlayerRender) UpdateOrAdd(id int32, s proto.PlayerState) {
	pos := Position{
		Vec3: mgl32.Vec3{s.X, s.Y, s.Z},
		Rx:   s.Rx,
		Ry:   s.Ry,
		T:    glfw.GetTime(),
	}

	p, ok := r.players[id]
	if !ok {
		b := NewBlock(64)
		log.Printf("add new player %d", id)
		cubeData := makeCubeData([]float32{}, b, [...]bool{true, true, true, true, true, true}, Vec3{0, 0, 0})
		var mesh *Mesh
		mesh = NewMesh(r.shader, cubeData, false)
		p = &Player{
			shader: r.shader,
			mesh:   mesh,
		}
		r.players[id] = p
		p.pos = pos
	}
	p.UpdateState(pos)
}

func (r *PlayerRender) Remove(id int32) {
	log.Printf("remove player %d", id)
	p, ok := r.players[id]
	if ok {
		mainthread.CallNonBlock(func() {
			p.Release()
		})
	}
	delete(r.players, id)

}

func (r *PlayerRender) Draw() {
	mat := game.blockRender.get3dmat()
	r.shader.Begin()
	r.texture.Begin()
	for _, p := range r.players {
		p.Draw(mat)
	}
	r.texture.End()
	r.shader.End()
}
