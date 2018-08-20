package main

const (
	grassBlock = 1
	sandBlock  = 2
	grass      = 17
	leaves     = 15
	wood       = 5
	air        = 0
)

type Block int

func IsPlant(tp Block) bool {
	if tp >= 17 && tp <= 31 {
		return true
	}
	return false
}

func IsTransparent(tp Block) bool {
	if IsPlant(tp) {
		return true
	}
	switch tp {
	case -1, 0, 10, 15:
		return true
	default:
		return false
	}
}

func IsObstacle(tp Block) bool {
	if IsPlant(tp) {
		return false
	}
	switch tp {
	case -1:
		return true
	case 0:
		return false
	default:
		return true
	}
}
