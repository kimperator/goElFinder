package goElFinder

import (
	"os"
	"path/filepath"
)

func (self *elf) ls() {
	self.res.List = []string{}
	for _, i := range self.req.Intersect {
		_, err := os.Stat(filepath.Join(self.volumes[self.target.id].Root, self.target.path, i));
		if !os.IsNotExist(err) {
			self.res.List = append(self.res.List, i)
		}
	}
}
