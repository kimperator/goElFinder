package goElFinder

import (
	"path/filepath"
	"os"
	"strings"
)

func (self *elf) rename(id, path string) error {
	newPath := filepath.Join(filepath.Dir(path), filepath.Base(self.req.Name))
	err := os.Rename(filepath.Join(self.volumes[id].Root, path), filepath.Join(self.volumes[id].Root, newPath))
	if err != nil {
		return err
	}
	added, err := self.volumes.infoFileDir(target{id: id, path: newPath})
	self.res.Added = append(self.res.Added, added)
	self.res.Removed = append(self.res.Removed, createHash(id, path))
	return nil
}

func (self *elf) renames(id, path string) error {
	if len(self.req.Renames) != 0 {
		for _, r := range self.req.Renames  {
			oldPath := filepath.Join(path, r)
			newPath := filepath.Join(path, strings.TrimRight(r, filepath.Ext(r)) + self.req.Suffix + filepath.Ext(r))

			err := os.Rename(filepath.Join(self.volumes[id].Root, oldPath), filepath.Join(self.volumes[id].Root, newPath)) //ToDo suffix clean
			if err != nil {
				return err
			}
			added, err := self.volumes.infoFileDir(target{id: id, path: newPath})
			self.res.Added = append(self.res.Added, added)
		}
	}
	return nil
}
