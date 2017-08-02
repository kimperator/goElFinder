package goElFinder

import (
	"path/filepath"
	"os"
	"fmt"
	"github.com/disintegration/imaging"
	"path"
	"github.com/disintegration/gift"
	"strconv"
	"image"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
	"image/color"
)

func (self *elf) tmb() error {
	self.res.Images= map[string]string{}
	for _, p := range self.targets {
		tmb := encode64(p.path)
		stmb := tmb + filepath.Ext(p.path)
		os.MkdirAll(filepath.Join(self.volumes[p.id].Root, filepath.Dir(p.path), ".tmb"), 0755)
		err := resizeImage(filepath.Join(self.volumes[p.id].Root, p.path), filepath.Join(self.volumes[p.id].Root, filepath.Dir(p.path), ".tmb", stmb), 48, 0)
		if err != nil {
			return err
		}
		self.res.Images[tmb] = stmb
	}
	return nil
}

func (self *elf) dim() error {
	var err error
	target := filepath.Join(self.volumes[self.target.id].Root, self.target.path)
	self.res.Dim, err = getImageDim(target)
	if err != nil {
		return err
	}
	return nil
}

func (self *elf) resize(id, path string) error {
	img := filepath.Join(self.volumes[id].Root, path)
	err := resizeImage(img, img, self.req.Width, self.req.Height)
	if err != nil {
		return err
	}
	changed, err := self.volumes.infoFileDir(target{id: id, path: path})
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, changed)
	return nil
}


func (self *elf) crop(id, path string) error {
	img := filepath.Join(self.volumes[id].Root, path)
	err := cropImage(img, img, self.req.X, self.req.Y, self.req.Width, self.req.Height)
	if err != nil {
		return err
	}
	changed, err := self.volumes.infoFileDir(target{id: id, path: path})
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, changed)
	return nil
}

func (self *elf) rotate(id, path string) error {
	img := filepath.Join(self.volumes[id].Root, path)
	err := rotateImage(img, img, self.req.Bg, self.req.Degree)
	if err != nil {
		return err
	}
	changed, err := self.volumes.infoFileDir(target{id: id, path: path})
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, changed)
	return nil
}

// Image dimension return string("(width)x(height)")
func getImageDim(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%dx%d", img.Width, img.Height), nil
}

// Resize image
func resizeImage(imagePath, imageTarget string, width, height int) error {
	src, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}
	dst := imaging.Resize(src, width, height, imaging.Box)
	os.MkdirAll(path.Dir(imageTarget), os.ModePerm)
	return imaging.Save(dst, imageTarget)
}

// Crop image
func cropImage(imagePath, imageTarget string, x, y, width, height int) error {
	src, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}
	rectangle := image.Rect(x, y, x+width, y+height)
	dst := imaging.Crop(src, rectangle)
	os.MkdirAll(path.Dir(imageTarget), os.ModePerm)
	return imaging.Save(dst, imageTarget)
}

// Rotate image
func rotateImage(imagePath, imageTarget, bg string, degree int) error {
	degree = 360 - degree
	src, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}

	switch degree {
	case 90:
		dst := imaging.Rotate90(src)
		err = imaging.Save(dst, imageTarget)

	case 180:
		dst := imaging.Rotate180(src)
		err = imaging.Save(dst, imageTarget)

	case 270:
		dst := imaging.Rotate270(src)
		err = imaging.Save(dst, imageTarget)
	default:
		r, g, b, a := hexColor(bg)
		gf := gift.New(gift.Rotate(float32(degree), color.NRGBA{R:r,G:g,B:b,A:a}, gift.LinearInterpolation))
		dst := image.NewRGBA(gf.Bounds(src.Bounds()))
		gf.Draw(dst, src)
		err = imaging.Save(dst, imageTarget)
	}

	os.MkdirAll(path.Dir(imageTarget), os.ModePerm)
	return err
}

// Parse hex color code to RGBA
func hexColor(s string) (r, g, b, a uint8) {
	var c string
	if len(s) < 3 {
		c = "0xffffffff"
	} else if s[0] == 35 {
		c = "0x" + s[1:]
	} else if s[0:1] != "0x" {
		c = "0x" + s
	}
	for len(c) <= 9 {
		c = c + "f"
	}
	decimal, _ := strconv.ParseUint(c, 0, 32)
	a = uint8(decimal & 0xFF)
	b = uint8((decimal >> 8) & 0xFF)
	g = uint8((decimal >> 16) & 0xFF)
	r = uint8((decimal >> 24) & 0xFF)
	return
}
