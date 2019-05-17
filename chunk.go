package main

import (
	"log"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	ChunkWidth = 32
)

func (v Vec3) Chunkid() Vec3 {
	return Vec3{
		int(math.Floor(float64(v.X) / ChunkWidth)),
		0,
		int(math.Floor(float64(v.Z) / ChunkWidth)),
	}
}

func NearBlock(pos mgl32.Vec3) Vec3 {
	return Vec3{
		int(round(pos.X())),
		int(round(pos.Y())),
		int(round(pos.Z())),
	}
}

type Chunk struct {
	id      Vec3
	version int64
	blocks  sync.Map // map[Vec3]int
}

func NewChunk(id Vec3) *Chunk {
	//log.Printf("new chunk %v", id)
	c := &Chunk{
		id:      id,
		version: 0,
	}
	return c
}

func (c *Chunk) Id() Vec3 {
	return c.id
}
func (c *Chunk) V() int64 {
	return c.version
}

func (c *Chunk) Block(id Vec3) *Block {
	if id.Chunkid() != c.id {
		log.Panicf("id %v chunk %v", id, c.id)
	}
	w, ok := c.blocks.Load(id)
	if ok {
		return w.(*Block)
	}
	if id.Y >= 12 {
		return NewBlock(typeAir)
	}
	//c.add(id, 1)
	return nil
}

func (c *Chunk) add(id Vec3, w *Block) {
	if id.Chunkid() != c.id {
		log.Panicf("id %v chunk %v", id, c.id)
	}
	c.version += 1
	c.blocks.Store(id, w)
}

func (c *Chunk) del(id Vec3) {
	if id.Chunkid() != c.id {
		log.Panicf("id %v chunk %v", id, c.id)
	}
	c.version += 1
	c.blocks.Delete(id)
}

func (c *Chunk) RangeBlocks(f func(id Vec3, w *Block)) {
	c.blocks.Range(func(key, value interface{}) bool {
		f(key.(Vec3), value.(*Block))
		return true
	})
}
