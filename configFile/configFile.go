package configFile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Preset struct {
	Name string `json:"name"`
	W    int    `json:"w"`
	H    int    `json:"h"`
	B    int    `json:"b"`
	Q    int    `json:"q"`
}

type Bucket struct {
	Name            string   `json:"name"`
	AccessKeyId     string   `json:"accessKeyId"`
	SecretAccessKey string   `json:"secretAccessKey"`
	Presets         []Preset `json:"presets"`
}

type Config struct {
	Env            string   `json:"env"`
	Port           int      `json:"port"`
	Url            string   `json:"url"`
	TmpDir         string   `json:"tmpDir"`
	Buckets        []Bucket `json:"buckets"`
	AllowedFormats []string `json:"allowedFormats"`
	Cache          struct {
		CacheRoot   string `json:"cacheRoot"`
		MaxSizeInMB int32  `json:"maxSizeInMB"`
	} `json:"cache"`
	Storage struct {
		Type        string `json:"type"`
		StorageRoot string `json:"storageRoot"`
	} `json:"storage"`
}

var singleton Config
var once sync.Once

func GetConfig() Config {
	once.Do(func() {
		jsonFile, err := os.Open("config/config.json")
		if err != nil {
			log.Panicln(err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		err = json.Unmarshal(byteValue, &singleton)
		if err != nil {
			log.Panicln(err)
		}
	})
	return singleton
}

func (c Config) InAllowedFormats(format string) bool {
	for _, i := range c.AllowedFormats {
		if i == format {
			return true
		}
	}
	return false
}

func (c Config) existsBucket(bucketName string) bool {
	for _, i := range c.Buckets {
		if i.Name == bucketName {
			return true
		}
	}
	return false
}

func (c Config) getBucket(bucketName string) Bucket {
	for _, i := range c.Buckets {
		if i.Name == bucketName {
			return i
		}
	}
	return Bucket{}
}

func (c Config) PresetExists(bucketName string, presetName string) bool {
	if c.existsBucket(bucketName) {
		bucket := c.getBucket(bucketName)
		for _, i := range bucket.Presets {
			if i.Name == presetName {
				return true
			}
		}
	}
	return false
}

func (c Config) GetPreset(bucketName string, presetName string) Preset {
	if c.PresetExists(bucketName, presetName) {
		bucket := c.getBucket(bucketName)
		for _, i := range bucket.Presets {
			if i.Name == presetName {
				return i
			}
		}
	}
	return Preset{}
}
