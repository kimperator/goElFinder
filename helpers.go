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
	if i := len(s) % 4; i != 0 {
		s += strings.Repeat("=", 4-i)
	}
	t, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(t), nil
}

func encode64(s string) string {
	t:= base64.StdEncoding.EncodeToString([]byte(s))
	return strings.TrimRight(t, "=")
}

func createHash(volumeId, path string) string {
	return volumeId + "_" + encode64(path)
}

func parseHash(target string) (volumeId, path string, err error) {
	splitTarget := strings.Split(target, "_")
	if len(splitTarget) != 2 {
		return "", "", errors.New("Bad target")
	}
	volumeId = splitTarget[0]
	path, err = decode64(splitTarget[1])
	if len(splitTarget) != 2 {
		return "", "", errors.New("Bad base64 path")
	}
	path = strings.TrimPrefix(filepath.Clean(path), "..")
	path = strings.TrimPrefix(filepath.Clean(path), string(filepath.Separator) + "..")
	if path == "" {
		path = "/"
	}
	fmt.Println("Clean path:", path)
	return volumeId, path, err
}

func getImageDim(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%dx%d", image.Width, image.Height), nil
}