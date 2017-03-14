package goElFinder

import (
	"io"
	"path/filepath"
	"os"
	"strconv"
	"strings"
	"fmt"
	"regexp"
	"errors"
)

func (self *response) upload(id, path, name string, file io.Reader) error {
	f, err := os.OpenFile(filepath.Join(conf[id].Root, path, name), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		self.Warning = append(self.Warning, err.Error())
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		self.Warning = append(self.Warning, err.Error())
		return err
	}
	i, err := _infoFileDir(id, path)
	if err != nil {
		self.Warning = append(self.Warning, err.Error())
	}
	self.Added = append(self.Added, i)
	return nil
}

func (self *response) chunkUpload(cid int, id, path, chunk string, file io.Reader) error {
	if file != nil {
		tmpPath := filepath.Join(conf[id].Root, path, fmt.Sprintf(".%d_%s~", cid, chunk))
		f, err := os.OpenFile(tmpPath, os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			self.Warning = append(self.Warning, err.Error())
			return err
		}
		_, err = io.Copy(f, file)
		if err != nil {
			self.Warning = append(self.Warning, err.Error())
			return err
		}
		f.Close()
		os.Rename(tmpPath, filepath.Join(conf[id].Root, path, fmt.Sprintf(".%d_%s", cid, chunk)))
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
			if _, err := os.Stat(filepath.Join(conf[id].Root, path, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, i, total))); os.IsNotExist(err) {
				return false
			}
		}
		return true
	}
	complete := allComplete()
	// -----------------------------------------------------------------

	if complete {
		self.Chunkmerged = fmt.Sprintf(".%d_%s.%d_part", cid, name, total)
		self.Name = name
	}
	fmt.Println("Check chunk result:", complete)
	if self.Added == nil {
		self.Added = []fileDir{}
	}
	return nil
}

func (self *response) chunkMerge(id, path, chunk string) error {
	var err error
	re := regexp.MustCompile(`(\.[0-9][0-9]*?)(_.*?)(\.[0-9][0-9]*?_part)`)
	ch := re.FindStringSubmatch(chunk)
	if len(ch) != 4 {
		return errors.New("Bad merged chunk name format")
	}

	cid, err := strconv.Atoi(ch[1][1:])
	if err != nil {
		return err
	}
	name := ch[2][1:]
	total, err :=  strconv.Atoi(strings.TrimRight(ch[3][1:], "_part"))
	if err != nil {
		return err
	}

	targetPath := filepath.Join(conf[id].Root, path, name)
	os.Rename(filepath.Join(conf[id].Root, path, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, 0, total)), targetPath)
	f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		self.Warning = append(self.Warning, err.Error())
		return err
	}
	defer f.Close()
	for i := 1; i <= total; i++ {
		chunkPath := filepath.Join(conf[id].Root, path, fmt.Sprintf(".%d_%s.%d_%d.part", cid, name, i, total))
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
		c.Close()
		err = os.Remove(chunkPath)
		if err != nil {
			return err
		}
	}
	fInfo, err := _infoFileDir(id, path)
	if err != nil {
		return err
	}
	self.Added = append(self.Added, fInfo)

	return nil
}
