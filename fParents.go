package goElFinder

import "path/filepath"

func (self *elf) parents(t target) error {
	self.res.Tree = []fileDir{}

	for t.path != string(filepath.Separator) {
		t.path = filepath.Join(t.path, "..")
		fd := self.volumes.listDirs(t)
		for _, f := range fd {
			if i, err := self.volumes.infoTarget(target{id: t.id, path: f}); err == nil {
				self.res.Tree = append(self.res.Tree, i)
			}
		}
	}
	for k := range self.volumes {
		p, err := self.volumes.infoTarget(target{id: k, path: ""})
		if err == nil {
			p.Options.Url = self.volumes[k].Url
			self.res.Tree = append(self.res.Tree, p)
		}
	}
	return nil
}
