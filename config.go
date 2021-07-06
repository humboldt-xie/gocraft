package main

import (
	"image"
	"image/draw"
	"io/ioutil"
	"log"

	"github.com/disintegration/imaging"
	"github.com/humboldt-xie/tinycraft/render"
	"github.com/humboldt-xie/tinycraft/world"
	"gopkg.in/yaml.v2"
)

type TextureConfig struct {
	Default string `yaml:"default"`
	Left    string `yaml:"left"`
	Right   string `yaml:"right"`
	Top     string `yaml:"top"`
	Bottom  string `yaml:"bottom"`
	Front   string `yaml:"front"`
	Back    string `yaml:"back"`
}

func (t *TextureConfig) ItemDesc() (*render.TextDesc, error) {
	td := render.TextDesc{}
	var err error
	td.Left, err = addTexture(rgba, t.Left, t.Default)
	if err != nil {
		return nil, err
	}
	td.Right, err = addTexture(rgba, t.Right, t.Default)
	if err != nil {
		return nil, err
	}
	td.Top, err = addTexture(rgba, t.Top, t.Default)
	if err != nil {
		return nil, err
	}
	td.Bottom, err = addTexture(rgba, t.Bottom, t.Default)
	if err != nil {
		return nil, err
	}
	td.Front, err = addTexture(rgba, t.Front, t.Default)
	if err != nil {
		return nil, err
	}
	td.Back, err = addTexture(rgba, t.Back, t.Default)
	if err != nil {
		return nil, err
	}
	return &td, nil
}

type ItemConfig struct {
	Id            int           `yaml:"id"`
	Model         string        `yaml:"model"` //模型
	IsObstacle    bool          `yaml:"is_obstacle"`
	IsTransparent bool          `yaml:"is_transparent"`
	Texture       TextureConfig `yaml:"texture"`
}

type Config struct {
	Items []ItemConfig `yaml:"items"`
}

var rect = image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{2560, 2560}}
var rgba = image.NewRGBA(rect)
var lastId = 0
var ids = map[string]int{}

func addTexture(img *image.RGBA, path string, def string) (int, error) {
	if path == "" {
		path = def
	}
	path = "mods/blocks/" + path
	if id, ok := ids[path]; ok {
		return id, nil
	}
	id := lastId
	ids[path] = id
	lastId++
	rect := img.Bounds()
	bheight := rect.Max.Y - rect.Min.Y
	width := (rect.Max.X - rect.Min.X) / 16
	ext, _, err := render.LoadImage(path)
	if err != nil {
		return id, err
	}
	row := id / 16 % 16
	col := id % 16
	rs := imaging.Resize(ext, width, width, imaging.Lanczos)
	//image.Pt(width*col, rect.Max.Y-(row+1)*width)
	target := rs.Bounds()
	target.Min.X = col * width
	target.Max.X = col*width + width
	target.Min.Y = bheight - row*width - width
	target.Max.Y = bheight - row*width
	draw.Draw(img, target, rs, image.Pt(0, 0), draw.Over)
	return id, nil
}

func InitConfig(file string) error {
	data, err := ioutil.ReadFile("mods/blocks/config.yaml")
	if err != nil {
		return err
	}
	config := Config{}
	yaml.Unmarshal(data, &config)
	log.Printf("%v", config)
	//bwidth := rect.Max.X - rect.Min.X
	for _, item := range config.Items {
		td, err := item.Texture.ItemDesc()
		if err != nil {
			return err
		}
		render.AddTextureDesc(item.Id, *td)
		bt := BlockType{}
		bt.Type = item.Id
		bt.Model = world.GetDrawType(item.Model)
		bt.IsObstacle = item.IsObstacle
		bt.IsTransparent = item.IsTransparent
		world.RegisterBlockType(item.Id, &bt)
		log.Printf("add item %v %v", item, td)
	}
	render.AddTextureDesc(2, render.TextDesc{1, 1, 1, 1, 1, 1})
	imaging.Save(rgba, "texture.png")
	return nil
}
