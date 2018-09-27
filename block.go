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
	Type      int
	DrawType  DrawType
	TextureID int
}

type Block struct {
	ID   int
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
	for i := 0; i < len(Blocks); i++ {
		if Blocks[i].Type == b.Type {
			return &Blocks[i]
		}
	}
	return nil
}

func IsPlant(tp *Block) bool {
	if tp == nil {
		return false
	}
	return tp.BlockType().DrawType == DTPlant
}

func (b *Block) IsTransparent() bool {
	if b == nil {
		return true
	}
	if IsPlant(b) || b.Life < 100 {
		return true
	}
	return b.BlockType().DrawType == DTAir
}

func IsObstacle(tp *Block) bool {
	if tp == nil {
		return true
	}
	switch tp.BlockType().DrawType {
	case DTPlant, DTAir:
		return false
	case DTBlock:
		return true
	default:
		return true
	}
}

var Blocks = []BlockType{
	BlockType{0, DTAir, 0},
	BlockType{1, DTBlock, 1},
	BlockType{2, DTBlock, 2},
	BlockType{3, DTBlock, 3},
	BlockType{4, DTBlock, 4},
	BlockType{5, DTBlock, 5},
	BlockType{6, DTBlock, 6},
	BlockType{7, DTBlock, 7},
	BlockType{8, DTBlock, 8},
	BlockType{9, DTBlock, 9},
	BlockType{10, DTAir, 10},
	BlockType{11, DTBlock, 11},
	BlockType{12, DTBlock, 12},
	BlockType{13, DTBlock, 13},
	BlockType{14, DTBlock, 14},
	BlockType{15, DTAir, 15},
	BlockType{16, DTBlock, 16},
	BlockType{17, DTPlant, 17},
	BlockType{18, DTPlant, 18},
	BlockType{19, DTPlant, 19},
	BlockType{20, DTPlant, 20},
	BlockType{21, DTPlant, 21},
	BlockType{22, DTPlant, 22},
	BlockType{23, DTPlant, 23},
	BlockType{24, DTPlant, 24},
	BlockType{25, DTPlant, 25},
	BlockType{26, DTPlant, 26},
	BlockType{27, DTPlant, 27},
	BlockType{28, DTPlant, 28},
	BlockType{29, DTPlant, 29},
	BlockType{30, DTPlant, 30},
	BlockType{31, DTPlant, 31},
	BlockType{32, DTBlock, 32},
	BlockType{33, DTBlock, 33},
	BlockType{34, DTBlock, 34},
	BlockType{35, DTBlock, 35},
	BlockType{36, DTBlock, 36},
	BlockType{37, DTBlock, 37},
	BlockType{38, DTBlock, 38},
	BlockType{39, DTBlock, 39},
	BlockType{40, DTBlock, 40},
	BlockType{41, DTBlock, 41},
	BlockType{42, DTBlock, 42},
	BlockType{43, DTBlock, 43},
	BlockType{44, DTBlock, 44},
	BlockType{45, DTBlock, 45},
	BlockType{46, DTBlock, 46},
	BlockType{47, DTBlock, 47},
	BlockType{48, DTBlock, 48},
	BlockType{49, DTBlock, 49},
	BlockType{50, DTBlock, 50},
	BlockType{51, DTBlock, 51},
	BlockType{52, DTBlock, 52},
	BlockType{53, DTBlock, 53},
	BlockType{54, DTBlock, 54},
	BlockType{55, DTBlock, 55},
	BlockType{56, DTBlock, 56},
	BlockType{57, DTBlock, 57},
	BlockType{58, DTBlock, 58},
	BlockType{59, DTBlock, 59},
	BlockType{60, DTBlock, 60},
	BlockType{61, DTBlock, 61},
	BlockType{62, DTBlock, 62},
	BlockType{63, DTBlock, 63},
	BlockType{64, DTBlock, 64},
}
