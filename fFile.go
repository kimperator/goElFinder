package goElFinder

import (
	"path/filepath"
	"mime"
	"io/ioutil"
)

func (self *elf) file() (fileName, mimeType string, data []byte, err error) {
	path := filepath.Join(self.volumes[self.target.id].Root, self.target.path)
	data, err = ioutil.ReadFile(path)
	fileName = filepath.Base(path)
	mimeType = mime.TypeByExtension(filepath.Ext(fileName))

	return fileName, mimeType, data, err
}
