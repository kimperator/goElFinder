package goElFinder

func (self *elf) tree(t target) error {
	fd := self.volumes.listDirs(t)
	self.res.Tree = []fileDir{}
	for _, f := range fd {
		if i, err := self.volumes.infoTarget(target{id:t.id, path:f}); err == nil {
			self.res.Tree = append(self.res.Tree, i)
		}
	}
	return nil
}
