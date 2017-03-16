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

// Config functions

func parsePathHash(tgt string) (target, error) { //ToDo check file name
	splitTarget := strings.SplitN(tgt, "_", 2)
	var (
		vi, vp string
		err error
	)
	if len(splitTarget) != 2 {
		//return volume, path, errors.New("Bad target")
		for k := range conf {
			if conf[k].Default == true {
				vi = k
				log.Println("Select default volume:", k)
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

	if _, ok := conf[vi]; !ok {
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