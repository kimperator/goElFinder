package elFinder

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
//	str = strings.Replace(str, "_", "/", -1)

	fmt.Println("decode:", str)
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

func parseHash(target string) (volumeId, path string, err error) { //ToDo check file name
	splitTarget := strings.SplitN(target, "_", 2)
	fmt.Println(splitTarget)
	if len(splitTarget) != 2 {
		return "", "", errors.New("Bad target")
	}
	volumeId = splitTarget[0]
	path, err = decode64(splitTarget[1])
	if len(splitTarget) != 2 {
		return "", "", errors.New("Bad base64 path")
	}
	path = strings.TrimPrefix(filepath.Clean(path), "..")
	path = strings.TrimPrefix(filepath.Clean(path), string(os.PathSeparator) + "..")
	if path == "" {
		path = string(os.PathSeparator)
	}
	fmt.Println("Clean path:", path)
	return volumeId, path, err
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