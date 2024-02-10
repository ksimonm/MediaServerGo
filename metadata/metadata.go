package metadata

import (
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"os"
	"strings"
)

type Metadata struct {
	Size     int64  `json:"size,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Format   string `json:"format,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

func Get(filePath string, mtype string) Metadata {
	file, err := os.Open(filePath)
	if err != nil {
		return Metadata{}
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return Metadata{}
	}

	if strings.HasPrefix(mtype, "image") {
		image, err := vips.NewImageFromFile(filePath)
		if err != nil {
			log.Println(err)
			return Metadata{}
		}
		return Metadata{Size: stat.Size(), Width: image.Width(), Height: image.Height(), Format: strings.ReplaceAll(mtype, "image/", "")}
	}

	return Metadata{Size: stat.Size()}
}
