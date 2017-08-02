package goElFinder

func (self *elf) url() {
	self.res.Url = self.volumes[self.target.id].Url + self.target.path
}
