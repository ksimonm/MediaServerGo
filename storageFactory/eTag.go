package storageFactory

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strconv"
)

func Generate(filePath string, chunkSize int64) (string, error) {
	var md5Array [][16]byte

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	nBytes, nChunks := int64(0), int64(0)
	r := bufio.NewReader(f)
	buf := make([]byte, 0, chunkSize)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		nChunks++
		nBytes += int64(len(buf))
		md5Array = append(md5Array, md5.Sum(buf))

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}
	//log.Println(md5Array)

	eTag := ""

	if len(md5Array) == 1 {
		eTag = hex.EncodeToString(md5Array[0][:])
	} else {
		s := ""
		for _, m := range md5Array {
			s = string(m[:])
		}
		eTag = hex.EncodeToString([]byte(s)) + "-" + strconv.Itoa(len(md5Array))
	}

	return eTag, nil
}
