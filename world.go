package main

import (
	"log"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	lru "github.com/hashicorp/golang-lru"
)

type World struct {
	mutex  sync.Mutex
	chunks *lru.Cache // map[Vec3]*Chunk
}

func NewWorld() *World {
	m := (*renderRadius) * (*renderRadius) * 4
	chunks, _ := lru.New(m)
	return &World{
		chunks: chunks,
	}
}

func (w *World) Collide(from, to mgl32.Vec3) (mgl32.Vec3, bool) {
	x, y, z := to.X(), to.Y(), to.Z()
	nx, ny, nz := round(to.X()), round(to.Y()), round(to.Z())
	const pad = 0.25

	head := Vec3{int(nx), int(ny), int(nz)}
	foot := Vec3{int(nx), int(ny), int(nz)}.Down()

	stop := false
	for _, b := range []Vec3{foot, head} {
		if w.Block(b.Left()).IsObstacle() && x < nx && nx-x > pad {
			x = nx - pad
		}
		if w.Block(b.Right()).IsObstacle() && x > nx && x-nx > pad {
			x = nx + pad
		}
		if w.Block(b.Down()).IsObstacle() && y < ny && ny-y > pad {
			y = ny - pad
			stop = true
		}
		if w.Block(b.Up()).IsObstacle() && y > ny && y-ny > pad {
			y = ny + pad
			stop = true
		}
		if w.Block(b.Back()).IsObstacle() && z < nz && nz-z > pad {
			z = nz - pad
		}
		if w.Block(b.Front()).IsObstacle() && z > nz && z-nz > pad {
			z = nz + pad
		}
	}
	return mgl32.Vec3{x, y, z}, stop
}

//HitTest pos 当前位置, vec 方向
func (w *World) HitTest(pos mgl32.Vec3, vec mgl32.Vec3) (*Vec3, *Vec3) {
	var (
		maxLen = float32(8.0)
		step   = float32(0.125)

		block, prev Vec3
		pprev       *Vec3
	)

	for length := float32(0); length < maxLen; length += step {
		block = NearBlock(pos.Add(vec.Mul(length)))
		if prev != block && w.HasBlock(block) {
			return &block, pprev
		}
		prev = block
		pprev = &prev
	}
	return nil, nil
}

func (w *World) Block(id Vec3) *Block {
	chunk := w.BlockChunk(id)
	if chunk == nil {
		return nil
	}
	block := chunk.Block(id)
	return block
}

func (w *World) BlockChunk(block Vec3) *Chunk {
	cid := block.Chunkid()
	chunk, ok := w.loadChunk(cid)
	if !ok {
		return nil
	}
	return chunk
}
func (w *World) CreateBlock(id Vec3, tp *Block) {
	otp := w.Block(id)
	if otp != nil {
		return
	}
	log.Printf("create %v", id)
	w.updateBlock(id, tp)
}
func (w *World) Generate(id Vec3) {
	log.Printf("generate %v", id)
	nw := typeSandBlock
	if noise2(-float32(id.X)*0.1, float32(id.Y)*0.1, 4, 0.8, 2) > 0.6 {
		nw = typeGrassBlock
		width := 10
		//length := 10
		height := 5
		y := id.Y - height
		minY := id.Y - height
		maxY := id.Y
		minX := id.X - width/2
		maxX := id.X + width/2
		minZ := id.Z - width/2
		maxZ := id.Z + width/2
		for ; y <= maxY; y++ {
			for x := minX; x <= id.X+width/2; x++ {
				for z := id.Z - width/2; z <= id.Z+width/2; z++ {
					if y == minY || y == maxY || x == minX || x == maxX || z == minZ || z == maxZ {
						nw = typeGrassBlock
					} else {
						nw = typeAir
					}
					w.CreateBlock(Vec3{x, y, z}, NewBlock(nw))
				}
			}
		}
	}
	for x := id.X - 1; x <= id.X+1; x++ {
		for y := id.Y - 1; y <= id.Y+1; y++ {
			for z := id.Z - 1; z <= id.Z+1; z++ {
				w.CreateBlock(Vec3{x, y, z}, NewBlock(nw))
			}
		}
	}
}
func (w *World) updateBlock(id Vec3, tp *Block) {
	chunk := w.BlockChunk(id)
	if chunk != nil {
		chunk.add(id, tp)
	}
	store.UpdateBlock(id, tp)

}
func (w *World) UpdateBlock(id Vec3, tp *Block) {
	w.updateBlock(id, tp)
	if id.Y <= 12 {
		w.Generate(id)
	}
}

func (w *World) HasBlock(id Vec3) bool {
	tp := w.Block(id)
	return tp != nil && tp.BlockType().DrawType != DTAir
}

func (w *World) Chunk(id Vec3) *Chunk {
	p, ok := w.loadChunk(id)
	if ok {
		return p
	}
	chunk := NewChunk(id)
	blocks := makeChunkMap(id)
	for block, tp := range blocks {
		chunk.add(block, tp)
	}
	err := store.RangeBlocks(id, func(bid Vec3, w *Block) {
		chunk.add(bid, w)
	})
	if err != nil {
		log.Printf("fetch chunk(%v) from db error:%s", id, err)
		return nil
	}
	ClientFetchChunk(id, func(bid Vec3, w *Block) {
		chunk.add(bid, w)
		store.UpdateBlock(bid, w)
	})
	w.storeChunk(id, chunk)
	return chunk
}

func (w *World) Chunks(ids []Vec3) []*Chunk {
	ch := make(chan *Chunk)
	var chunks []*Chunk
	for _, id := range ids {
		id := id
		go func() {
			ch <- w.Chunk(id)
		}()
	}
	for range ids {
		chunk := <-ch
		if chunk != nil {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

func makeChunkMap(cid Vec3) map[Vec3]*Block {
	m := make(map[Vec3]*Block)
	p, q := cid.X, cid.Z
	for dx := 0; dx < ChunkWidth; dx++ {
		for dz := 0; dz < ChunkWidth; dz++ {
			x, z := p*ChunkWidth+dx, q*ChunkWidth+dz
			f := noise2(float32(x)*0.01, float32(z)*0.01, 4, 0.5, 2)
			g := noise2(float32(-x)*0.01, float32(-z)*0.01, 2, 0.9, 2)
			mh := int(g*32 + 16)
			h := int(f * float32(mh))
			tb := typeGrassBlock
			if h <= 12 {
				h = 12
				tb = typeSandBlock
			}
			// grass and sand
			for y := 11; y < h; y++ {
				m[Vec3{x, y, z}] = NewBlock(tb)
			}

			// flowers
			if tb == typeGrassBlock {
				if noise2(-float32(x)*0.1, float32(z)*0.1, 4, 0.8, 2) > 0.6 {
					m[Vec3{x, h, z}] = NewBlock(typeGrass)
				}
				if noise2(float32(x)*0.05, float32(-z)*0.05, 4, 0.8, 2) > 0.7 {
					tb := 18 + int(noise2(float32(x)*0.1, float32(z)*0.1, 4, 0.8, 2)*7)
					m[Vec3{x, h, z}] = NewBlock(tb)
				}
			}

			// tree
			if tb == typeGrassBlock {
				ok := true
				if dx-4 < 0 || dz-4 < 0 ||
					dx+4 > ChunkWidth || dz+4 > ChunkWidth {
					ok = false
				}
				if ok && noise2(float32(x), float32(z), 6, 0.5, 2) > 0.79 {
					for y := h + 3; y < h+8; y++ {
						for ox := -3; ox <= 3; ox++ {
							for oz := -3; oz <= 3; oz++ {
								d := ox*ox + oz*oz + (y-h-4)*(y-h-4)
								if d < 11 {
									m[Vec3{x + ox, y, z + oz}] = NewBlock(typeLeaves)
								}
							}
						}
					}
					for y := h; y < h+7; y++ {
						m[Vec3{x, y, z}] = NewBlock(typeWood)
					}
				}
			}

			// cloud
			for y := 64; y < 72; y++ {
				if noise3(float32(x)*0.01, float32(y)*0.1, float32(z)*0.01, 8, 0.5, 2) > 0.69 {
					m[Vec3{x, y, z}] = NewBlock(typeCloud)
				}
			}
		}
	}
	return m
}

func (w *World) loadChunk(id Vec3) (*Chunk, bool) {
	chunk, ok := w.chunks.Get(id)
	if !ok {
		return nil, false
	}
	return chunk.(*Chunk), true
}

func (w *World) storeChunk(id Vec3, chunk *Chunk) {
	w.chunks.Add(id, chunk)
}
