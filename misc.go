package goElFinder

import (
	"os"
	"mime"
	"path/filepath"
	"fmt"
	"crypto/sha1"
	"strings"
)

func (self *response) checkRight(path string) bool {
	return self._getRight(filepath.Join(self.config.rootDir, path))
}

func (self *response) _getRight(path string) bool {
	path = self._trimRootDir(path)
	if path == "" {
		return true
	}
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}
	for k, v := range self.config.dirsRight {
		if strings.HasPrefix(path, k) {
			return v
		}
	}
	return self.config.defaultRight
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

func (self *response) _getFileDirInfo(path string, info os.FileInfo) (fileDir, error) {
	var p fileDir
	p.Name = info.Name()
	p.Size = info.Size()
	p.Ts = info.ModTime().Unix()

	if path != self.config.rootDir {
		if self._trimRootDir(filepath.Join(path, ".." + string(os.PathSeparator))) != "" { //ToDo
			p.Phash = createHash(self.config.id, self._trimRootDir(filepath.Join(path, ".." + string(os.PathSeparator))))
		} else {
			p.Phash = createHash(self.config.id, string(os.PathSeparator))
		}
		p.Hash = createHash(self.config.id, self._trimRootDir(path))
	} else {
		p.Isroot = 1
		p.Hash = createHash(self.config.id, string(os.PathSeparator))
		p.Phash = ""
	}
	if info.IsDir() {
		p.Mime = "directory"
		p.Volumeid = self.config.id
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

		self.Options.Archivers.Create = append( self.Options.Archivers.Create, "application/zip" )
		self.Options.Archivers.Extract = append( self.Options.Archivers.Extract, "application/zip" )
		self.Options.Archivers.Createext = map[string]string{}
		self.Options.Archivers.Createext["application/zip"] = "zip"

		if path == string(os.PathSeparator) {
			p.Phash = ""
			p.Isroot = 1
			p.Locked = 1
		}
	} else {
		p.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
		if p.Mime == "image/jpeg" || p.Mime == "image/png" || p.Mime == "image/gif" {
			tmb := fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(path))))+filepath.Ext(path)
			if _, err := os.Stat(filepath.Join(self.config.rootDir, ".tmb", tmb)); !os.IsNotExist(err) {
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
	return strings.TrimPrefix(path, self.config.rootDir)
}
