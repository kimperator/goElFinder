package goElFinder

import (
	"os"
	"path/filepath"
	"fmt"
	"io"
//	"github.com/Unknwon/com"
)

func (self *elf) paste() error {
	for _, t := range self.targets {
		info, err := os.Stat(self.getRealPath(t))
		if err != nil {
			return err
		}
		if info.IsDir() {
			//ToDo check isset destination
			err = copyDir(self.getRealPath(t), filepath.Join(self.getRealPath(self.dst),filepath.Base(t.path)))
			if err != nil {
				return err
			}
		} else {
			err = copyFile(self.getRealPath(t), filepath.Join(self.getRealPath(self.dst),filepath.Base(t.path)))
			if err != nil {
				return err
			}
		}
		added, err := self.volumes.infoFileDir(target{id: self.dst.id, path: filepath.Join(self.dst.path, filepath.Base(t.path))})
		if err != nil {
			return err
		}
		self.res.Added = append(self.res.Added, added)
		if self.req.Cut {
			err = os.RemoveAll(self.getRealPath(t))
			if err != nil {
				return err
			}
			self.res.Removed = append(self.res.Removed, createHash(t.id, t.path))
		}
	}

	return nil
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

func copyDir(source string, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()


		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

/*/ File functions
func copyFile(src, dest string) error {
	return com.Copy(src, dest) //ToDo use it?
}
*/