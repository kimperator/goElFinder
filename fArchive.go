package goElFinder

import (
	"os"
	"path/filepath"
	"strings"
	"archive/zip"
	"io"
	"fmt"
	"golang.org/x/text/encoding/charmap"
)

func (self *elf) archive() error {
	arch := filepath.Join(self.realTargetPath(self.target), filepath.Clean(self.req.Name))
	var sources []string
	for i := range self.req.Targets {
		source, err := self.realPath(self.req.Targets[i])
		if err != nil {
			return err
		}
		sources = append(sources, source)
	}

	err := packZip(arch, sources...)
	if err != nil {
		return err
	}

	tgt, err := self.volumes.parsePathHash(self.req.Target)
	if err != nil {
		fmt.Println(err)
		return err
	}
	info, err := self.volumes.infoTarget(target{id: tgt.id, path: filepath.Join(tgt.path, filepath.Clean(self.req.Name))})
	if err != nil {
		fmt.Println(err)
		return err
	}
	self.res.Added = append(self.res.Added, info)

	return nil
}

func (self *elf) extract() error {
	var (
		dir, to string
		added []string
	)

	realPath := self.realTargetPath(self.target)
	if self.req.MakeDir {
		dir = strings.TrimSuffix(filepath.Base(realPath), filepath.Ext(realPath))
		added = []string{string(filepath.Separator)}
	}
	to = filepath.Join(filepath.Dir(realPath), dir)

fmt.Printf("target: `%s` makedir: `%v` dir:`%s` to: `%s` added: `%v`\n", self.target, self.req.MakeDir, dir, to, added)


	extracted, err := unpackZip(self.realTargetPath(self.target), to)
	if err != nil {
		return err
	}
	added = append(added, extracted...)

	tgt, err := self.volumes.parsePathHash(self.req.Target)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tgtTo := filepath.Join(string(filepath.Separator), strings.TrimLeft(to, self.volumes[tgt.id].Root))
	for f := range added {
		t := target{
			id: tgt.id,
			path: filepath.Join(tgtTo, added[f]),
		}
fmt.Printf("extract: `%s` - `%+v`\n", added[f], t)

		info, err := self.volumes.infoTarget(t)
		if err != nil {
			fmt.Println(err)
			return err
		}
		self.res.Added = append(self.res.Added, info)
	}


	return nil
}

func unpackZip(archive, target string) (extracted []string, err error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return extracted, err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return extracted, err
	}

	for _, file := range reader.File {
fmt.Printf(" --- extract: `%s` ", file.Name)
		name, err := charmap.CodePage866.NewDecoder().String(file.Name)
		if err != nil {
			return extracted, err
		}

		extracted = append(extracted, name)

fmt.Printf(" name: `%s`\n", name)
		path := filepath.Join(target, name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return extracted, err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return extracted, err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return extracted, err
		}
	}

	return extracted, nil
}


func packZip(target string, sources ...string) error {
	//zipfile, err := os.Create(target)
	zipfile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			continue
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(source)
		}

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasPrefix(info.Name(), ".") {

				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}

				if baseDir != "" {
					header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
				}

				if info.IsDir() {
					header.Name += "/"
				} else {
					header.Method = zip.Deflate
				}

				header.Name, err = charmap.CodePage866.NewEncoder().String(header.Name)
				if err != nil {
					return err
				}

				writer, err := archive.CreateHeader(header)
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(writer, file)
				return err
			} else {
				if info.IsDir() {
					return filepath.SkipDir
				}
			}
			return nil
		})
	}

	archive.Flush()
	return nil
}


func decodeCP866(str string) string {
	dec := charmap.CodePage866.NewDecoder()
	out, _ := dec.Bytes([]byte(str))
	return string(out)
}

func encodeCP866(str string) string {
	enc := charmap.CodePage866.NewEncoder()
	out, _ := enc.String(str)
	return out
}