package goElFinder

import (
	"path/filepath"
	"os"
	"mime"
	"fmt"
	"strings"
	"io/ioutil"
)


// Request functions
func (self *response) open(init, tree bool, target string) error {
	var (
		err error
	)

	if init {
		target = "/"
		root, err := self._infoPath(self.config.rootDir + "/")
		if err != nil {
			return err
		}
		self.Api = APIver
		self.Cwd = root
		self.Options.Path = "files/1"
		self.Options.Separator = "/"
		self.Options.Url = "http://ly.dmbasis.ru:8080/files/1/"
		self.Options.TmbUrl = "http://ly.dmbasis.ru:8080/files/1/.tmb/"

	}
		obj, err := self._infoPath(filepath.Join(self.config.rootDir, target))
		if err != nil {
			return err
		}
		self.Cwd = obj


	err = filepath.Walk(filepath.Join(self.config.rootDir, target), self._info)
	if err != nil {
		return err
	}

	return nil
}

func (self *response) file(target string) (fileName, mimeType string, data []byte, err error) {
	target = filepath.Join(self.config.rootDir, target)
	data, err = ioutil.ReadFile(target)
	fileName = filepath.Base(target)
	mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	return fileName, mimeType, data, err
}

func (self *response) mkdir(path, name string) error {
	create := filepath.Join(self.config.rootDir, path, name)
	err := os.MkdirAll(create, 0777)
	if err != nil {
		return err
	}
	added, err := self._infoPath(create)
	if err != nil {
		return err
	}
	self.Added = append(self.Added, added)
	if self.Hashes == nil {
		self.Hashes = map[string]string{}
	}
	self.Hashes[name] = createHash(self.config.id, filepath.Join(path, name))
	changed, err := self._infoPath(filepath.Join(self.config.rootDir, path))
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}

func (self *response) rm(path string) error {
	err := os.RemoveAll(filepath.Join(self.config.rootDir, path))
	if err != nil {
		return err
	}
	self.Removed = append(self.Removed, createHash(self.config.id, path))
	return nil
}

func (self *response) rename(path, name string) error {
	newPath := filepath.Join(self.config.rootDir, filepath.Dir(path), filepath.Base(name))
	err := os.Rename(filepath.Join(self.config.rootDir, path), newPath)
	if err != nil {
		return err
	}
	fmt.Println("Rename:", filepath.Join(self.config.rootDir, path), "to", newPath)
	added, err := self._infoPath(newPath)
	self.Added = append(self.Added, added)
	self.Removed = append(self.Removed, createHash(self.config.id, path))
	return nil
}

func (self *response) renames(target, suffix string, renames []string) error {
	fmt.Println("Target:", target, "Suffix:", suffix, "Renames:", renames)
	if len(renames) != 0 {
		for _, r := range renames  {
			oldPath := filepath.Join(self.config.rootDir, target, r)
			newPath := filepath.Join(self.config.rootDir, target, strings.TrimRight(r, filepath.Ext(r)) + suffix + filepath.Ext(r))
			fmt.Println("Renames:", oldPath, "to", newPath)
			err := os.Rename(oldPath, newPath) //ToDo suffix clean
			if err != nil {
				return err
			}
			added, err := self._infoPath(newPath)
			self.Added = append(self.Added, added)
		}
	}
	return nil
}




func (self *response) ls(target string, intersect []string) {
	self.List = []string{}
	for _, i := range intersect {
		_, err := os.Stat(filepath.Join(self.config.rootDir, target, i));
		fmt.Println("File:", filepath.Join(self.config.rootDir, target, i), "is", !os.IsNotExist(err))
		if !os.IsNotExist(err) {
			self.List = append(self.List, i)
		}
	}
}



// ----------------------------------------------------------------------

