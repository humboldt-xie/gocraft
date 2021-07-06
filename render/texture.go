package render

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func LoadImage(fname string) (*image.RGBA, image.Rectangle, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, image.Rectangle{}, err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)

	bounds := rgba.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r := rgba.At(x, y).(color.RGBA)
			if r.A == 0 {
				r.R = 255
				r.G = 0
				r.B = 255
			}
			rgba.Set(x, y, r)
		}
	}

	return rgba, img.Bounds(), nil
}

/*
type Texture struct {
	image  *image.RGBA
	images map[string]*image.RGBA
}

func (t *Texture) Init() {
	//default 16*256 image
	rect := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{256 * 16, 256 * 256}}
	t.image = image.NewRGBA(rect)
	t.images = make(map[string]*image.RGBA)
}

func (t *Texture) Add(filename string) {
	img, rect, err := loadImage(filename)
	if err != nil {
		continue
	}
	for y := 0; y < rect.Max.Y; y += 256 {
		for x := 0; x < rect.Max.X; x += 256 {

		}
	}
}*/
