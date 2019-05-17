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

type DrawType int

const (
	_ DrawType = iota
	DTAir
	DTBlock
	DTPlant
	DTMan
)

type BlockType struct {
	Type          int
	DrawType      DrawType
	IsTransparent bool
	IsObstacle    bool
}

func (t *BlockType) Data(w *Block, vertices []float32, show [6]bool, block Vec3) []float32 {

	return nil
}

type Block struct {
	Type int
	Life int
}

func (b *Block) New() *Block {
	return NewBlock(b.Type)
}

func NewBlock(t int) *Block {
	block := &Block{Type: t, Life: 100}
	return block
}

func (b *Block) BlockType() *BlockType {
	return idToType[b.Type]
}

// 是否透明 返回true 则不绘制
func (b *Block) IsTransparent() bool {
	if b == nil {
		return true
	}
	if b.Life < 100 {
		return true
	}
	return b.BlockType().IsTransparent
}

// 是否可穿越
func (tp *Block) IsObstacle() bool {
	if tp == nil {
		return false
	}
	return tp.BlockType().IsObstacle
}

var idToType = map[int]*BlockType{}

func init() {
	for i, _ := range Blocks {
		idToType[Blocks[i].Type] = &Blocks[i]
	}
}

var Blocks = []BlockType{
	BlockType{Type: 0, IsObstacle: false, IsTransparent: true, DrawType: DTAir},
	BlockType{Type: 1, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 2, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 3, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 4, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 5, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 6, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 7, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 8, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 9, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 10, IsObstacle: false, IsTransparent: true, DrawType: DTAir},
	BlockType{Type: 11, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 12, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 13, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 14, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 15, IsObstacle: false, IsTransparent: true, DrawType: DTAir},
	BlockType{Type: 16, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 17, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 18, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 19, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 20, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 21, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 22, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 23, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 24, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 25, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 26, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 27, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 28, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 29, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 30, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 31, IsObstacle: false, IsTransparent: true, DrawType: DTPlant},
	BlockType{Type: 32, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 33, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 34, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 35, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 36, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 37, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 38, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 39, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 40, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 41, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 42, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 43, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 44, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 45, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 46, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 47, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 48, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 49, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 50, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 51, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 52, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 53, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 54, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 55, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 56, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 57, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 58, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 59, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 60, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 61, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 62, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 63, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 64, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
	BlockType{Type: 65, IsObstacle: true, IsTransparent: false, DrawType: DTBlock},
}
