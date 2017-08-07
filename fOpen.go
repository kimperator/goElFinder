package goElFinder

func (self *elf) open() error {
	var err error

	if self.req.Init {
		self.res.Api = APIver
	}

	//obj, err := self._infoPath(filepath.Join(self.current.rootDir, target))
	obj, err := self.volumes.infoTarget(self.target)
	if err != nil {
		return err
	}
	self.res.Cwd = obj
	self.res.Files = []fileDir{}
	if self.req.Tree {
		for k := range self.volumes {
			p, err := self.volumes.infoTarget(target{id: k})
			if err == nil {
				p.Options.Url = self.volumes[k].Url
				self.res.Files = append(self.res.Files, p)
			}
		}
	}

	fd := self.volumes.listAll(self.target)
	for _, f := range fd {
		i, err := self.volumes.infoTarget(target{id: self.target.id, path: f})
		if err == nil {
			self.res.Files = append(self.res.Files, i)
		}
	}
	return nil
}
