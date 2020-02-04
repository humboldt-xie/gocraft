package world

const (
	_ DrawType = iota
	DTAir
	DTBlock
	DTPlant
	DTMan
)

type Vec3 struct {
	X, Y, Z int
}

func (v Vec3) Left() Vec3 {
	return Vec3{v.X - 1, v.Y, v.Z}
}
func (v Vec3) Right() Vec3 {
	return Vec3{v.X + 1, v.Y, v.Z}
}
func (v Vec3) Up() Vec3 {
	return Vec3{v.X, v.Y + 1, v.Z}
}
func (v Vec3) Down() Vec3 {
	return Vec3{v.X, v.Y - 1, v.Z}
}
func (v Vec3) Front() Vec3 {
	return Vec3{v.X, v.Y, v.Z + 1}
}
func (v Vec3) Back() Vec3 {
	return Vec3{v.X, v.Y, v.Z - 1}
}

type BlockEngine struct {
}

func (e *BlockEngine) New(t int, pos Vec3) *Block {
	return nil
}

var (
	typeGrassBlock = 1
	typeSandBlock  = 2
	typeGrass      = 17
	typeLeaves     = 15
	typeWood       = 5
	typeCloud      = 16
	TypeAir        = 0
)

type DrawType int

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

var idToType = map[int]*BlockType{}

func RegisterBlockType(id int, ty *BlockType) {
	idToType[id] = ty
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
