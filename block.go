package main

var (
	typeGrassBlock = 1
	typeSandBlock  = 2
	typeGrass      = 17
	typeLeaves     = 15
	typeWood       = 5
	typeCloud      = 16
	typeAir        = 0
)

type Block struct {
	Type int
	Life int
}

func (b *Block) MakeData(vertices []float32, show [6]bool, id Vec3) []float32 {
	return makeData(b, vertices, show, id)
}

func (b *Block) New() *Block {
	return NewBlock(b.Type)
}

func NewBlock(t int) *Block {
	return &Block{Type: t, Life: 100}
}

func IsPlant(tp *Block) bool {
	if tp == nil {
		return false
	}
	if tp.Type >= 17 && tp.Type <= 31 {
		return true
	}
	return false
}

func (b *Block) IsTransparent() bool {
	if b == nil {
		return true
	}
	if IsPlant(b) || b.Life < 100 {
		return true
	}
	switch b.Type {
	case -1, 0, 10, 15:
		return true
	default:
		return false
	}
}

func IsObstacle(tp *Block) bool {
	if tp == nil {
		return true
	}
	if IsPlant(tp) {
		return false
	}
	switch tp.Type {
	case -1:
		return true
	case 0:
		return false
	default:
		return true
	}
}
