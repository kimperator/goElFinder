package goElFinder

import (
	"path/filepath"
	"io/ioutil"
)

func (self *elf) get() error {
	b, err := ioutil.ReadFile(filepath.Join(self.volumes[self.target.id].Root, self.target.path))
	if err != nil {
		return err
	}
	self.res.Content = string(b)
	return nil
}
