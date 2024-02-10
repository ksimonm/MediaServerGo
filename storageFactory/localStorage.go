package storageFactory

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"simon/mediaServer/configFile"
)

type localStorage struct {
	storageRoot string
}

func (s localStorage) Exists(bucket string, key string) bool {
	if _, err := os.Stat(filepath.Join(s.storageRoot, bucket, key)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s localStorage) PutFile(originalFilePath string, bucket string, key string, removeOriginal bool) (string, error) {
	filePathInStorage := filepath.Join(s.storageRoot, bucket, key)
	if removeOriginal {
		err := os.Rename(originalFilePath, filePathInStorage)
		if err != nil {
			return "", err
		}
	} else {
		source, err := os.Open(originalFilePath)
		if err != nil {
			return "", err
		}
		defer source.Close()

		destination, err := os.Create(filePathInStorage)
		if err != nil {
			return "", err
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			return "", err
		}
	}
	return filePathInStorage, nil
}

func (s localStorage) PutIoReadCloser(body io.ReadCloser, bucket string, key string) (string, error) {
	filePath := filepath.Join(s.storageRoot, bucket, key)
	log.Println(filePath)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (s localStorage) Get(bucket string, key string) (string, error) {
	if s.Exists(bucket, key) {
		return filepath.Join(s.storageRoot, bucket, key), nil
	} else {
		return "", errors.New("file not exists")
	}
}

func (s localStorage) Delete(bucket string, key string) error {
	return nil
}

func newLocalStorage(config configFile.Config) Storage {
	return localStorage{storageRoot: config.Storage.StorageRoot}
}
