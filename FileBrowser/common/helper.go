package common

import (
	"os"
	"path"
)

func CreateOrOpenFile(path string) (*os.File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Create(path)
	}
	return os.Open(path)
}

func CreateDirIfNotExists(folderName ...string) error {
	folder := path.Join(folderName...)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return os.MkdirAll(folder, 0700)
	}
	return nil
}
