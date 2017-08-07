package goElFinder

import (
	"path/filepath"
	"errors"
	"log"
	"strings"
	"encoding/base64"
	"fmt"
	"os"
	"mime"
	"crypto/sha1"
	"io/ioutil"
)

func (self *elf) realPath(target string) (string, error) {
	t, err :=  self.volumes.parsePathHash(target)
	if err != nil {
		return "", err
	}

	if !self.volumes.getRight(t) {
		return "", errors.New("errLocked")
	}

	return filepath.Join(self.volumes[t.id].Root, t.path), nil
}

// Return real filesystem path ToDo remove and replace this
func (self *elf) realTargetPath(t target) string {
	return filepath.Join(self.volumes[t.id].Root, t.path)
}

// Parse request (target(s), destination(s), upload path) and check rights
func (self *elf) parse() (err error) {
	self.target, err = self.volumes.parsePathHash(self.req.Target)
	if err != nil {
		return err
	}
	if !self.volumes.getRight(self.target) { // ToDo check right with parse
		return errors.New("errLocked")
	}

	self.dst, err = self.volumes.parsePathHash(self.req.Dst)
	if err != nil {
		return err
	}
	if !self.volumes.getRight(self.dst) {
		return errors.New("errLocked")
	}

	self.src, err = self.volumes.parsePathHash(self.req.Src)
	if err != nil {
		return err
	}
	if !self.volumes.getRight(self.src) {
		return errors.New("errLocked")
	}

	for _, ft := range self.req.Targets {
		var p target
		p, err = self.volumes.parsePathHash(ft)
		if err != nil {
			log.Println(err)
		}
		if !self.volumes.getRight(p) {
			return errors.New("errLocked")
		}
		self.targets = append(self.targets, p)
	}

	if len(self.req.UploadPath) != 0 {
		for i := range self.req.UploadPath {
			var p target
			p, err = self.volumes.parsePathHash(self.req.UploadPath[i])
			if err != nil {
				log.Println(err)
			}
			if !self.volumes.getRight(p) {
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
			if !getRight(p) {
				return errors.New("errLocked")
			}
			self.dirs = append(self.dirs, p)
		}
	}
*/
	return nil
}

func (volumes Volumes) infoTarget(t target) (fd fileDir, err error) {
	if !volumes.getRight(t) {
		return fd, errors.New("Permission denied")
	}
	realpath := filepath.Join(volumes[t.id].Root, t.path)
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
		fd.Size = _size(volumes[t.id].Root,t.path)
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
			fd.Options.Path = volumes[t.id].Root
			fd.Options.Separator = string(os.PathSeparator)
			fd.Phash = ""
			fd.Isroot = 1
			fd.Locked = 0
		}

		fd.Mime = "directory"

		n, _ := u.Readdir(0)
		hasDir := func() byte {
			for _, t := range n {
				if t.IsDir() && !strings.HasPrefix(t.Name(), ".") {
					return 1
				}
			}
			return 0
		}
		fd.Dirs = hasDir()

		fd.Options.Url = volumes[t.id].Url + "/"
		//ToDo fd.Options.TmbUrl = volumes[id].Url + path + "/.tmb/"

		fd.Options.Archivers.Create = []string{ "application/zip" }
		fd.Options.Archivers.Extract = []string{ "application/zip" }
		fd.Options.Archivers.Createext = map[string]string{}
		fd.Options.Archivers.Createext["application/zip"] = "zip"

	} else {
		fd.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
		if fd.Mime == "image/jpeg" || fd.Mime == "image/png" || fd.Mime == "image/gif" {
			tmb := fmt.Sprintf("%x", sha1.Sum([]byte(filepath.Base(t.path))))+filepath.Ext(t.path)
			if _, err := os.Stat(filepath.Join(volumes[t.id].Root, filepath.Dir(t.path), ".tmb", tmb)); !os.IsNotExist(err) {
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

func (volumes Volumes) listDirs(t target) (dirs []string) {
	realpath := filepath.Join(volumes[t.id].Root, t.path)
	entries, err := ioutil.ReadDir(realpath) // ToDo use os.Readdirnames() ?
	if err != nil {
		return
	}
	for _,ent := range entries {
		if ent.IsDir() {
			if volumes.getRight(t) {
				dirs = append(dirs, filepath.Join(t.path, ent.Name()))
			}
		}
	}
	fmt.Println(dirs)
	return
}

func (volumes Volumes) listAll(t target) (all []string) {
	entries, err := ioutil.ReadDir(filepath.Join(volumes[t.id].Root, t.path))
	if err != nil {
		return
	}
	for _,ent := range entries {
		if volumes.getRight(t) {
			all = append(all, filepath.Join(t.path, ent.Name()))
		}
	}
	return
}

func (volumes Volumes) getRight(p target) bool {
	if p.path == "" || p.path == string(os.PathSeparator) {
		return true
	}
	if strings.HasPrefix(filepath.Base(p.path), ".") {
		return false
	}
	for _, v := range volumes[p.id].DenyDirs {
		if strings.HasPrefix(p.path, v) {
			return false
		}
	}
	for _, v := range volumes[p.id].AllowDirs {
		if strings.HasPrefix(p.path, v) {
			return true
		}
	}
	return volumes[p.id].DefaultRight
}

func (volumes Volumes) parsePathHash(tgt string) (target, error) { //ToDo check file name
	splitTarget := strings.SplitN(tgt, "_", 2)
	var (
		vi, vp string
		err error
	)
	if len(splitTarget) != 2 {
		//return volume, path, errors.New("Bad target")
		for k := range volumes {
			if volumes[k].Default {
				vi = k
//				log.Println("Select default volume:", k)
				break
			}
		}
	} else {
		vi = splitTarget[0]
		vp, err = decode64(splitTarget[1])
		if err != nil {
			return target{id: vi, path: vp}, errors.New("Bad base64 path")
		}
	}

	if _, ok := volumes[vi]; !ok {
		return target{id: vi, path: vp}, errors.New("Bad volume id")
	}

	// Clean path
	if vp == "" {
		vp = string(os.PathSeparator)
	} else {
		vp = strings.TrimPrefix(filepath.Clean(vp), "..")
		vp = strings.TrimPrefix(filepath.Clean(vp), string(os.PathSeparator) + "..")
	}

	return target{id: vi, path: vp}, err
}

// Code/decode functions
func decode64(s string) (string, error) {
	str := strings.Replace(s, " ", "+",-1)
	t, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(t), nil
}

func encode64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func createHash(volumeId, path string) string {
	return volumeId + "_" + encode64(path)
}

