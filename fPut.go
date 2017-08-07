package goElFinder

import (
	"path/filepath"
	"io/ioutil"
)

func (self *elf) put() error {
	err := ioutil.WriteFile(filepath.Join(self.volumes[self.target.id].Root, self.target.path), []byte(self.req.Content), 0666)
	if err != nil {
		return err
	}
	info, err := self.volumes.infoTarget(self.target)
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, info)
	return nil
}
