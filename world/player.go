package world

import (
	"github.com/go-gl/glfw/v3.3/glfw"
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

type AI interface {
	Think(p *Player)
}

type Physics interface {
	GetSpeed() mgl32.Vec3
	Speed(a mgl32.Vec3)
	Update(p *Player, dt float64)
}

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
	Physics Physics
}

func (c *Player) Update(dt float64) {
	if c.Physics != nil {
		c.Physics.Update(c, dt)
	}
	if c.ai != nil {
		c.ai.Think(c)
	}
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
		Physics: phy,
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
func (p *Player) ComputeFootMat() mgl32.Mat4 {
	body_pos := p.Position
	front := p.WalkFront()
	right := p.Right()
	up := right.Cross(front).Normalize()
	pos := mgl32.Vec3{body_pos.X(), body_pos.Y() - 1, body_pos.Z()}
	return mgl32.LookAtV(pos, pos.Add(front), up).Inv()
}

// 线性插值计算玩家位置
func (p *Player) ComputeMat() mgl32.Mat4 {
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
