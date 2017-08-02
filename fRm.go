package goElFinder

import (
	"os"
	"path/filepath"
)

func (self *elf) rm() error {
	for i := range self.req.Targets {
		err := os.RemoveAll(filepath.Join(self.volumes[self.targets[i].id].Root, self.targets[i].path))
		if err != nil {
			return err
		}
		self.res.Removed = append(self.res.Removed, createHash(self.targets[i].id, self.targets[i].path))
	}

	return nil
}
