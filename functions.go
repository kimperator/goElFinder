package elFinder

import (
	"path/filepath"
	"os"
	"mime"
	"fmt"
	"strings"
	"io/ioutil"
	"io"
)

// Config functions
func (volume *volumeResponse) setRoot(path string) {
	volume.config.rootDir, _ = filepath.Abs(path)
	fmt.Println(volume.config.rootDir)
}

func (volume *volumeResponse) allowDirs(dirs []string) {
	if volume.config.dirsRight == nil {
		volume.config.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		volume.config.dirsRight[v] = true
	}
}

func (volume *volumeResponse) denyDirs(dirs []string) {
	if volume.config.dirsRight == nil {
		volume.config.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		volume.config.dirsRight[v] = false
	}
}

func (volume *volumeResponse) setDefaultRight(right bool) {
	volume.config.defaultRight = right
}


// Request functions
func (volume *volumeResponse) open(path string) error {
	var (
		err error
	)

	if volume.config.init {
		path = "/"
		root, err := volume._infoPath(volume.config.rootDir)
		if err != nil {
			return err
		}
		volume.Api = APIver
		volume.Cwd = root
	} else {
		obj, err := volume._infoPath(filepath.Join(volume.config.rootDir, path))
		if err != nil {
			return err
		}
		volume.Cwd = obj
	}

	err = filepath.Walk(filepath.Join(volume.config.rootDir, path), volume._info)
	if err != nil {
		return err
	}

	if volume.config.tree {

	}
	return nil
}

func (volume *volumeResponse) file(path string) (fileName, mimeType string, data []byte, err error) {
	target := filepath.Join(volume.config.rootDir, path)
	data, err = ioutil.ReadFile(target)
	fileName = filepath.Base(path)
	mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	return fileName, mimeType, data, err
}

func (volume *volumeResponse) mkdir(path, name string) error {
	create := filepath.Join(volume.config.rootDir, path, name)
	err := os.MkdirAll(create, 0777)
	if err != nil {
		return err
	}
	added, err := volume._infoPath(create)
	if err != nil {
		return err
	}
	volume.Added = append(volume.Added, added)
	if volume.Hashes == nil {
		volume.Hashes = map[string]string{}
	}
	volume.Hashes[name] = createHash("l0", filepath.Join(path, name))
	changed, err := volume._infoPath(filepath.Join(volume.config.rootDir, path))
	if err != nil {
		return err
	}
	volume.Changed = append(volume.Changed, changed)
	return nil
}

func (volume *volumeResponse) rm(path string) error {
	err := os.RemoveAll(filepath.Join(volume.config.rootDir, path))
	if err != nil {
		return err
	}
	volume.Removed = append(volume.Removed, createHash("l0", path))
	return nil
}

func (volume *volumeResponse) rename(path, name string) error {
	newPath := filepath.Join(volume.config.rootDir, filepath.Dir(path), filepath.Base(name))
	err := os.Rename(filepath.Join(volume.config.rootDir, path), newPath)
	if err != nil {
		return err
	}
	added, err := volume._infoPath(newPath)
	volume.Added = append(volume.Added, added)
	volume.Removed = append(volume.Removed, createHash("l0", path))
	return nil
}

func (volume *volumeResponse) upload(path, name string, file io.Reader) error {
	path = filepath.Join(volume.config.rootDir, path, name)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		volume.Warning = append(volume.Warning, err.Error())
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		volume.Warning = append(volume.Warning, err.Error())
		return err
	}
	info, err := f.Stat()
	if err != nil {
		volume.Warning = append(volume.Warning, err.Error())
		return err
	}
	i, err := volume._getFileDirInfo(path, info)
	if err != nil {
		volume.Warning = append(volume.Warning, err.Error())
	}
	volume.Added = append(volume.Added, i)
	return nil
}

func (volume *volumeResponse) dim(path string) error {
	var err error
	target := filepath.Join(volume.config.rootDir, path)
	volume.Dim, err = getImageDim(target)
	if err != nil {
		return err
	}
	return nil
}

func (volume *volumeResponse) checkRight(path string) bool {
	return volume._getRight(filepath.Join(volume.config.rootDir, path))
}

func (volume *volumeResponse) _getRight(path string) bool {
	path = volume._trimRootDir(path)
	if path == "" {
		return true
	}
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}
	for k, v := range volume.config.dirsRight {
		if strings.HasPrefix(path, k) {
			return v
		}
	}
	return volume.config.defaultRight
}

func (volume *volumeResponse) _infoPath(path string) (volumeFileDir, error) {
	u, err := os.Open(path)
	if err != nil {
		return volumeFileDir{}, err
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return volumeFileDir{}, err
	}
	return volume._getFileDirInfo(path, info)
}

func (volume *volumeResponse) _info(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if volume._getRight(path) {
		p, err := volume._getFileDirInfo(path, info)
		if err != nil {
			return err
		}
		volume.Files = append(volume.Files, p)
	}
	return nil
}

func (volume *volumeResponse) _getFileDirInfo(path string, info os.FileInfo) (volumeFileDir, error) {
	var p volumeFileDir
	p.Name = info.Name()
	p.Size = info.Size()
	p.Ts = info.ModTime().Unix()

	if path != volume.config.rootDir {
		if volume._trimRootDir(filepath.Join(path, ".." + string(filepath.Separator))) != "" { //ToDo
			p.Phash = createHash("l0", volume._trimRootDir(filepath.Join(path, ".." + string(filepath.Separator))))
		} else {
			p.Phash = createHash("l0", string(filepath.Separator))
		}
		p.Hash = createHash("l0", volume._trimRootDir(path))
	} else {
		p.Isroot = 1
		p.Hash = createHash("l0", string(filepath.Separator))
		p.Phash = ""
	}
	if info.IsDir() {
		p.Mime = "directory"
		p.Volumeid = "l0_"
		u, err := os.Open(path)
		if err != nil {
			return volumeFileDir{}, err
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

		if path == "/" {
			p.Phash = ""
			p.Isroot = 1
			p.Locked = 1
		}
	} else {
		p.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
	}

	//ToDo get permission
	p.Read = 1
	p.Write = 1

	return p, nil
}

func (volume *volumeResponse) _trimRootDir(path string) string {
	return strings.TrimPrefix(path, volume.config.rootDir)
}

/*
func (response *volumeResponse) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	var p volumeFileDir
	p.Name = info.Name()
	p.Size = info.Size()
	p.Ts = info.ModTime().Unix()
	p.Hash = createHash("l0", path)
	p.Phash = createHash("l0", filepath.Join(path, "../"))
	if info.IsDir() {
		p.Mime = "directory"
		p.Volumeid = "l0_"
		u, err := os.Open(filepath.Join(basePath, path))
		if err != nil {
			return err
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

		if path == "/" {
			p.Phash = ""
			p.Isroot = 1
			p.Locked = 1
		}
	} else {
		p.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
	}

	//ToDo get permission
	p.Read = 1
	p.Write = 1

	response.Files = append(response.Files, p)

	return nil
}

func parseInfo(path string, info os.FileInfo, err error) volumeFileDir {
	var p volumeFileDir
	if err != nil {
		log.Print(err)
		return p
	}
	p.Name = info.Name()
	p.Size = info.Size()
	p.Ts = info.ModTime().Unix()
	p.Hash = createHash("l0", path)
	p.Phash = createHash("l0", filepath.Join(path, "../"))
	if info.IsDir() {
		p.Mime = "directory"
		p.Volumeid = "l0_"
		u, err := os.Open(filepath.Join(basePath, path))
		if err != nil {
			return volumeFileDir{}
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

		if path == "/" {
			p.Phash = ""
			p.Isroot = 1
			p.Locked = 1
		}
	} else {
		p.Mime = mime.TypeByExtension(filepath.Ext(info.Name()))
	}

	//ToDo get permission
	p.Read = 1
	p.Write = 1

	return p

}



func volumeOpenDir(path string, response *volumeResponse) {

	response.Cwd = fileInfo(path)

	if isDir(path) {
		response.Files = parents(path)
	}
}

func parents(path string) []volumeFileDir {
	tree := []volumeFileDir{fileInfo("/")}
	from := path
	for filepath.Base(path) != "/" {
		ls, err := listDir(path)
		if err != nil {
			return tree
		}
		for _, l := range ls {
			t := filepath.Join(path,l)
			if t != from {
				tree = append(tree, fileInfo(t))
			}
		}
		path = filepath.Join(path, "../")
	}

	ls, err := listDir("/")
	if err != nil {
		return tree
	}
	for _, l := range ls {
		t := filepath.Join(path,l)
		if t != from {
			tree = append(tree, fileInfo(t))
		}

	}
	return tree
}


func listDir(path string) ([]string, error) {
	u, err := os.Open(filepath.Join(basePath, path))
	if err != nil {
		return []string{}, err
	}
	defer u.Close()

	return u.Readdirnames(0)
}

func getFile(path string) (string, []byte, error) {
	path = filepath.Join(basePath, path)
	data, err := ioutil.ReadFile(path)
	return filepath.Base(path), data, err
}


func fileInfo(path string) volumeFileDir {
	var p volumeFileDir
	u, err := os.Open(filepath.Join(basePath, path))
	if err != nil {
		return volumeFileDir{}
	}
	defer u.Close()

	s, err := u.Stat()
	if err != nil {
		return volumeFileDir{}
	}
	p.Name = s.Name()
	p.Size = s.Size()
	p.Ts = s.ModTime().Unix()
	p.Hash = "l0_" + encode64(path)
	p.Phash = "l0_" + encode64(filepath.Join(path, "../"))
	if s.IsDir() {
		p.Mime = "directory"
		p.Volumeid = "l0_"

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

		if path == "/" {
			p.Phash = ""
			p.Isroot = 1
			p.Locked = 1
		}
	} else {
		buf, _ := ioutil.ReadFile(filepath.Join(basePath, path))
		kind, err  := filetype.Match(buf)
		if err == nil {
			p.Mime = kind.MIME.Value
		}
	}

	//ToDo get permission
	p.Read = 1
	p.Write = 1

	return p
}

func isDir(path string) (bool) {
	u, err := os.Open(filepath.Join(basePath, path))
	if err != nil {
		return false
	}
	defer u.Close()

	s, err := u.Stat()
	if err != nil {
		return false
	}
	return s.IsDir()
}

func isFile(path string) (bool) {
	return !isDir(path)
}
*/