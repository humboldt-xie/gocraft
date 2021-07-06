package render

import (
	"log"

	"github.com/faiface/mainthread"
	"github.com/humboldt-xie/tinycraft/world"
)

type ChunkMesh struct {
	world   *world.World
	br      *BlockRender
	id      Vec3
	mesh    *Mesh
	version int64
	sigch   chan bool
}

func NewChunkMesh(world *world.World, br *BlockRender, id Vec3) *ChunkMesh {
	c := world.Chunk(id)
	newMesh := br.makeChunkMesh(c, false)
	nc := &ChunkMesh{world: world, br: br, id: c.Id(), version: c.V(), mesh: newMesh, sigch: make(chan bool)}
	go nc.UpdateLoop(world)
	return nc
}

func (r *ChunkMesh) Close() {
	close(r.sigch)
}

func (r *ChunkMesh) DirtyChunk() {
	r.version -= 1
	r.checkChunk()
}

func (r *ChunkMesh) UpdateLoop(world *world.World) {
	defer func() {
		mainthread.CallNonBlock(func() {
			r.mesh.Release()
		})
	}()
	for {
		_, ok := <-r.sigch
		if !ok {
			return
		}
		r.updateMesh()
	}
}
func (r *ChunkMesh) checkChunk() {
	// nonblock signal
	log.Printf("check chunk %v", r.id)
	select {
	case r.sigch <- true:
	default:
	}
}

func (r *ChunkMesh) updateMesh() {
	c := r.world.Chunk(r.id)
	if r.version == c.V() {
		return
	}
	newMesh := r.br.makeChunkMesh(c, false)
	oldMesh := r.mesh
	r.mesh = newMesh
	r.version = c.V()
	mainthread.CallNonBlock(func() {
		oldMesh.Release()
	})
}
