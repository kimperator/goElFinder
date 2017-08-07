package goElFinder

import (
	"path/filepath"
	"os"
)

func (self *elf) mkdir() error {
	create := filepath.Join(self.target.path, self.req.Name)
	err := os.MkdirAll(filepath.Join(self.volumes[self.target.id].Root, create), 0755)
	if err != nil {
		return err
	}
	added, err := self.volumes.infoTarget(target{id: self.target.id, path: create})
	if err != nil {
		return err
	}
	self.res.Added = append(self.res.Added, added)
	if self.res.Hashes == nil {
		self.res.Hashes = map[string]string{}
	}
	self.res.Hashes[self.res.Name] = createHash(self.target.id, create)
	return nil
}

func (self *elf) mkdirs() error {
	for _, d := range self.req.Dirs {
		err := os.MkdirAll(filepath.Join(self.volumes[self.target.id].Root, self.target.path, d), 0755)
		if err != nil {
			return err
		}
		added, err := self.volumes.infoTarget(target{id: self.target.id, path: filepath.Join(self.target.path, d)})
		if err != nil {
			return err
		}
		self.res.Added = append(self.res.Added, added)
		if self.res.Hashes == nil {
			self.res.Hashes = map[string]string{}
		}
		self.res.Hashes[self.res.Name] = createHash(self.target.id, d)

	}

	/*err := []string{}
	for _, f := range e.dirs {
		e.req.Name = f
		er := e.mkdir(id, e.path.path)
		if er != nil {
			err = append(err, er.Error())
		}
	}
	if len(err) > 0 {
		e.res.Error = err
	}*/

	return nil
}
