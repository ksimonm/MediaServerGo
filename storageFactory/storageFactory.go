package storageFactory

import (
	"io"
	"simon/mediaServer/configFile"
	"sync"
)

type Storage interface {
	Exists(bucket string, key string) bool
	PutFile(filePath string, bucket string, key string, removeOriginal bool) (string, error)
	PutIoReadCloser(body io.ReadCloser, bucket string, key string) (string, error)
	Get(bucket string, key string) (string, error)
	Delete(bucket string, key string) error
}

var singleton Storage
var once sync.Once

func GetStorage(config configFile.Config) Storage {
	once.Do(func() {
		if config.Storage.Type == "local" {
			singleton = newLocalStorage(config)
		}
	})
	return singleton
}
