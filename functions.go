package goElFinder

import (
	"path/filepath"
	"os"
	"mime"
	"strings"
	"io/ioutil"
	"errors"
)


// Request functions
func (self *response) open(id, path string, init, tree bool) error {
	var err error

	if init {
		self.Api = APIver
	}

	//obj, err := self._infoPath(filepath.Join(self.current.rootDir, target))
	obj, err := _infoFileDir(id, path)
	if err != nil {
		return err
	}
	self.Cwd = obj
	self.Files = []fileDir{}
	if tree {
		for k := range conf {
				p, err := _infoFileDir(k, "")
				if err == nil {
					p.Options.Url = conf[k].Url
					self.Files = append(self.Files, p)
				}
		}
	}

	fd := _listAll(id, path)
	for _, f := range fd {
		i, err := _infoFileDir(id, f)
		if err == nil {
			self.Files = append(self.Files, i)
		}
	}
	return nil
}

func (self *response) tree(id, path string) error {
	fd := _listDirs(id, path)
	self.Tree = []fileDir{}
	for _, f := range fd {
		if i, err := _infoFileDir(id, f); err == nil {
			self.Tree = append(self.Tree, i)
		}
	}
	return nil
}

func (self *response) parents(id, path string) error {
	self.Tree = []fileDir{}

	for path != string(filepath.Separator) {
		path = filepath.Join(path, "..")
		fd := _listDirs(id, path)
		for _, f := range fd {
			if i, err := _infoFileDir(id, f); err == nil {
				self.Tree = append(self.Tree, i)
			}
		}
	}
	for k := range conf {
		p, err := _infoFileDir(k, "")
		if err == nil {
			p.Options.Url = conf[k].Url
			self.Tree = append(self.Tree, p)
		}
	}
	return nil
}

func (self *response) file(id, path string) (fileName, mimeType string, data []byte, err error) {
	if _getRight(id, path) {
		path = filepath.Join(conf[id].Root, path)
		data, err = ioutil.ReadFile(path)
		fileName = filepath.Base(path)
		mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	} else {
		err = errors.New("Permission denied")
	}

	return fileName, mimeType, data, err
}

func (self *response) mkdir(id, path, name string) error {
	create := filepath.Join(path, name)
	err := os.MkdirAll(filepath.Join(conf[id].Root, create), 0755)
	if err != nil {
		return err
	}
	added, err := _infoFileDir(id, create)
	if err != nil {
		return err
	}
	self.Added = append(self.Added, added)
	if self.Hashes == nil {
		self.Hashes = map[string]string{}
	}
	self.Hashes[name] = createHash(id, filepath.Join(path, name))
	changed, err := _infoFileDir(id, path)
	if err != nil {
		return err
	}
	self.Changed = append(self.Changed, changed)
	return nil
}

func (self *response) rm(id, path string) error {
	err := os.RemoveAll(filepath.Join(conf[id].Root, path))
	if err != nil {
		return err
	}
	self.Removed = append(self.Removed, createHash(id, path))
	return nil
}

func (self *response) rename(id, path, name string) error {
	newPath := filepath.Join(filepath.Dir(path), filepath.Base(name))
	err := os.Rename(filepath.Join(conf[id].Root, path), filepath.Join(conf[id].Root,newPath))
	if err != nil {
		return err
	}
	added, err := _infoFileDir(id, newPath)
	self.Added = append(self.Added, added)
	self.Removed = append(self.Removed, createHash(id, path))
	return nil
}

func (self *response) renames(id, path, suffix string, renames []string) error {
	if len(renames) != 0 {
		for _, r := range renames  {
			oldPath := filepath.Join(path, r)
			newPath := filepath.Join(path, strings.TrimRight(r, filepath.Ext(r)) + suffix + filepath.Ext(r))

			err := os.Rename(filepath.Join(conf[id].Root, oldPath), filepath.Join(conf[id].Root, newPath)) //ToDo suffix clean
			if err != nil {
				return err
			}
			added, err := _infoFileDir(id, newPath)
			self.Added = append(self.Added, added)
		}
	}
	return nil
}




func (self *response) ls(id, target string, intersect []string) {
	self.List = []string{}
	for _, i := range intersect {
		_, err := os.Stat(filepath.Join(conf[id].Root, target, i));
		if !os.IsNotExist(err) {
			self.List = append(self.List, i)
		}
	}
}



// ----------------------------------------------------------------------

