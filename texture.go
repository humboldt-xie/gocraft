package main

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
