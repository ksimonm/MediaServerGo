package imageProcessing

import (
	"errors"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"simon/mediaServer/metadata"
	"strconv"
	"strings"
)

func processFloatCropParam(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1.0
	}
	return f
}

func Process(filePath string, height, width, bound, quality int, format, originalFormat, cropX, cropY, cropW, cropH string) ([]byte, metadata.Metadata, error) {
	log.Printf("Process h: %v w: %v b: %v q: %v f: %v of: %v crop: %v %v %v %v\n", height, width, bound, quality, format, originalFormat, cropX, cropY, cropW, cropH)

	var inputImage *vips.ImageRef

	if cropX != "" && cropY != "" && cropW != "" && cropH != "" {
		if (strings.Contains(cropX, ".") || cropX == "0") && (strings.Contains(cropY, ".") || cropY == "0") && strings.Contains(cropW, ".") && strings.Contains(cropH, ".") {
			log.Println("CROP FLOAT")
			cropXf := processFloatCropParam(cropX)
			cropYf := processFloatCropParam(cropY)
			cropWf := processFloatCropParam(cropW)
			cropHf := processFloatCropParam(cropH)

			if cropXf == -1.0 || cropYf == -1.0 || cropWf == -1.0 || cropHf == -1.0 {
				return nil, metadata.Metadata{}, errors.New("crop params not valid")
			}

			log.Printf("crop float: %v %v %v %v", cropXf, cropYf, cropWf, cropHf)

			var err error
			inputImage, err = vips.NewImageFromFile(filePath)
			if err != nil {
				return nil, metadata.Metadata{}, err
			}

			err = inputImage.ExtractArea(int(float64(inputImage.Height())*cropXf), int(float64(inputImage.Width())*cropYf), int(float64(inputImage.Height())*cropHf), int(float64(inputImage.Width())*cropWf))
			if err != nil {
				return nil, metadata.Metadata{}, err
			}
			log.Printf("After crop h: %v w: %v", inputImage.Height(), inputImage.Width())

		} else if !strings.Contains(cropX, ".") && !strings.Contains(cropY, ".") && !strings.Contains(cropW, ".") && !strings.Contains(cropH, ".") {
			log.Println("CROP INT")
		} else {
			return nil, metadata.Metadata{}, errors.New("crop params not valid")
		}
	}

	if inputImage == nil {
		var err error
		inputImage, err = vips.NewImageFromFile(filePath)
		if err != nil {
			return nil, metadata.Metadata{}, err
		}
		log.Printf("Orig h: %v w: %v", inputImage.Height(), inputImage.Width())
	}

	if height > inputImage.Height() {
		height = inputImage.Height()
	}

	if width > inputImage.Width() {
		width = inputImage.Width()
	}

	if height != 0 && width != 0 {
		err := inputImage.ResizeWithVScale(float64(width)/float64(inputImage.Width()), float64(height)/float64(inputImage.Height()), vips.KernelAuto)
		if err != nil {
			return nil, metadata.Metadata{}, err
		}
	} else {
		scale := 1.0
		if bound != 0 {
			if inputImage.Width() > inputImage.Height() {
				scale = float64(bound) / float64(inputImage.Width())
			} else {
				scale = float64(bound) / float64(inputImage.Height())
			}
		} else if height != 0 {
			scale = float64(height) / float64(inputImage.Height())
		} else if width != 0 {
			scale = float64(width) / float64(inputImage.Width())
		}

		err := inputImage.Resize(scale, vips.KernelAuto)
		if err != nil {
			return nil, metadata.Metadata{}, err
		}
	}

	if quality == 0 {
		quality = 100
	}

	if format == "png" {
		ep := vips.NewPngExportParams()
		ep.Compression = quality

		imageBytes, _, err := inputImage.ExportPng(ep)
		if err != nil {
			return nil, metadata.Metadata{}, err
		}

		m := metadata.Metadata{Size: int64(len(imageBytes)), Height: inputImage.Height(), Width: inputImage.Width(), Format: "png"}
		return imageBytes, m, nil
	} else {
		ep := vips.NewJpegExportParams()

		ep.Quality = int(quality)
		ep.StripMetadata = true

		imageBytes, _, err := inputImage.ExportJpeg(ep)
		if err != nil {
			return nil, metadata.Metadata{}, err
		}
		m := metadata.Metadata{Size: int64(len(imageBytes)), Height: inputImage.Height(), Width: inputImage.Width(), Format: "jpg"}
		return imageBytes, m, nil
	}
}
