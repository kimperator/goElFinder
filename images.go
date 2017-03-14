package goElFinder

import (
	"path/filepath"
	"os"
)

func (self *response) tmb(id string, path map[string]string) error {
	self.Images= map[string]string{}
	for i, t := range path {
		tmb := encode64(t)
		stmb := tmb + filepath.Ext(t)
		os.MkdirAll(filepath.Join(conf[i].Root, filepath.Dir(t), ".tmb"), 0755)
		err := resizeImage(filepath.Join(conf[i].Root, t), filepath.Join(conf[i].Root, filepath.Dir(t), ".tmb", stmb), 48, 0)
		if err != nil {
			return err
		}
		self.Images[tmb] = stmb
	}
	return nil
}

func (self *response) dim(id, path string) error {
	var err error
	target := filepath.Join(conf[id].Root, path)
	self.Dim, err = getImageDim(target)
	if err != nil {
		return err
	}
	return nil
}

func (self *response) resize(id, path string, width, height int) error {
	img := filepath.Join(conf[id].Root, path)
	err := resizeImage(img, img, width, height)
	if err != nil {
		return err
	}
	changed, err := _infoFileDir(id,path)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}


func (self *response) crop(id, path string, x, y, width, height int) error {
	img := filepath.Join(conf[id].Root, path)
	err := cropImage(img, img, x, y, width, height)
	if err != nil {
		return err
	}
	changed, err := _infoFileDir(id, path)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}

func (self *response) rotate(id, path, bg string, degree int) error {
	img := filepath.Join(conf[id].Root, path)
	err := rotateImage(img, img, bg, degree)
	if err != nil {
		return err
	}
	changed, err := _infoFileDir(id, path)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}