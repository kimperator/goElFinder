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

func _infoFileDir(t target) (fd fileDir, err error) {
	//path = filepath.Clean(path)
	if !_getRight(t) {
		return fd, errors.New("Permission denied")
	}
	realpath := filepath.Join(conf[t.id].Root, t.path)
	u, err := os.Open(realpath)
	if err != nil {
		return fd, err
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return fd, err
	}

	fd.Volumeid = t.id
	fd.Name = info.Name()

	if info.IsDir() {
		fd.Size = _size(t.id,t.path)
	} else {
		fd.Size = info.Size()
	}

	fd.Ts = info.ModTime().Unix()
	fd.Phash = createHash(t.id, filepath.Join(t.path, ".." + string(os.PathSeparator)))
	fd.Hash = createHash(t.id, t.path)

	if info.IsDir() {
		if t.path == string(os.PathSeparator) || t.path == "" {
			fd.Isroot = 1
			fd.Hash = createHash(t.id, string(os.PathSeparator))
			fd.Phash = ""
			fd.Options.Path = conf[t.id].Root
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

		fd.Options.Url = conf[t.id].Url + "/"
		//ToDo fd.Options.TmbUrl = conf[id].Url + path + "/.tmb/"

		fd.Options.Archivers.Create = []string{ "application/zip" }
		fd.Options.Archivers.Extract = []string{ "application/zip" }
		fd.Options.Archivers.Createext = map[string]string{}
		fd.Options.Archivers.Createext["application/zip"] = "zip"

	} else {
		fd.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
		if fd.Mime == "image/jpeg" || fd.Mime == "image/png" || fd.Mime == "image/gif" {
			tmb := fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(t.path))))+filepath.Ext(t.path)
			if _, err := os.Stat(filepath.Join(conf[t.id].Root, filepath.Dir(t.path), ".tmb", tmb)); !os.IsNotExist(err) {
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

func _listDirs(t target) (dirs []string) {
	realpath := filepath.Join(conf[t.id].Root, t.path)
	entries, err := ioutil.ReadDir(realpath) // ToDo use os.Readdirnames() ?
	if err != nil {
		return
	}
	for _,ent := range entries {
		if ent.IsDir() {
			if _getRight(t) {
				dirs = append(dirs, filepath.Join(t.path, ent.Name()))
			}
		}
	}
	fmt.Println(dirs)
	return
}

func _listAll(t target) (all []string) {
	entries, err := ioutil.ReadDir(filepath.Join(conf[t.id].Root, t.path))
	if err != nil {
		return
	}
	for _,ent := range entries {
		if _getRight(t) {
			all = append(all, filepath.Join(t.path, ent.Name()))
		}
	}
	return
}

func _getRight(p target) bool {
	if p.path == "" || p.path == string(os.PathSeparator) {
		return true
	}
	if strings.HasPrefix(filepath.Base(p.path), ".") {
		return false
	}
	for _, v := range conf[p.id].DenyDirs {
		if strings.HasPrefix(p.path, v) {
			return false
		}
	}
	for _, v := range conf[p.id].AllowDirs {
		if strings.HasPrefix(p.path, v) {
			return true
		}
	}
	return conf[p.id].DefaultRight
}

func _size(id, path string) int64 {
	if !_getRight(target{id: id, path: path}) {
		return 0
	}
	var size int64 = 0
	u, err := os.Open(filepath.Join(conf[id].Root, path))
	if err != nil {
		return 0
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return 0
	}
	if info.IsDir() {
		entries, err := ioutil.ReadDir(filepath.Join(conf[id].Root, path))
		if err != nil {
			return 0
		}
		for _, e := range entries {
			if info.IsDir() {
				size = size + _size(id, filepath.Join(path, e.Name()))
			} else {
				size =  size + e.Size()
			}
		}
	} else {
		size = info.Size()
	}
	return size
}
