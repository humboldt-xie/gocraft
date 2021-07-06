package render

import (
	"sync"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/humboldt-xie/tinycraft/world"
)

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
	img, rect, err := LoadImage(*texturePath)
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

		cubeData := makeCubeData([]float32{}, world.NewBlock(64), FaceFilter{true, true, true, true, true, true}, Vec3{0, 0, 0})
		r.mesh = NewMesh(r.shader, cubeData, true)
		cubeDataFoot := makeCubeData([]float32{}, world.NewBlock(65), FaceFilter{true, true, true, true, true, true}, Vec3{0, 0, 0})
		r.meshFoot = NewMesh(r.shader, cubeDataFoot, true)

	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

/*func (r *PlayerRender) Add(id int32, p *Player) {
	//r.players[id] = p
	log.Printf("%v\n", game)
	game.players.Store(id, p)
}*/

/*func (r *PlayerRender) UpdateOrAdd(id int32, s world.PlayerState, ismainthread bool) {
	return
	pos := world.Position{
		Vec3: mgl32.Vec3{s.X, s.Y, s.Z},
		Rx:   s.Rx,
		Ry:   s.Ry,
		T:    glfw.GetTime(),
	}
	var p *world.Player

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
}*/

func (r *PlayerRender) DrawPlayer(p *world.Player, mat mgl32.Mat4) {
	mat2 := mat.Mul4(p.ComputeMat())
	r.shader.SetUniformAttr(0, mat2)
	r.mesh.Draw()
	mat1 := mat.Mul4(p.ComputeFootMat())
	r.shader.SetUniformAttr(0, mat1)
	//r.meshFoot.Draw()
}

/*func (r *PlayerRender) Update(dt float64) {
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
}*/

func (r *PlayerRender) Draw(mat mgl32.Mat4, players sync.Map) {
	//mat := game.blockRender.get3dmat()
	r.shader.Begin()
	r.texture.Begin()
	players.Range(func(k, v interface{}) bool {
		p := v.(*world.Player)
		r.DrawPlayer(p, mat)
		return true
	})
	r.texture.End()
	r.shader.End()
}
