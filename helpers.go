package goElFinder

import (
	"encoding/base64"
	"strings"
	"errors"
	"path/filepath"
	"os"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
)

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

func parseHash(config Config, target string) (volume response, path string, err error) { //ToDo check file name
	var volumeId string
	splitTarget := strings.SplitN(target, "_", 2)
	if len(splitTarget) != 2 {
		return volume, path, errors.New("Bad target")
	}
	volumeId = splitTarget[0]
	path, err = decode64(splitTarget[1])
	if len(splitTarget) != 2 {
		return volume, path, errors.New("Bad base64 path")
	}
	path = strings.TrimPrefix(filepath.Clean(path), "..")
	path = strings.TrimPrefix(filepath.Clean(path), string(os.PathSeparator) + "..")
	if path == "" {
		path = string(os.PathSeparator)
	}

	if _, ok := config[volumeId]; !ok {
		return volume, path, errors.New("Bad volume id")
	}

	//var volume response
	volume.config.id = volumeId
	volume.setRoot(config[volumeId].Root)
	volume.setDefaultRight(config[volumeId].DefaultRight)
	volume.allowDirs(config[volumeId].AllowDirs)
	volume.denyDirs(config[volumeId].DenyDirs)

	return volume, path, err
}

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