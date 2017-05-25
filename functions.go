package goElFinder

import (
	"path/filepath"
	"os"
	"mime"
	"strings"
	"io/ioutil"
	"errors"
	"log"
)

func (self *elf) _getRealPath(t target) string {
	return filepath.Join(conf[t.id].Root, t.path)
}

func (self *elf) _parse() (err error) {
	self.target, err = parsePathHash(self.req.Target)
	if err != nil {
		return err
	}
	if !_getRight(self.target) { // ToDo check right with parse
		return errors.New("errLocked")
	}

	self.dst, err = parsePathHash(self.req.Dst)
	if err != nil {
		return err
	}
	if !_getRight(self.dst) {
		return errors.New("errLocked")
	}

	self.src, err = parsePathHash(self.req.Src)
	if err != nil {
		return err
	}
	if !_getRight(self.src) {
		return errors.New("errLocked")
	}

	for _, ft := range self.req.Targets {
		var p target
		p, err = parsePathHash(ft)
		if err != nil {
			log.Println(err)
		}
		if !_getRight(p) {
			return errors.New("errLocked")
		}
		self.targets = append(self.targets, p)
	}

	if len(self.req.UploadPath) != 0 {
		for i := range self.req.UploadPath {
			var p target
			p, err = parsePathHash(self.req.UploadPath[i])
			if err != nil {
				log.Println(err)
			}
			if !_getRight(p) {
				return errors.New("errLocked")
			}
			self.uploadpath = append(self.uploadpath, p)
		}
	}
/*
	if len(self.req.Dirs) != 0 {
		for i := range self.req.Dirs {
			var p target
			p, err = parsePathHash(self.req.Dirs[i])
			if err != nil {
				log.Println(err)
			}
			if !_getRight(p) {
				return errors.New("errLocked")
			}
			self.dirs = append(self.dirs, p)
		}
	}
*/
	return nil
}

// Request functions
func (self *elf) open() error {
	var err error

	if self.req.Init {
		self.res.Api = APIver
	}

	//obj, err := self._infoPath(filepath.Join(self.current.rootDir, target))
	obj, err := _infoFileDir(self.target)
	if err != nil {
		return err
	}
	self.res.Cwd = obj
	self.res.Files = []fileDir{}
	if self.req.Tree {
		for k := range conf {
				p, err := _infoFileDir(target{id: k})
				if err == nil {
					p.Options.Url = conf[k].Url
					self.res.Files = append(self.res.Files, p)
				}
		}
	}

	fd := _listAll(self.target)
	for _, f := range fd {
		i, err := _infoFileDir(target{id: self.target.id, path: f})
		if err == nil {
			self.res.Files = append(self.res.Files, i)
		}
	}
	return nil
}

func (self *elf) tree(t target) error {
	fd := _listDirs(t)
	self.res.Tree = []fileDir{}
	for _, f := range fd {
		if i, err := _infoFileDir(target{id:t.id, path:f}); err == nil {
			self.res.Tree = append(self.res.Tree, i)
		}
	}
	return nil
}

func (self *elf) parents(t target) error {
	self.res.Tree = []fileDir{}

	for t.path != string(filepath.Separator) {
		t.path = filepath.Join(t.path, "..")
		fd := _listDirs(t)
		for _, f := range fd {
			if i, err := _infoFileDir(target{id: t.id, path: f}); err == nil {
				self.res.Tree = append(self.res.Tree, i)
			}
		}
	}
	for k := range conf {
		p, err := _infoFileDir(target{id: k, path: ""})
		if err == nil {
			p.Options.Url = conf[k].Url
			self.res.Tree = append(self.res.Tree, p)
		}
	}
	return nil
}

func (self *elf) size() int64 {
	var size int64
	for _, p := range self.targets {
		s := _size(p.id, p.path)
		size = size + s
	}

	return size
}
func (self *elf) file() (fileName, mimeType string, data []byte, err error) {
	if _getRight(self.target) {
		path := filepath.Join(conf[self.target.id].Root, self.target.path)
		data, err = ioutil.ReadFile(path)
		fileName = filepath.Base(path)
		mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	} else {
		err = errors.New("Permission denied")
	}

	return fileName, mimeType, data, err
}

func (self *elf) mkdir() error {
	create := filepath.Join(self.target.path, self.req.Name)
	err := os.MkdirAll(filepath.Join(conf[self.target.id].Root, create), 0755)
	if err != nil {
		return err
	}
	added, err := _infoFileDir(target{id: self.target.id, path: create})
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
		err := os.MkdirAll(filepath.Join(conf[self.target.id].Root, self.target.path, d), 0755)
		if err != nil {
			return err
		}
		added, err := _infoFileDir(target{id: self.target.id, path: filepath.Join(self.target.path, d)})
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

func (self *elf) rm() error {
	for i := range self.req.Targets {
		err := os.RemoveAll(filepath.Join(conf[self.targets[i].id].Root, self.targets[i].path))
		if err != nil {
			return err
		}
		self.res.Removed = append(self.res.Removed, createHash(self.targets[i].id, self.targets[i].path))
	}

	return nil
}

func (self *elf) rename(id, path string) error {
	newPath := filepath.Join(filepath.Dir(path), filepath.Base(self.req.Name))
	err := os.Rename(filepath.Join(conf[id].Root, path), filepath.Join(conf[id].Root, newPath))
	if err != nil {
		return err
	}
	added, err := _infoFileDir(target{id: id, path: newPath})
	self.res.Added = append(self.res.Added, added)
	self.res.Removed = append(self.res.Removed, createHash(id, path))
	return nil
}

func (self *elf) renames(id, path string) error {
	if len(self.req.Renames) != 0 {
		for _, r := range self.req.Renames  {
			oldPath := filepath.Join(path, r)
			newPath := filepath.Join(path, strings.TrimRight(r, filepath.Ext(r)) + self.req.Suffix + filepath.Ext(r))

			err := os.Rename(filepath.Join(conf[id].Root, oldPath), filepath.Join(conf[id].Root, newPath)) //ToDo suffix clean
			if err != nil {
				return err
			}
			added, err := _infoFileDir(target{id: id, path: newPath})
			self.res.Added = append(self.res.Added, added)
		}
	}
	return nil
}

func (self *elf) ls() {
	self.res.List = []string{}
	for _, i := range self.req.Intersect {
		_, err := os.Stat(filepath.Join(conf[self.target.id].Root, self.target.path, i));
		if !os.IsNotExist(err) {
			self.res.List = append(self.res.List, i)
		}
	}
}

func (self *elf) mkfile() error {

	err := os.MkdirAll(filepath.Join(conf[self.target.id].Root, self.target.path), 0755)
	if err != nil {
		return err
	}
	create := filepath.Join(self.target.path, self.req.Name)
	f, err := os.OpenFile(filepath.Join(conf[self.target.id].Root, create), os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	f.Close()

	added, err := _infoFileDir(target{id: self.target.id, path: create})
	if err != nil {
		return err
	}

	self.res.Added = append(self.res.Added, added)

	return nil
}

func (self *elf) get() error {
	b, err := ioutil.ReadFile(filepath.Join(conf[self.target.id].Root, self.target.path))
	if err != nil {
		return err
	}
	self.res.Content = string(b)
	return nil
}

func (self *elf) put() error {
	err := ioutil.WriteFile(filepath.Join(conf[self.target.id].Root, self.target.path), []byte(self.req.Content), 0666)
	if err != nil {
		return err
	}
	info, err := _infoFileDir(self.target)
	if err != nil {
		return err
	}
	self.res.Changed = append(self.res.Changed, info)
	return nil
}

func (self *elf) url() {
	self.res.Url = conf[self.target.id].Url + self.target.path
}

func (self *elf) paste() error {
	for _, t := range self.targets {
		info, err := os.Stat(self._getRealPath(t))
		if err != nil {
			return err
		}
		if info.IsDir() {
			//ToDo check isset destination
			err = copyDir(self._getRealPath(t), filepath.Join(self._getRealPath(self.dst),filepath.Base(t.path)))
			if err != nil {
				return err
			}
		} else {
			err = copyFile(self._getRealPath(t), filepath.Join(self._getRealPath(self.dst),filepath.Base(t.path)))
			if err != nil {
				return err
			}
		}
		added, err := _infoFileDir(target{id: self.dst.id, path: filepath.Join(self.dst.path, filepath.Base(t.path))})
		if err != nil {
			return err
		}
		self.res.Added = append(self.res.Added, added)
		if self.req.Cut {
			err = os.RemoveAll(self._getRealPath(t))
			if err != nil {
				return err
			}
			self.res.Removed = append(self.res.Removed, createHash(t.id, t.path))
		}
	}

	return nil
}