package goElFinder

import (
	"fmt"
	"crypto/sha1"
	"path/filepath"
)

func (self *response) tmb(targets []string) error {
	self.Images= map[string]string{}
	for _, t := range targets {
		//tmb := encode64(fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(t)))))+filepath.Ext(t)
		tmb := encode64(t)
		stmb := fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(t))))+filepath.Ext(t)
		err := resizeImage(filepath.Join(self.config.rootDir, t), filepath.Join(self.config.rootDir, ".tmb", stmb), 48, 0)
		if err != nil {
			return err
		}
		self.Images[tmb] = stmb
	}
	return nil
}

func (self *response) dim(path string) error {
	var err error
	target := filepath.Join(self.config.rootDir, path)
	self.Dim, err = getImageDim(target)
	if err != nil {
		return err
	}
	return nil
}

func (self *response) resize(path string, width, height int) error {
	img := filepath.Join(self.config.rootDir, path)
	err := resizeImage(img, img, width, height)
	if err != nil {
		return err
	}
	changed, err := self._infoPath(img)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}


func (self *response) crop(path string, x, y, width, height int) error {
	img := filepath.Join(self.config.rootDir, path)
	err := cropImage(img, img, x, y, width, height)
	if err != nil {
		return err
	}
	changed, err := self._infoPath(img)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}

func (self *response) rotate(path, bg string, degree int) error {
	img := filepath.Join(self.config.rootDir, path)
	err := rotateImage(img, img, bg, degree)
	if err != nil {
		return err
	}
	changed, err := self._infoPath(img)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}