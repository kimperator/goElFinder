package goElFinder

import (
	"path/filepath"
	"os"
)

func (self *elf) tmb() error {
	self.res.Images= map[string]string{}
	for _, p := range self.targets {
		tmb := encode64(p.path)
		stmb := tmb + filepath.Ext(p.path)
		os.MkdirAll(filepath.Join(conf[p.id].Root, filepath.Dir(p.path), ".tmb"), 0755)
		err := resizeImage(filepath.Join(conf[p.id].Root, p.path), filepath.Join(conf[p.id].Root, filepath.Dir(p.path), ".tmb", stmb), 48, 0)
		if err != nil {
			return err
		}
		self.res.Images[tmb] = stmb
	}
	return nil
}

func (self *elf) dim() error {
	var err error
	target := filepath.Join(conf[self.target.id].Root, self.target.path)
	self.res.Dim, err = getImageDim(target)
	if err != nil {
		return err
	}
	return nil
}

func (self *elf) resize(id, path string) error {
	img := filepath.Join(conf[id].Root, path)
	err := resizeImage(img, img, self.req.Width, self.req.Height)
	if err != nil {
		return err
	}
	changed, err := _infoFileDir(target{id: id, path: path})
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, changed)
	return nil
}


func (self *elf) crop(id, path string) error {
	img := filepath.Join(conf[id].Root, path)
	err := cropImage(img, img, self.req.X, self.req.Y, self.req.Width, self.req.Height)
	if err != nil {
		return err
	}
	changed, err := _infoFileDir(target{id: id, path: path})
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, changed)
	return nil
}

func (self *elf) rotate(id, path string) error {
	img := filepath.Join(conf[id].Root, path)
	err := rotateImage(img, img, self.req.Bg, self.req.Degree)
	if err != nil {
		return err
	}
	changed, err := _infoFileDir(target{id: id, path: path})
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, changed)
	return nil
}