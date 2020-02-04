package render

import (
	"github.com/humboldt-xie/tinycraft/world"
)

type Block = world.Block
type Vec3 = world.Vec3

const (
	sleft = iota
	sright
	sup
	sdown
	sfront
	sback
)

// show: left, right, up, down, front, back,
func makeCubeData(vertices []float32, w *Block, show [6]bool, block Vec3) []float32 {
	texture := tex.Texture(w)
	l, r := texture.Left, texture.Right
	u, d := texture.Up, texture.Down
	f, b := texture.Front, texture.Back
	x, y, z := float32(block.X), float32(block.Y), float32(block.Z)
	cubeHeight := float32(0.5) //float32(0.5 * (float32(w.Life) - 50) / 50)
	//cubeWeight := float32(0.5 * (float32(w.Life) / 100)) //1.0 / 2
	cubeWeight := float32(0.5) //1.0 / 2

	if show[sleft] {
		vertices = append(vertices, []float32{
			// left
			x - cubeWeight, y - 0.5, z - cubeWeight, l[0].X(), l[0].Y(), -1, 0, 0,
			x - cubeWeight, y - 0.5, z + cubeWeight, l[1][0], l[1][1], -1, 0, 0,
			x - cubeWeight, y + cubeHeight, z + cubeWeight, l[2][0], l[2][1], -1, 0, 0,
			x - cubeWeight, y + cubeHeight, z + cubeWeight, l[3][0], l[3][1], -1, 0, 0,
			x - cubeWeight, y + cubeHeight, z - cubeWeight, l[4][0], l[4][1], -1, 0, 0,
			x - cubeWeight, y - 0.5, z - cubeWeight, l[5][0], l[5][1], -1, 0, 0,
		}...)
	}
	if show[sright] {
		vertices = append(vertices, []float32{
			// right
			x + cubeWeight, y - 0.5, z + cubeWeight, r[0][0], r[0][1], 1, 0, 0,
			x + cubeWeight, y - 0.5, z - cubeWeight, r[1][0], r[1][1], 1, 0, 0,
			x + cubeWeight, y + cubeHeight, z - cubeWeight, r[2][0], r[2][1], 1, 0, 0,
			x + cubeWeight, y + cubeHeight, z - cubeWeight, r[3][0], r[3][1], 1, 0, 0,
			x + cubeWeight, y + cubeHeight, z + cubeWeight, r[4][0], r[4][1], 1, 0, 0,
			x + cubeWeight, y - 0.5, z + cubeWeight, r[5][0], r[5][1], 1, 0, 0,
		}...)
	}
	if show[sup] {
		vertices = append(vertices, []float32{
			// top
			x - cubeWeight, y + cubeHeight, z + cubeWeight, u[0][0], u[0][1], 0, 1, 0,
			x + cubeWeight, y + cubeHeight, z + cubeWeight, u[1][0], u[1][1], 0, 1, 0,
			x + cubeWeight, y + cubeHeight, z - cubeWeight, u[2][0], u[2][1], 0, 1, 0,
			x + cubeWeight, y + cubeHeight, z - cubeWeight, u[3][0], u[3][1], 0, 1, 0,
			x - cubeWeight, y + cubeHeight, z - cubeWeight, u[4][0], u[4][1], 0, 1, 0,
			x - cubeWeight, y + cubeHeight, z + cubeWeight, u[5][0], u[5][1], 0, 1, 0,
		}...)
	}

	if show[sdown] {
		vertices = append(vertices, []float32{
			// bottom
			x - cubeWeight, y - 0.5, z - cubeWeight, d[0][0], d[0][1], 0, -1, 0,
			x + cubeWeight, y - 0.5, z - cubeWeight, d[1][0], d[1][1], 0, -1, 0,
			x + cubeWeight, y - 0.5, z + cubeWeight, d[2][0], d[2][1], 0, -1, 0,
			x + cubeWeight, y - 0.5, z + cubeWeight, d[3][0], d[3][1], 0, -1, 0,
			x - cubeWeight, y - 0.5, z + cubeWeight, d[4][0], d[4][1], 0, -1, 0,
			x - cubeWeight, y - 0.5, z - cubeWeight, d[5][0], d[5][1], 0, -1, 0,
		}...)
	}

	if show[sfront] {
		vertices = append(vertices, []float32{
			// front
			x - cubeWeight, y - 0.5, z + cubeWeight, f[0][0], f[0][1], 0, 0, 1,
			x + cubeWeight, y - 0.5, z + cubeWeight, f[1][0], f[1][1], 0, 0, 1,
			x + cubeWeight, y + cubeHeight, z + cubeWeight, f[2][0], f[2][1], 0, 0, 1,
			x + cubeWeight, y + cubeHeight, z + cubeWeight, f[3][0], f[3][1], 0, 0, 1,
			x - cubeWeight, y + cubeHeight, z + cubeWeight, f[4][0], f[4][1], 0, 0, 1,
			x - cubeWeight, y - 0.5, z + cubeWeight, f[5][0], f[5][1], 0, 0, 1,
		}...)
	}

	if show[sback] {
		vertices = append(vertices, []float32{
			// back
			x + cubeWeight, y - 0.5, z - cubeWeight, b[0][0], b[0][1], 0, 0, -1,
			x - cubeWeight, y - 0.5, z - cubeWeight, b[1][0], b[1][1], 0, 0, -1,
			x - cubeWeight, y + cubeHeight, z - cubeWeight, b[2][0], b[2][1], 0, 0, -1,
			x - cubeWeight, y + cubeHeight, z - cubeWeight, b[3][0], b[3][1], 0, 0, -1,
			x + cubeWeight, y + cubeHeight, z - cubeWeight, b[4][0], b[4][1], 0, 0, -1,
			x + cubeWeight, y - 0.5, z - cubeWeight, b[5][0], b[5][1], 0, 0, -1,
		}...)
	}

	return vertices
}

func makeWireFrameData(vertices []float32, show [6]bool) []float32 {
	if show[sleft] {
		vertices = append(vertices, []float32{
			// left
			-0.5, -0.5, -0.5,
			-0.5, -0.5, +0.5,

			-0.5, -0.5, +0.5,
			-0.5, +0.5, +0.5,

			-0.5, +0.5, +0.5,
			-0.5, +0.5, -0.5,

			-0.5, +0.5, -0.5,
			-0.5, -0.5, -0.5,
		}...)
	}
	if show[sright] {
		vertices = append(vertices, []float32{
			// right
			+0.5, -0.5, +0.5,
			+0.5, -0.5, -0.5,

			+0.5, -0.5, -0.5,
			+0.5, +0.5, -0.5,

			+0.5, +0.5, -0.5,
			+0.5, +0.5, +0.5,

			+0.5, +0.5, +0.5,
			+0.5, -0.5, +0.5,
		}...)
	}

	if show[sup] {
		vertices = append(vertices, []float32{
			// top
			-0.5, +0.5, +0.5,
			+0.5, +0.5, +0.5,

			+0.5, +0.5, +0.5,
			+0.5, +0.5, -0.5,

			+0.5, +0.5, -0.5,
			-0.5, +0.5, -0.5,

			-0.5, +0.5, -0.5,
			-0.5, +0.5, +0.5,
		}...)
	}

	if show[sdown] {
		vertices = append(vertices, []float32{
			// bottom
			+0.5, -0.5, +0.5,
			-0.5, -0.5, +0.5,

			-0.5, -0.5, +0.5,
			-0.5, -0.5, -0.5,

			-0.5, -0.5, -0.5,
			+0.5, -0.5, -0.5,

			+0.5, -0.5, -0.5,
			+0.5, -0.5, +0.5,
		}...)
	}

	if show[sfront] {
		// z front
		vertices = append(vertices, []float32{
			-0.5, -0.5, +0.5,
			+0.5, -0.5, +0.5,

			+0.5, -0.5, +0.5,
			+0.5, +0.5, +0.5,

			+0.5, +0.5, +0.5,
			-0.5, +0.5, +0.5,

			-0.5, +0.5, +0.5,
			-0.5, -0.5, +0.5,
		}...)
	}

	if show[sback] {
		vertices = append(vertices, []float32{
			// back
			+0.5, -0.5, -0.5,
			-0.5, -0.5, -0.5,

			-0.5, -0.5, -0.5,
			-0.5, +0.5, -0.5,

			-0.5, +0.5, -0.5,
			+0.5, +0.5, -0.5,

			+0.5, +0.5, -0.5,
			+0.5, -0.5, -0.5,
		}...)
	}

	return vertices
}

func makePlantData(vertices []float32, w *Block, show [6]bool, block Vec3) []float32 {
	texture := tex.Texture(w)
	l, r := texture.Left, texture.Right
	f, b := texture.Front, texture.Back
	x, y, z := float32(block.X), float32(block.Y), float32(block.Z)
	cubeHeight := float32(0.5)
	cubeWeight := float32(0.5)
	vertices = append(vertices, []float32{
		// left
		// x y z tex-x tex-y
		x, y - 0.5, z - cubeWeight, l[0][0], l[0][1], -1, 0, 0,
		x, y - 0.5, z + cubeWeight, l[1][0], l[1][1], -1, 0, 0,
		x, y + cubeHeight, z + cubeWeight, l[2][0], l[2][1], -1, 0, 0,
		x, y + cubeHeight, z + cubeWeight, l[3][0], l[3][1], -1, 0, 0,
		x, y + cubeHeight, z - cubeWeight, l[4][0], l[4][1], -1, 0, 0,
		x, y - 0.5, z - cubeWeight, l[5][0], l[5][1], -1, 0, 0,
	}...)
	vertices = append(vertices, []float32{
		// right
		x, y - 0.5, z + cubeWeight, r[0][0], r[0][1], 1, 0, 0,
		x, y - 0.5, z - cubeWeight, r[1][0], r[1][1], 1, 0, 0,
		x, y + cubeHeight, z - cubeWeight, r[2][0], r[2][1], 1, 0, 0,
		x, y + cubeHeight, z - cubeWeight, r[3][0], r[3][1], 1, 0, 0,
		x, y + cubeHeight, z + cubeWeight, r[4][0], r[4][1], 1, 0, 0,
		x, y - 0.5, z + cubeWeight, r[5][0], r[5][1], 1, 0, 0,
	}...)

	vertices = append(vertices, []float32{
		// front
		x - cubeWeight, y - 0.5, z, f[0][0], f[0][1], 0, 0, 1,
		x + cubeWeight, y - 0.5, z, f[1][0], f[1][1], 0, 0, 1,
		x + cubeWeight, y + cubeHeight, z, f[2][0], f[2][1], 0, 0, 1,
		x + cubeWeight, y + cubeHeight, z, f[3][0], f[3][1], 0, 0, 1,
		x - cubeWeight, y + cubeHeight, z, f[4][0], f[4][1], 0, 0, 1,
		x - cubeWeight, y - 0.5, z, f[5][0], f[5][1], 0, 0, 1,
	}...)

	vertices = append(vertices, []float32{
		// back
		x + cubeWeight, y - 0.5, z, b[0][0], b[0][1], 0, 0, -1,
		x - cubeWeight, y - 0.5, z, b[1][0], b[1][1], 0, 0, -1,
		x - cubeWeight, y + cubeHeight, z, b[2][0], b[2][1], 0, 0, -1,
		x - cubeWeight, y + cubeHeight, z, b[3][0], b[3][1], 0, 0, -1,
		x + cubeWeight, y + cubeHeight, z, b[4][0], b[4][1], 0, 0, -1,
		x + cubeWeight, y - 0.5, z, b[5][0], b[5][1], 0, 0, -1,
	}...)
	return vertices
}

func makeData(w *Block, vertices []float32, show [6]bool, block Vec3) []float32 {
	switch w.BlockType().DrawType {
	case world.DTAir:
		return vertices
	case world.DTPlant:
		return makePlantData(vertices, w, show, block)
	default:
		return makeCubeData(vertices, w, show, block)
	}
}
