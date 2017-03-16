package goElFinder

import (
	"encoding/base64"
	"strings"
	"path/filepath"
	"os"
	"fmt"
//	"github.com/Unknwon/com"
	"image"
	"github.com/disintegration/imaging"
	"github.com/disintegration/gift"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
	"path"
	"image/color"
	"strconv"
	"archive/zip"
	"io"
)

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


/*/ File functions
func copyFile(src, dest string) error {
	return com.Copy(src, dest) //ToDo use it?
}
*/

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


// Image functions
func getImageDim(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%dx%d", img.Width, img.Height), nil
}

func resizeImage(imagePath, imageTarget string, width, height int) error {
	src, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}
	dst := imaging.Resize(src, width, height, imaging.Box)
	os.MkdirAll(path.Dir(imageTarget), os.ModePerm)
	return imaging.Save(dst, imageTarget)
}

func cropImage(imagePath, imageTarget string, x, y, width, height int) error {
	src, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}
	rectangle := image.Rect(x, y, x+width, y+height)
	dst := imaging.Crop(src, rectangle)
	os.MkdirAll(path.Dir(imageTarget), os.ModePerm)
	return imaging.Save(dst, imageTarget)
}

func rotateImage(imagePath, imageTarget, bg string, degree int) error {
	degree = 360 - degree
	src, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}

	switch degree {
	case 90:
		dst := imaging.Rotate90(src)
		err = imaging.Save(dst, imageTarget)

	case 180:
		dst := imaging.Rotate180(src)
		err = imaging.Save(dst, imageTarget)

	case 270:
		dst := imaging.Rotate270(src)
		err = imaging.Save(dst, imageTarget)
	default:
		r, g, b, a := hexColor(bg)
		gf := gift.New(gift.Rotate(float32(degree), color.NRGBA{R:r,G:g,B:b,A:a}, gift.LinearInterpolation))
		dst := image.NewRGBA(gf.Bounds(src.Bounds()))
		gf.Draw(dst, src)
		err = imaging.Save(dst, imageTarget)
	}

	os.MkdirAll(path.Dir(imageTarget), os.ModePerm)
	return err
}

func hexColor(s string) (r, g, b, a uint8) {
	var c string
	if len(s) < 3 {
		c = "0xffffffff"
	} else if s[0] == 35 {
		c = "0x" + s[1:]
	} else if s[0:1] != "0x" {
		c = "0x" + s
	}
	for len(c) <= 9 {
		c = c + "f"
	}
	decimal, _ := strconv.ParseUint(c, 0, 32)
	a = uint8(decimal & 0xFF)
	b = uint8((decimal >> 8) & 0xFF)
	g = uint8((decimal >> 16) & 0xFF)
	r = uint8((decimal >> 24) & 0xFF)

	return
}

// Archive functions

func unpackZip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}


func packZip(source, target string) error {
	//zipfile, err := os.Create(target)
	zipfile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

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
	})

	return err
}