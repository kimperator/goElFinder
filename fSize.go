package goElFinder

import (
	"os"
	"path/filepath"
	"io/ioutil"
)

func (self *elf) size() int64 {
	var size int64
	for _, p := range self.targets {
		s := _size(self.volumes[p.id].Root, p.path)
		size = size + s
	}

	return size
}

func _size(root, path string) int64 {
	var size int64 = 0
	u, err := os.Open(filepath.Join(root, path))
	if err != nil {
		return 0
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return 0
	}
	if info.IsDir() {
		entries, err := ioutil.ReadDir(filepath.Join(root, path))
		if err != nil {
			return 0
		}
		for _, e := range entries {
			if info.IsDir() {
				size = size + _size(root, filepath.Join(path, e.Name()))
			} else {
				size =  size + e.Size()
			}
		}
	} else {
		size = info.Size()
	}
	return size
}

