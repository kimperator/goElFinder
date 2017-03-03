package goElFinder

import (
	"path/filepath"
	"log"
	"strings"
	"os"
	"errors"
)

type Config map[string]Volume

type config struct {
	id string
	rootDir string // ToDo [name]realPath???
	dirsRight map[string]bool
	defaultRight bool
}

// Config functions
func (self *response) setRoot(path string) {
	var err error
	self.config.rootDir, err = filepath.Abs(path)
	if err != nil {
		log.Print(err)
	}
}

func (self *response) allowDirs(dirs []string) {
	if self.config.dirsRight == nil {
		self.config.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		self.config.dirsRight[v] = true
	}

}

func (self *response) denyDirs(dirs []string) {
	if self.config.dirsRight == nil {
		self.config.dirsRight = map[string]bool{}
	}
	for _, v := range dirs {
		self.config.dirsRight[v] = false
	}
}

func (self *response) setDefaultRight(right bool) {
	self.config.defaultRight = right
}


func parsePathHash(config Config, target string) (volume response, path string, err error) { //ToDo check file name
	var volumeId string
	splitTarget := strings.SplitN(target, "_", 2)
	if len(splitTarget) != 2 {
		//return volume, path, errors.New("Bad target")
		for k := range config {
			volumeId = k
			log.Println("Select first volume:", k)
			break
		}
	} else {
		volumeId = splitTarget[0]
		path, err = decode64(splitTarget[1])
		if err != nil {
			return volume, path, errors.New("Bad base64 path")
		}
	}

	if _, ok := config[volumeId]; !ok {
		return volume, path, errors.New("Bad volume id")
	}

	if path == "" {
		path = string(os.PathSeparator)
	} else {
		path = strings.TrimPrefix(filepath.Clean(path), "..")
		path = strings.TrimPrefix(filepath.Clean(path), string(os.PathSeparator) + "..")
	}

	//var volume response
	volume.config.id = volumeId
	volume.setRoot(config[volumeId].Root)
	volume.setDefaultRight(config[volumeId].DefaultRight)
	volume.allowDirs(config[volumeId].AllowDirs)
	volume.denyDirs(config[volumeId].DenyDirs)

	return volume, path, err
}