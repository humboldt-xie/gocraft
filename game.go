package main

import (
	"fmt"
	"time"

	_ "image/png"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/icexin/gocraft-server/proto"
)

type AI interface {
	Think(p *Player)
}

type Physics interface {
	Speed(a mgl32.Vec3)
	Update(p *Player, dt float64)
}

type Game struct {
	win *glfw.Window

	player   *Player
	lx, ly   float64
	prevtime float64

	blockRender  *BlockRender
	lineRender   *LineRender
	playerRender *PlayerRender

	world   *World
	itemidx int
	item    *BlockType
	fps     FPS

	exclusiveMouse bool
	closed         bool
}

func NewGame(w, h int) (*Game, error) {
	var (
		err  error
		game *Game
	)
	game = new(Game)
	game.item = &Blocks[0]

	mainthread.Call(func() {
		win := initGL(w, h)
		win.SetMouseButtonCallback(game.onMouseButtonCallback)
		win.SetCursorPosCallback(game.onCursorPosCallback)
		win.SetFramebufferSizeCallback(game.onFrameBufferSizeCallback)
		win.SetKeyCallback(game.onKeyCallback)
		game.win = win
	})
	game.world = NewWorld()
	game.blockRender, err = NewBlockRender()
	if err != nil {
		return nil, err
	}
	mainthread.Call(func() {
		game.blockRender.UpdateItem(game.item)
	})
	game.lineRender, err = NewLineRender()
	if err != nil {
		return nil, err
	}
	game.playerRender, err = NewPlayerRender()
	if err != nil {
		return nil, err
	}

	game.player = NewPlayer(mgl32.Vec3{0, 16, 0}, nil, &SimplePhysics{})
	game.playerRender.Add(0, game.player)

	go game.blockRender.UpdateLoop()
	go game.syncPlayerLoop()
	return game, nil
}

func (g *Game) setExclusiveMouse(exclusive bool) {
	if exclusive {
		g.win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		g.win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	g.exclusiveMouse = exclusive
}
func (g *Game) UpdateBlock(id Vec3, tp *Block) {
	g.world.UpdateBlock(id, tp)
	g.blockRender.DirtyBlock(id)
	go ClientUpdateBlock(id, tp)
}

func (g *Game) PutBlock(player *Player, item *BlockType) {
	head := player.Head()
	foot := player.Foot()
	_, prev := g.world.HitTest(player.Pos(), player.Front())
	if prev != nil && *prev != head && *prev != foot {
		g.UpdateBlock(*prev, NewBlock(item.Type))
	}

}
func (g *Game) BreakBlock(player *Player) {
	block, _ := g.world.HitTest(player.Pos(), player.Front())
	if block != nil {
		tblock := g.world.Block(*block)
		if tblock != nil {
			tblock.Life -= 40
		}
		if tblock == nil || tblock.Life <= 0 {
			tblock = NewBlock(typeAir)
		}
		g.UpdateBlock(*block, tblock)
	}
}

func (g *Game) onMouseButtonCallback(win *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if !g.exclusiveMouse {
		g.setExclusiveMouse(true)
		return
	}

	if button == glfw.MouseButton2 && action == glfw.Press {
		g.PutBlock(g.player, g.item)
	}
	if button == glfw.MouseButton1 && action == glfw.Press {
		g.BreakBlock(g.player)
	}
}

func (g *Game) onFrameBufferSizeCallback(window *glfw.Window, width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (g *Game) onCursorPosCallback(win *glfw.Window, xpos float64, ypos float64) {
	if !g.exclusiveMouse {
		return
	}
	if g.lx == 0 && g.ly == 0 {
		g.lx, g.ly = xpos, ypos
		return
	}
	dx, dy := xpos-g.lx, g.ly-ypos
	g.lx, g.ly = xpos, ypos
	g.player.ChangeAngle(float32(dx), float32(dy))
}

func (g *Game) onKeyCallback(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}
	switch key {
	case glfw.KeyTab:
		g.player.FlipFlying()
	case glfw.KeySpace:
		g.player.Jump(8)
	case glfw.KeyN:
		for i := 0; i < 1; i++ {
			pos := g.player.Pos()
			front := g.player.Front()
			PlayerID += 1
			g.playerRender.UpdateOrAdd(PlayerID, proto.PlayerState{
				X:  pos.X() + front.X() + float32(i)/5,
				Y:  pos.Y() + front.Y(),
				Z:  pos.Z() + front.Z() + float32(i)/5,
				Rx: g.player.State().Rx,
				Ry: g.player.State().Ry,
			}, true)
		}
	case glfw.KeyE:
		g.itemidx = (1 + g.itemidx) % len(Blocks)
		g.item = &Blocks[g.itemidx]
		g.blockRender.UpdateItem(g.item)
	case glfw.KeyR:
		g.itemidx--
		if g.itemidx < 0 {
			g.itemidx = len(Blocks) - 1
		}
		g.item = &Blocks[g.itemidx]
		g.blockRender.UpdateItem(g.item)
	}
}

func (g *Game) handleKeyInput(dt float64) {
	speed := float32(0.1)
	if g.player.flying {
		speed = 0.1
	}
	if g.win.GetKey(glfw.KeyEscape) == glfw.Press {
		g.setExclusiveMouse(false)
	}
	if g.win.GetKey(glfw.KeyW) == glfw.Press {
		g.player.Move(MoveForward, speed)
	}
	if g.win.GetKey(glfw.KeyS) == glfw.Press {
		g.player.Move(MoveBackward, speed)
	}
	if g.win.GetKey(glfw.KeyA) == glfw.Press {
		g.player.Move(MoveLeft, speed)
	}
	if g.win.GetKey(glfw.KeyD) == glfw.Press {
		g.player.Move(MoveRight, speed)
	}
}

func (g *Game) CurrentBlockid() Vec3 {
	pos := g.player.Pos()
	return NearBlock(pos)
}

func (g *Game) ShouldClose() bool {
	return g.closed
}

func (g *Game) renderStat() {
	g.fps.Update()
	p := g.player.Pos()
	cid := NearBlock(p).Chunkid()
	blockPos, _ := g.world.HitTest(g.player.Pos(), g.player.Front())

	life := 0
	if blockPos != nil {
		block := g.world.Block(*blockPos)
		if block != nil {
			life = block.Life
		}
	}
	stat := g.blockRender.Stat()
	title := fmt.Sprintf("[%.2f %.2f %.2f] %v [%d/%d %d] %d %d/100", p.X(), p.Y(), p.Z(),
		cid, stat.RendingChunks, stat.CacheChunks, stat.Faces, g.fps.Fps(), life)
	g.win.SetTitle(title)
}

func (g *Game) syncPlayerLoop() {
	tick := time.NewTicker(time.Second / 10)
	for range tick.C {
		ClientUpdatePlayerState(g.player.State())
	}
}

func (g *Game) Update() {
	/*pos := g.player.Pos()
	g.playerRender.UpdateOrAdd(1, proto.PlayerState{
		X:  pos.X() + 1.0,
		Y:  pos.Y(),
		Z:  pos.Z() + 1.0,
		Rx: 5,
		Ry: 0,
	})*/
	mainthread.Call(func() {
		var dt float64
		now := glfw.GetTime()
		dt = now - g.prevtime
		g.prevtime = now
		if dt > 0.02 {
			dt = 0.02
		}

		g.handleKeyInput(dt)

		g.playerRender.Update(dt)

		gl.ClearColor(0.57, 0.71, 0.77, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		g.blockRender.Draw()
		g.lineRender.Draw()
		g.playerRender.Draw()
		g.renderStat()

		g.win.SwapBuffers()
		glfw.PollEvents()
		g.closed = g.win.ShouldClose()
	})
}
