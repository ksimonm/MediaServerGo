package controllers

import (
	"encoding/xml"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"simon/mediaServer/configFile"
	"simon/mediaServer/storageFactory"
	"strconv"
	"strings"
)

type S3 struct{}

func (h S3) Put(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")
	if strings.Contains(key, "//") {
		key = strings.ReplaceAll(key, "//", "/")
	}
	log.Printf("PUT %v %v uploadId: %v partNumber: %v\n", bucket, key, c.Request.URL.Query().Get("uploadId"), c.Request.URL.Query().Get("partNumber"))
	config := configFile.GetConfig()

	if c.Request.URL.Query().Get("uploadId") == "" || c.Request.URL.Query().Get("partNumber") == "" {
		storage := storageFactory.GetStorage(config)
		filePath, err := storage.PutIoReadCloser(c.Request.Body, bucket, key)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		}

		eTag, err := storageFactory.Generate(filePath, 5*1024*1024)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		}

		c.Header("etag", `"`+string(eTag)+`"`)
	} else {
		filePath := filepath.Join(config.TmpDir, "chunks", c.Request.URL.Query().Get("uploadId"), c.Request.URL.Query().Get("partNumber"))
		log.Println(filePath)

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			log.Panicln(err)
		}

		file, err := os.Create(filePath)
		if err != nil {
			log.Panicln(err)
		}
		defer file.Close()

		_, err = io.Copy(file, c.Request.Body)
		if err != nil {
			log.Panicln(err)
		}

		eTag, err := storageFactory.Generate(filePath, 5*1024*1024)
		if err != nil {
			log.Println(err)
		}

		c.Header("etag", `"`+string(eTag)+`"`)
	}
	c.String(http.StatusOK, "")
}

func (h S3) Post(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")
	config := configFile.GetConfig()
	storage := storageFactory.GetStorage(config)

	log.Printf("POST %v %v Query: %v\n", bucket, key, c.Request.URL.Query())

	if c.Request.URL.Query().Get("uploadId") == "" {
		uuidV4, err := uuid.NewRandom()
		if err != nil {
			log.Panicln(err)
		}
		log.Println(uuidV4)

		type InitiateMultipartUploadResult struct {
			Bucket   string
			Key      string
			UploadId string
		}

		xmlObj := InitiateMultipartUploadResult{Key: key, Bucket: bucket, UploadId: uuidV4.String()}

		c.XML(http.StatusOK, xmlObj)
	} else {
		body, _ := ioutil.ReadAll(c.Request.Body)
		println(string(body))

		type Part struct {
			ETag       string
			PartNumber int
		}

		type CompleteMultipartUpload struct {
			Part []Part
		}

		var completeMultipartUpload CompleteMultipartUpload
		err := xml.Unmarshal(body, &completeMultipartUpload)
		if err != nil {
			log.Println(err)
		}

		log.Println(completeMultipartUpload)

		filePath := filepath.Join(config.TmpDir, "chunks", c.Request.URL.Query().Get("uploadId"), key)
		log.Println(filePath)

		out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Panicln("failed to open outpout file:", err)
		}
		defer out.Close()

		for _, part := range completeMultipartUpload.Part {
			partPath := filepath.Join(config.TmpDir, "chunks", c.Request.URL.Query().Get("uploadId"), strconv.Itoa(part.PartNumber))
			partFile, err := os.Open(partPath)
			if err != nil {
				log.Panicln("failed to open:", err)
			}

			n, err := io.Copy(out, partFile)
			if err != nil {
				log.Panicln("failed to copy:", err)
			}
			err = partFile.Close()
			if err != nil {
				log.Panicln("failed to close", err)
			}
			log.Printf("wrote %d bytes of %s to %s\n", n, partPath, filePath)
		}

		filePathInStorage, err := storage.PutFile(filePath, bucket, key, true)
		if err != nil {
			log.Panicln("")
		}
		log.Println(filePathInStorage)

		eTag, err := storageFactory.Generate(filePathInStorage, 5*1024*1024)
		if err != nil {
			log.Println(err)
		}

		log.Println(eTag)

		type CompleteMultipartUploadResult struct {
			Bucket   string
			Key      string
			Location string
			ETag     string
		}

		xmlObj := CompleteMultipartUploadResult{
			Key:      key,
			Bucket:   bucket,
			Location: config.Url + "/" + bucket + key,
			ETag:     eTag,
		}

		c.XML(http.StatusOK, xmlObj)
	}
}

func (h S3) Delete(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")
	if strings.Contains(key, "//") {
		key = strings.ReplaceAll(key, "//", "/")
	}
	log.Printf("DELETE %v %v uploadId: %v partNumber: %v\n", bucket, key, c.Request.URL.Query().Get("uploadId"), c.Request.URL.Query().Get("partNumber"))
	config := configFile.GetConfig()

	filePath := filepath.Join(config.TmpDir, "chunks", c.Request.URL.Query().Get("uploadId"), c.Request.URL.Query().Get("partNumber"))

	e := os.RemoveAll(filePath)
	if e != nil {
		log.Println(e)
	}

	c.String(http.StatusOK, "")
}
