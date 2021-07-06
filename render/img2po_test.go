package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/disintegration/imaging"
)

func TestImg2po(t *testing.T) {
	img, _, _ := LoadImage("test_data/test.png")
	if isEdge(img, image.Point{0, 0}) {
		t.Fatalf("no edge")
	}
	p := image.Point{1, 12}
	if !isEdge(img, p) {
		_, _, _, a := img.At(p.X, p.Y).RGBA()
		t.Logf("%v %v %v", img.At(p.X, p.Y), p, a)
		ps := arround(p)
		for _, v := range ps {
			_, _, _, a := img.At(v.X, v.Y).RGBA()
			t.Logf("%v %v %v", img.At(v.X, v.Y), v, a)
		}
		t.Fatalf("no edge")
	}

	vecs, _ := Img2Po(nil)
	for _, v := range vecs {
		t.Logf("%v", v)
	}
	for i, v := range vecs {
		img.Set(v.X, v.Y, color.RGBA{uint8(255), uint8(i), 0, 255})
	}
	imaging.Save(img, "test_data/result.png")
	t.Fatalf("%#v", vecs)

}
