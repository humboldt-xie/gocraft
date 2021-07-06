package render

import (
	"fmt"
	"image"
)

func arround(p image.Point) []image.Point {
	return []image.Point{
		{p.X, p.Y - 1},
		{p.X + 1, p.Y - 1},
		{p.X + 1, p.Y},
		{p.X + 1, p.Y + 1},
		{p.X - 1, p.Y - 1},
		{p.X - 1, p.Y},
		{p.X - 1, p.Y + 1},
		{p.X, p.Y + 1},
	}
}

func isEdge(img image.Image, p image.Point) bool {
	ps := arround(p)
	rec := img.Bounds()
	_, _, _, a := img.At(p.X, p.Y).RGBA()
	if a == 0 {
		return false
	}
	for _, v := range ps {
		if v.X < 0 || v.Y < 0 || v.X > rec.Max.X || v.Y > rec.Max.Y {
			fmt.Printf("out of range %v", v)
			continue
		}
		_, _, _, a := img.At(v.X, v.Y).RGBA()
		if a == 0 {
			return true
		}
	}
	return false
}

func deepFind(has map[image.Point]bool, res []image.Point, img image.Image, cur image.Point) ([]image.Point, error) {
	res = append(res, cur)
	has[cur] = true
	ps := arround(cur)
	for _, v := range ps {
		fmt.Printf("check %v->%v has:%v edge:%v %v\n", cur, v, has[v], isEdge(img, v), img.At(v.X, v.Y))
		if isEdge(img, v) && !has[v] {
			return deepFind(has, res, img, v)
		}
	}
	return res, nil

}

func Img2Po(img image.Image) ([]image.Point, error) {
	img, rec, err := LoadImage("test_data/test.png")
	if err != nil {
		return []image.Point{}, err
	}
	has := map[image.Point]bool{}
	res := []image.Point{}
	MinY := rec.Max.Y
	cur := image.Point{}
	for x := 0; x < rec.Max.X; x++ {
		for y := 0; y < rec.Max.Y; y++ {
			tc := image.Point{x, y}
			if isEdge(img, tc) && tc.Y < MinY {
				fmt.Printf("isedge %v \n", tc)
				cur = tc
				MinY = tc.Y
			}
		}
	}
	return deepFind(has, res, img, cur)
}
