package goElFinder

import (
	"os"
	"path/filepath"
)

func (self *elf) mkfile() error {

	err := os.MkdirAll(filepath.Join(self.volumes[self.target.id].Root, self.target.path), 0755)
	if err != nil {
		return err
	}
	create := filepath.Join(self.target.path, self.req.Name)
	f, err := os.OpenFile(filepath.Join(self.volumes[self.target.id].Root, create), os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	f.Close()

	added, err := self.volumes.infoFileDir(target{id: self.target.id, path: create})
	if err != nil {
		return err
	}

	self.res.Added = append(self.res.Added, added)

	return nil
}
