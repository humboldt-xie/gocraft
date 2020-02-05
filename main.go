package main

import (
	"flag"
	"log"
	"time"

	_ "image/png"

	"net/http"
	_ "net/http/pprof"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/humboldt-xie/tinycraft/render"
	"github.com/humboldt-xie/tinycraft/world"
	//"github.com/icexin/gocraft-server/proto"
)

var (
	pprofPort = flag.String("pprof", "", "http pprof port")

	game *Game
)

func initGL(w, h int) *glfw.Window {
	err := glfw.Init()
	if err != nil {
		log.Fatal(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, gl.TRUE)

	win, err := glfw.CreateWindow(w, h, "gocraft", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	win.MakeContextCurrent()
	err = gl.Init()
	if err != nil {
		log.Fatal(err)
	}
	glfw.SwapInterval(1) // enable vsync
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	//gl.BlendFunc(gl.SRC_ALPHA, gl.ZERO)
	//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	//gl.Enable(gl.BLEND)
	return win
}

type FPS struct {
	lastUpdate time.Time
	cnt        int
	fps        int
}

func (f *FPS) Update() {
	f.cnt++
	now := time.Now()
	p := now.Sub(f.lastUpdate)
	if p >= time.Second {
		f.fps = int(float64(f.cnt) / p.Seconds())
		f.cnt = 0
		f.lastUpdate = now
	}
}

func (f *FPS) Fps() int {
	return f.fps
}

func run() {
	err := render.LoadTextureDesc()
	if err != nil {
		log.Fatal(err)
	}

	err = world.InitBoltStore()
	if err != nil {
		log.Panic(err)
	}
	defer world.CloseStore()

	/*if *listenAddr != "" {
		err := InitService()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = InitClient()
		if err != nil {
			log.Panic(err)
		}
		if client != nil {
			defer client.Close()
		}
	}*/

	game, err = NewGame(800, 600)
	if err != nil {
		log.Panic(err)
	}

	/*game.player = store.GetPlayer()
	if client == nil {
		game.playerRender.Add(0, game.player)
	} else {
		game.playerRender.Add(client.ClientID, game.player)
	}*/

	//tick := time.Tick(time.Second / 60)
	md := time.Second / 120
	d := md
	timer := time.NewTimer(d)
	for !game.ShouldClose() {
		<-timer.C
		start := time.Now()
		game.Update()
		d = md - time.Since(start)
		if d < 0 {
			d = 1
		}
		timer.Reset(d)
		log.Printf("update spend %fs %fs", float64(time.Since(start))/float64(time.Second), float64(d+time.Since(start))/float64(time.Second))
	}
	//store.UpdatePlayer(game.player)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()
	go func() {
		if *pprofPort != "" {
			log.Fatal(http.ListenAndServe(*pprofPort, nil))
		}
	}()
	mainthread.Run(run)
}
