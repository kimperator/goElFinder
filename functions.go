package elFinder

import (
	"path/filepath"
	"os"
	"mime"
	"fmt"
	"strings"
	"io/ioutil"
	"io"
	"strconv"
	"regexp"
	"errors"
)

// Config functions
func (volume *response) setRoot(path string) {
	volume.config.rootDir, _ = filepath.Abs(path)
	fmt.Println(volume.config.rootDir)
}

func (volume *response) allowDirs(dirs []string) {
	if volume.config.dirsRight == nil {
		volume.config.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		volume.config.dirsRight[v] = true
	}
}

func (volume *response) denyDirs(dirs []string) {
	if volume.config.dirsRight == nil {
		volume.config.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		volume.config.dirsRight[v] = false
	}
}

func (volume *response) setDefaultRight(right bool) {
	volume.config.defaultRight = right
}

// Request functions
func (volume *response) open(path string) error {
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

func (volume *response) file(path string) (fileName, mimeType string, data []byte, err error) {
	target := filepath.Join(volume.config.rootDir, path)
	data, err = ioutil.ReadFile(target)
	fileName = filepath.Base(path)
	mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	return fileName, mimeType, data, err
}

func (volume *response) mkdir(path, name string) error {
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

func (volume *response) rm(path string) error {
	err := os.RemoveAll(filepath.Join(volume.config.rootDir, path))
	if err != nil {
		return err
	}
	volume.Removed = append(volume.Removed, createHash("l0", path))
	return nil
}

func (volume *response) rename(path, name string) error {
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

func (volume *response) upload(path, name string, file io.Reader) error {
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

func (volume *response) chunkUpload(cid int, target, chunk string, file io.Reader) error {
	//	fmt.Println("Chunk", cid, target, chunk, file)
	if cid != 0 {
		if file != nil {
			tmpPath := filepath.Join(volume.config.rootDir, target, fmt.Sprintf(".%d_%s~", cid, chunk))
			f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				volume.Warning = append(volume.Warning, err.Error())
				return err
			}
			_, err = io.Copy(f, file)
			if err != nil {
				volume.Warning = append(volume.Warning, err.Error())
				return err
			}
			f.Close()
			os.Rename(tmpPath, filepath.Join(volume.config.rootDir, target, fmt.Sprintf(".%d_%s", cid, chunk)))
		}

		// check complete ---------------------------------------------------
		re := regexp.MustCompile(`(.*?)(\.[0-9][0-9]*?_[0-9][0-9]*?)(\.part)`)
		ch := re.FindStringSubmatch(chunk)
		if len(ch) != 4 {
			return errors.New("Bad chunk name format")
		}
		name := ch[1]
		t := strings.Split(ch[2], "_")
		total, err := strconv.Atoi(t[1])
		if err != nil {
			return err
		}
		allComplete := func() bool {
			for i := 0; i <= total; i++ {
				//fmt.Println("Check chank:", filepath.Join(volume.config.rootDir, target, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, i, total)))
				if _, err := os.Stat(filepath.Join(volume.config.rootDir, target, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, i, total))); os.IsNotExist(err) {
					return false
				}
			}
			return true
		}
		complete := allComplete()
		// -----------------------------------------------------------------

		if complete {
			volume.Chunkmerged = fmt.Sprintf(".%d_%s.%d_part", cid, name, total)
			volume.Name = name
		}
		fmt.Println("Check chunk result:", complete)
		volume.Added = []fileDir{}

	} else {
		var err error
		re := regexp.MustCompile(`(\.[0-9][0-9]*?)(_.*?)(\.[0-9][0-9]*?_part)`)
		ch := re.FindStringSubmatch(chunk)
		if len(ch) != 4 {
			return errors.New("Bad merged chunk name format")
		}
		cid, err = strconv.Atoi(ch[1][1:])
		if err != nil {
			return err
		}
		name := ch[2][1:]
		total, err :=  strconv.Atoi(strings.TrimRight(ch[3][1:], "_part"))
		if err != nil {
			return err
		}

		targetPath := filepath.Join(volume.config.rootDir, target, name)
		os.Rename(filepath.Join(volume.config.rootDir, target, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, 0, total)), targetPath)
		f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			volume.Warning = append(volume.Warning, err.Error())
			return err
		}
		defer f.Close()
		for i := 1; i <= total; i++ {
			chunkPath := filepath.Join(volume.config.rootDir, target, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, i, total))
			c, err := os.OpenFile(chunkPath, os.O_RDONLY, 0666)
			if err != nil {
				return err
			}
			cStat, err := c.Stat()
			if err != nil {
				return err
			}
			b := make([]byte,cStat.Size())
			_, err = c.Read(b)
			if err != nil {
				return err
			}
			_, err = f.Write(b)
			if err != nil {
				return err
			}
			//fmt.Println("Read", chunkPath, "read:", nr, "bytes write:", nw, "bytes")
			c.Close()
			err = os.Remove(chunkPath)
			if err != nil {
				return err
			}
		}
		fStat, err := f.Stat()
		if err != nil {
			return err
		}
		fInfo, err := volume._getFileDirInfo(targetPath, fStat)
		if err != nil {
			return err
		}
		volume.Added = append(volume.Added, fInfo)
		//fmt.Println("End chunk merge files. Cid:", cid, "Target path:", targetPath, "Total:", total+1, "part")
	}

	return nil
}

func (volume *response) dim(path string) error {
	var err error
	target := filepath.Join(volume.config.rootDir, path)
	volume.Dim, err = getImageDim(target)
	if err != nil {
		return err
	}
	return nil
}

func (volume *response) checkRight(path string) bool {
	return volume._getRight(filepath.Join(volume.config.rootDir, path))
}

func (volume *response) _getRight(path string) bool {
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

func (volume *response) _infoPath(path string) (fileDir, error) {
	u, err := os.Open(path)
	if err != nil {
		return fileDir{}, err
	}
	defer u.Close()
	info, err := u.Stat()
	if err != nil {
		return fileDir{}, err
	}
	return volume._getFileDirInfo(path, info)
}

func (volume *response) _info(path string, info os.FileInfo, err error) error {
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

func (volume *response) _getFileDirInfo(path string, info os.FileInfo) (fileDir, error) {
	var p fileDir
	p.Name = info.Name()
	p.Size = info.Size()
	p.Ts = info.ModTime().Unix()

	if path != volume.config.rootDir {
		if volume._trimRootDir(filepath.Join(path, ".." + string(os.PathSeparator))) != "" { //ToDo
			p.Phash = createHash("l0", volume._trimRootDir(filepath.Join(path, ".." + string(os.PathSeparator))))
		} else {
			p.Phash = createHash("l0", string(os.PathSeparator))
		}
		p.Hash = createHash("l0", volume._trimRootDir(path))
	} else {
		p.Isroot = 1
		p.Hash = createHash("l0", string(os.PathSeparator))
		p.Phash = ""
	}
	if info.IsDir() {
		p.Mime = "directory"
		p.Volumeid = "l0_"
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

		if path == string(os.PathSeparator) {
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

func (volume *response) _trimRootDir(path string) string {
	return strings.TrimPrefix(path, volume.config.rootDir)
}