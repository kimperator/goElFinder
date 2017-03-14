package goElFinder

import (
	"os"
	"mime"
	"path/filepath"
	"fmt"
	"strings"
	"io/ioutil"
	"errors"
	"crypto/sha1"
)

func _infoFileDir(id, path string) (fd fileDir, err error) {
	//path = filepath.Clean(path)
	if !_getRight(id, path) {
		return fd, errors.New("Permission denied")
	}
	realpath := filepath.Join(conf[id].Root, path)
	fmt.Printf("Info FileDir. Id: %s path: %s realpath: %s\n", id, path, realpath)
	u, err := os.Open(realpath)
	if err != nil {
		return fd, err
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return fd, err
	}

	fd.Volumeid = id
	fd.Name = info.Name()
	fd.Size = info.Size()
	fd.Ts = info.ModTime().Unix()
	fd.Phash = createHash(id, filepath.Join(path, ".." + string(os.PathSeparator)))
	fd.Hash = createHash(id, path)

	if info.IsDir() {
		if path == string(os.PathSeparator) || path == "" {
			fd.Isroot = 1
			fd.Hash = createHash(id, string(os.PathSeparator))
			fd.Phash = ""
			fd.Options.Path = conf[id].Root
			fd.Options.Separator = string(os.PathSeparator)
			fd.Phash = ""
			fd.Isroot = 1
			fd.Locked = 0
		}

		fd.Mime = "directory"

		n, _ := u.Readdir(0)
		hasDir := func() byte {
			for _, t := range n {
				if t.IsDir() {
					return 1
				}
			}
			return 0
		}
		fd.Dirs = hasDir()

		fd.Options.Url = conf[id].Url + "/"
		//ToDo fd.Options.TmbUrl = conf[id].Url + path + "/.tmb/"

		fd.Options.Archivers.Create = []string{ "application/zip" }
		fd.Options.Archivers.Extract = []string{ "application/zip" }
		fd.Options.Archivers.Createext = map[string]string{}
		fd.Options.Archivers.Createext["application/zip"] = "zip"

	} else {
		fd.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
		if fd.Mime == "image/jpeg" || fd.Mime == "image/png" || fd.Mime == "image/gif" {
			tmb := fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(path))))+filepath.Ext(path)
			if _, err := os.Stat(filepath.Join(conf[id].Root, filepath.Dir(path), ".tmb", tmb)); !os.IsNotExist(err) {
				fd.Tmb = tmb
				fmt.Println("Tmb for", fd.Name, "yes")
			} else {
				fd.Tmb = "1"
				fmt.Println("Tmb for", fd.Name, "no")
			}
		}
	}

	//ToDo get permission
	fd.Read = 1
	fd.Write = 1

	return fd, nil
}

func _listDirs(id, path string) (dirs []string) {
	realpath := filepath.Join(conf[id].Root, path)
	entries, err := ioutil.ReadDir(realpath)
	if err != nil {
		return
	}
	for _,ent := range entries {
		if ent.IsDir() {
			if _getRight(id, path) {
				dirs = append(dirs, filepath.Join(path, ent.Name()))
			}
		}
	}
	fmt.Println(dirs)
	return
}

func _listAll(id, path string) (all []string) {
	entries, err := ioutil.ReadDir(filepath.Join(conf[id].Root, path))
	if err != nil {
		return
	}
	for _,ent := range entries {
		if _getRight(id, path) {
			all = append(all, filepath.Join(path, ent.Name()))
		}
	}
	return
}

func _getRight(id, path string) bool {
	if path == "" || path == string(os.PathSeparator) {
		return true
	}
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}
	for _, v := range conf[id].DenyDirs {
		if strings.HasPrefix(path, v) {
			return false
		}
	}
	for _, v := range conf[id].AllowDirs {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	return conf[id].DefaultRight
}





















/*
func (self *response) getRight(path string) bool {
	return self._getRight(filepath.Join(self.current.rootDir, path))
}

func (self *response) _getRight(path string) bool {
	path = self._trimRootDir(path)
	if path == "" {
		return true
	}
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}
	for k, v := range self.current.dirsRight {
		if strings.HasPrefix(path, k) {
			return v
		}
	}
	return self.current.defaultRight
}

func (self *response) _info(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if self._getRight(path) {
		p, err := self._getFileDirInfo(path, info)
		if err != nil {
			return err
		}
		self.Files = append(self.Files, p)
	}
	return nil
}

func (self *response) _infoPath(path string) (fileDir, error) {
	u, err := os.Open(path)
	if err != nil {
		return fileDir{}, err
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return fileDir{}, err
	}
	return self._getFileDirInfo(path, info)
}

func (self *response) _getFileDirInfo(path string, info os.FileInfo) (fileDir, error) {
	var p fileDir
	p.Name = info.Name()
	p.Size = info.Size()
	p.Ts = info.ModTime().Unix()

	if path != self.current.rootDir {
		if self._trimRootDir(filepath.Join(path, ".." + string(os.PathSeparator))) != "" { //ToDo
			p.Phash = createHash(self.current.id, self._trimRootDir(filepath.Join(path, ".." + string(os.PathSeparator))))
		} else {
			p.Phash = createHash(self.current.id, string(os.PathSeparator))
		}
		p.Hash = createHash(self.current.id, self._trimRootDir(path))
	} else {
		p.Isroot = 1
		p.Hash = createHash(self.current.id, string(os.PathSeparator))
		p.Phash = ""
	}
	if info.IsDir() {
		p.Mime = "directory"
		p.Volumeid = self.current.id
		u, err := os.Open(path)
		if err != nil {
			return fileDir{}, err
		}
		defer u.Close()
		n, _ := u.Readdir(0)
		hasDir := func() byte {
			for _, t := range n {
				if t.IsDir() {
					return 1
				}
			}
			return 0
		}
		p.Dirs = hasDir()

		self.Options.Archivers.Create = []string{ "application/zip" }
		self.Options.Archivers.Extract = []string{ "application/zip" }
		self.Options.Archivers.Createext = map[string]string{}
		self.Options.Archivers.Createext["application/zip"] = "zip"

		fmt.Println("---",path)
		if path == self.current.rootDir {
			self.Options.Path = filepath.Base(self.current.rootDir)
			self.Options.Separator = "/"
			self.Options.Url = self.current.url
			self.Options.TmbUrl = self.current.url + ".tmb/"

			p.Phash = ""
			p.Isroot = 1
			p.Locked = 1
		}
	} else {
		p.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
		if p.Mime == "image/jpeg" || p.Mime == "image/png" || p.Mime == "image/gif" {
			tmb := fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(path))))+filepath.Ext(path)
			if _, err := os.Stat(filepath.Join(self.current.rootDir, ".tmb", tmb)); !os.IsNotExist(err) {
				p.Tmb = tmb
				fmt.Println("Tmb for", p.Name, "yes")
			} else {
				p.Tmb = "1"
				fmt.Println("Tmb for", p.Name, "no")
			}
		}
	}

	//ToDo get permission
	p.Read = 1
	p.Write = 1

	return p, nil
}

func (self *response) _trimRootDir(path string) string {
	return strings.TrimPrefix(path, self.current.rootDir)
}

*/
