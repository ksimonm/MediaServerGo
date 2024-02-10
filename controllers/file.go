package controllers

import (
	"bytes"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"math"
	"net/http"
	"simon/mediaServer/configFile"
	"simon/mediaServer/imageProcessing"
	"simon/mediaServer/metadata"
	"simon/mediaServer/storageFactory"
	"strconv"
)

type File struct{}

type jsonResponse struct {
	Err string `json:"err"`
}

func processIntParam(shortParam string, longParam string, defaultValue int, min int, max int) int {
	res := defaultValue

	if longParam != "" {
		shortParam = longParam
	}
	if shortParam != "" {
		res64, err := strconv.ParseInt(shortParam, 10, 32)
		if err != nil {
			log.Println(err)
			return defaultValue
		}
		res = int(res64)

		if res < min {
			res = min
		}

		if res > max {
			res = max
		}
	}

	return res
}

func processStringParam(shortParam string, longParam string) string {
	if longParam != "" {
		shortParam = longParam
	}
	return shortParam
}

func (h File) Get(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")
	log.Printf("GET %v %v queryParams: %v\n", bucket, key, c.Request.URL.Query())

	config := configFile.GetConfig()
	storage := storageFactory.GetStorage(config)

	if storage.Exists(bucket, key) {
		filePath, err := storage.Get(bucket, key)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		}

		mtype, err := mimetype.DetectFile(filePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, jsonResponse{Err: "bad file"})
			return
		}

		if len(c.Request.URL.Query()) == 0 || (len(c.Request.URL.Query()) == 1 && c.Request.URL.Query().Get("metadataOnly") == "true") {
			// return original
			log.Println(c.Request.URL.Query().Get("metadataOnly"))
			if c.Request.URL.Query().Get("metadataOnly") == "true" {
				meta := metadata.Get(filePath, mtype.String())
				c.JSON(http.StatusOK, meta)
			} else {
				c.File(filePath)
			}
		} else {
			queryParams := c.Request.URL.Query()

			width := processIntParam(queryParams.Get("w"), queryParams.Get("width"), 0, 1, math.MaxInt32)
			height := processIntParam(queryParams.Get("h"), queryParams.Get("height"), 0, 1, math.MaxInt32)
			bound := processIntParam(queryParams.Get("b"), queryParams.Get("bound"), 0, 1, math.MaxInt32)
			quality := processIntParam(queryParams.Get("q"), queryParams.Get("quality"), 100, 1, 100)

			format := processStringParam(queryParams.Get("f"), queryParams.Get("format"))
			if format != "" && !config.InAllowedFormats(format) {
				msg := jsonResponse{Err: format + " not allowed"}
				c.JSON(http.StatusInternalServerError, msg)
				return
			}

			preset := processStringParam(queryParams.Get("p"), queryParams.Get("preset"))
			if preset != "" && config.PresetExists(bucket, preset) {
				presetObj := config.GetPreset(bucket, preset)
				if presetObj.H != 0 {
					height = presetObj.H
				}
				if presetObj.W != 0 {
					width = presetObj.W
				}
				if presetObj.B != 0 {
					bound = presetObj.B
				}
				if presetObj.Q != 0 {
					quality = presetObj.Q
				}
			} else if preset != "" {
				msg := jsonResponse{Err: preset + " not exists"}
				c.JSON(http.StatusInternalServerError, msg)
				return
			}
			log.Printf("h: %v w: %v b: %v q: %v f: %v", height, width, bound, quality, format)

			processedImage, imageMetadata, err := imageProcessing.Process(filePath, height, width, bound, quality, format, mtype.String(), queryParams.Get("crop.x"), queryParams.Get("crop.y"), queryParams.Get("crop.w"), queryParams.Get("crop.h"))
			if err != nil {
				log.Println(err)
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			if c.Request.URL.Query().Get("metadataOnly") == "true" {
				c.JSON(http.StatusOK, imageMetadata)
				return
			}

			_, err = io.Copy(c.Writer, bytes.NewReader(processedImage))
			if err != nil {
				log.Println(err)
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		}
	} else {
		c.String(http.StatusNotFound, "")
	}
}
