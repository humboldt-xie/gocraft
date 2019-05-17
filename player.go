package main

import (
	"log"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Movement int

var PlayerID = int32(1)

const (
	MoveForward Movement = iota
	MoveBackward
	MoveLeft
	MoveRight
)

type Position struct {
	mgl32.Vec3
	Rx, Ry float32
	T      float64
}

func (p *Position) Front() mgl32.Vec3 {
	front := mgl32.Vec3{
		cos(radian(p.Ry)) * cos(radian(p.Rx)),
		sin(radian(p.Ry)),
		cos(radian(p.Ry)) * sin(radian(p.Rx)),
	}
	return front.Normalize()
}

type Player struct {
	Position
	ID      int
	Sens    float32
	pre     Position
	flying  bool
	ai      AI
	physics Physics
}

func (c *Player) Head() Vec3 {
	return NearBlock(c.Pos())
}
func (c *Player) Foot() Vec3 {
	return c.Head().Down()
}

func (c *Player) State() Position {
	return c.Position
}

func (c *Player) Jump(delta float32) {
	block := game.CurrentBlockid()
	if game.world.HasBlock(Vec3{block.X, block.Y - 2, block.Z}) {
		c.physics.Speed(mgl32.Vec3{0, delta, 0})
	}
}

func (c *Player) Move(dir Movement, delta float32) {
	if c.flying {
		delta = 5 * delta
	}
	c.pre = c.Position
	c.Position.T = glfw.GetTime()
	switch dir {
	case MoveForward:
		if c.flying {
			c.Position.Vec3 = c.Position.Add(c.Front().Mul(delta))
		} else {
			c.Position.Vec3 = c.Position.Add(c.WalkFront().Mul(delta))
		}
	case MoveBackward:
		if c.flying {
			c.Position.Vec3 = c.Position.Sub(c.Front().Mul(delta))
		} else {
			c.Position.Vec3 = c.Position.Sub(c.WalkFront().Mul(delta))
		}
	case MoveLeft:
		c.Position.Vec3 = c.Position.Sub(c.Right().Mul(delta))
	case MoveRight:
		c.Position.Vec3 = c.Position.Add(c.Right().Mul(delta))
	}
	c.Position.T = glfw.GetTime()
	c.UpdateState(c.Position)
}

func (c *Player) ChangeAngle(dx, dy float32) {
	if mgl32.Abs(dx) > 200 || mgl32.Abs(dy) > 200 {
		return
	}
	c.pre = c.Position
	c.Position.T = glfw.GetTime()
	c.Position.Rx += dx * c.Sens
	c.Position.Ry += dy * c.Sens
	if c.Position.Ry > 89 {
		c.Position.Ry = 89
	}
	if c.Position.Ry < -89 {
		c.Position.Ry = -89
	}
}

func NewPlayer(pos mgl32.Vec3, ai AI, phy Physics) *Player {
	p := &Player{
		//front:  mgl32.Vec3{0, 0, -1},
		ai:      ai,
		physics: phy,
		Sens:    0.14,
		flying:  false,
	}
	//r.players[id] = p
	p.Position = Position{Vec3: pos, T: glfw.GetTime(), Rx: -90, Ry: 0}
	p.pre = p.Position
	return p
}

func (c *Player) Matrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position.Vec3, c.Position.Add(c.Front()), c.Up())
}

// 线性插值计算玩家位置
func (p *Player) computeFootMat() mgl32.Mat4 {
	body_pos := p.Position
	front := p.WalkFront()
	right := p.Right()
	up := right.Cross(front).Normalize()
	pos := mgl32.Vec3{body_pos.X(), body_pos.Y() - 1, body_pos.Z()}
	return mgl32.LookAtV(pos, pos.Add(front), up).Inv()
}

// 线性插值计算玩家位置
func (p *Player) computeMat() mgl32.Mat4 {
	t1 := p.Position.T - p.pre.T
	t2 := glfw.GetTime() - p.Position.T
	t := min(float32(t2/t1), 1)

	x := mix(p.Position.X(), p.pre.X(), t)
	y := mix(p.Position.Y(), p.pre.Y(), t)
	z := mix(p.Position.Z(), p.pre.Z(), t)
	rx := mix(p.Position.Rx, p.pre.Rx, t)
	ry := mix(p.Position.Ry, p.pre.Ry, t)

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
	p.pre, p.Position = p.Position, s
}

func (c *Player) Restore(state Position) {
	c.Position = state //= mgl32.Vec3{state.X, state.Y, state.Z,RX:state.RX}
}

func (c *Player) SetPos(pos mgl32.Vec3) {
	c.pre = c.Position
	c.Position.Vec3 = pos
	c.Position.T = glfw.GetTime()
}

func (c *Player) Pos() mgl32.Vec3 {
	return c.Position.Vec3
}
func (c *Player) Up() mgl32.Vec3 {
	return c.Right().Cross(c.Front()).Normalize()
}
func (c *Player) WalkFront() mgl32.Vec3 {
	return mgl32.Vec3{0, 1, 0}.Cross(c.Right()).Normalize()
}
func (c *Player) Right() mgl32.Vec3 {
	front := c.Front()
	return front.Cross(mgl32.Vec3{0, 1, 0}).Normalize()
}

func (c *Player) Front() mgl32.Vec3 {
	front := mgl32.Vec3{
		cos(radian(c.Position.Ry)) * cos(radian(c.Position.Rx)),
		sin(radian(c.Position.Ry)),
		cos(radian(c.Position.Ry)) * sin(radian(c.Position.Rx)),
	}
	return front.Normalize()
}

func (c *Player) FlipFlying() {
	c.flying = !c.flying
}

func (c *Player) Flying() bool {
	return c.flying
}

type PlayerRender struct {
	shader  *glhf.Shader
	texture *glhf.Texture
	//players map[int32]*Player
	mesh     *Mesh
	meshFoot *Mesh
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
		//players: make(map[int32]*Player),
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
		r.texture = glhf.NewTexture(rect.Dx(), rect.Dy(), false, img.Pix)

		cubeData := makeCubeData([]float32{}, NewBlock(64), [...]bool{true, true, true, true, true, true}, Vec3{0, 0, 0})
		r.mesh = NewMesh(r.shader, cubeData, true)
		cubeDataFoot := makeCubeData([]float32{}, NewBlock(65), [...]bool{true, true, true, true, true, true}, Vec3{0, 0, 0})
		r.meshFoot = NewMesh(r.shader, cubeDataFoot, true)

	})
	if err != nil {
		return nil, err
	}

	return r, nil
}
func (r *PlayerRender) Add(id int32, p *Player) {
	//r.players[id] = p
	log.Printf("%v\n", game)
	game.players.Store(id, p)
}

func (r *PlayerRender) UpdateOrAdd(id int32, s PlayerState, ismainthread bool) {
	pos := Position{
		Vec3: mgl32.Vec3{s.X, s.Y, s.Z},
		Rx:   s.Rx,
		Ry:   s.Ry,
		T:    glfw.GetTime(),
	}
	var p *Player

	mp, ok := game.players.Load(id)
	if !ok {
		log.Printf("add new player %d", id)
		f := pos.Front()
		p = NewPlayer(pos.Vec3, &CircleAI{}, &SimplePhysics{vx: f.X() * 50, vz: f.Z() * 50, vy: f.Y() * 50})
		r.Add(id, p)
	} else {
		p = mp.(*Player)
	}
	p.UpdateState(pos)
}

func (r *PlayerRender) Remove(id int32) {
	log.Printf("remove player %d", id)
	/*p, ok := r.players[id]
	if ok {
		mainthread.CallNonBlock(func() {
			p.Release()
		})
	}*/
	game.players.Delete(id)

}

type CircleAI struct {
	g *Game
}

func (c *CircleAI) Think(p *Player) {
	//_, prev := game.world.HitTest(p.Pos(), p.Front())
	p.ChangeAngle(1, 0)
	//if prev == nil {
	p.Move(MoveForward, 0.1)
	game.BreakBlock(p)
	//}
}

type SimplePhysics struct {
	vx, vz, vy float32
}

func (sp *SimplePhysics) Speed(a mgl32.Vec3) {
	sp.vx = a.X()
	sp.vy = a.Y()
	sp.vz = a.Z()
}

func (sp *SimplePhysics) Update(p *Player, dt float64) {
	from := p.Pos()
	pos := p.Pos()
	stop := false
	if !p.Flying() {
		sp.vy -= float32(dt * 20)
		if sp.vy < -50 {
			sp.vy = -50
		}
		pos = mgl32.Vec3{
			from.X() + sp.vx*float32(dt),
			from.Y() + sp.vy*float32(dt),
			from.Z() + sp.vz*float32(dt),
		}
	}

	pos, stop = game.world.Collide(from, pos)
	if stop {
		sp.vx = 0
		sp.vz = 0
		if sp.vy > -5 {
			sp.vy = 0
		} else if sp.vy < -5 {
			sp.vy = -sp.vy * 0.1
		}
	}
	p.SetPos(pos)
}

func (r *PlayerRender) DrawPlayer(p *Player, mat mgl32.Mat4) {
	mat2 := mat.Mul4(p.computeMat())
	r.shader.SetUniformAttr(0, mat2)
	r.mesh.Draw()
	mat1 := mat.Mul4(p.computeFootMat())
	r.shader.SetUniformAttr(0, mat1)
	r.meshFoot.Draw()
}
func (r *PlayerRender) Update(dt float64) {
	game.players.Range(func(k, v interface{}) bool {
		p := v.(*Player)
		if p.ai != nil {
			p.ai.Think(p)
		}
		if p.physics != nil {
			p.physics.Update(p, dt)
		}
		return true
	})
}

func (r *PlayerRender) Draw() {
	mat := game.blockRender.get3dmat()
	r.shader.Begin()
	r.texture.Begin()
	game.players.Range(func(k, v interface{}) bool {
		p := v.(*Player)
		r.DrawPlayer(p, mat)
		return true
	})
	r.texture.End()
	r.shader.End()
}
