package goElFinder

import (
	"path/filepath"
	"log"
	"strings"
	"os"
	"errors"
)

type Volumes map[string]Local

type Local struct {
	Default bool
	Name string
	Root string
	Url string
	DefaultRight bool
	AllowDirs []string
	DenyDirs []string
} //ToDo WebDir and alias


type config struct {
	id string
	rootDir string // ToDo [name]realPath???
	url string
	dirsRight map[string]bool
	defaultRight bool
}

/*/ Config functions
func (self *response) setRoot(path string) {
	var err error
	self.current.rootDir, err = filepath.Abs(path)
	if err != nil {
		log.Print(err)
	}
}

func (self *response) allowDirs(dirs []string) {
	if self.current.dirsRight == nil {
		self.current.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		self.current.dirsRight[v] = true
	}

}

func (self *response) denyDirs(dirs []string) {
	if self.current.dirsRight == nil {
		self.current.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		self.current.dirsRight[v] = false
	}
}

func (self *response) setDefaultRight(right bool) {
	self.current.defaultRight = right
}

func (self *response) setUrl(url string) {
	self.current.url = url
}
*/

func parsePathHash(config Volumes, target string) (id, path string, err error) { //ToDo check file name
	splitTarget := strings.SplitN(target, "_", 2)
	if len(splitTarget) != 2 {
		//return volume, path, errors.New("Bad target")
		for k := range config {
			if config[k].Default == true {
				id = k
				log.Println("Select default volume:", k)
				break
			}
		}
	} else {
		id = splitTarget[0]
		path, err = decode64(splitTarget[1])
		if err != nil {
			return id, path, errors.New("Bad base64 path")
		}
	}

	if _, ok := config[id]; !ok {
		return id, path, errors.New("Bad volume id")
	}

	// Clean path
	if path == "" {
		path = string(os.PathSeparator)
	} else {
		path = strings.TrimPrefix(filepath.Clean(path), "..")
		path = strings.TrimPrefix(filepath.Clean(path), string(os.PathSeparator) + "..")
	}

	/*/var volume response
	volume.current.id = id
	volume.setRoot(config[id].Root)
	volume.setUrl(config[id].Url)
	volume.setDefaultRight(config[id].DefaultRight)
	volume.allowDirs(config[id].AllowDirs)
	volume.denyDirs(config[id].DenyDirs)

	volume.volumes = config
*/
	return id, path, err
}