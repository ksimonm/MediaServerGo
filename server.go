package main

import (
	"fmt"
	"log"
	"simon/mediaServer/configFile"
	"simon/mediaServer/controllers"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := configFile.GetConfig()
	log.Println(config)

	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	vips.Startup(nil)
	defer vips.Shutdown()

	r := gin.Default()

	s3 := new(controllers.S3)
	file := new(controllers.File)
	tiles := new(controllers.Tiles)

	r.POST("/:bucket/*key", s3.Post)
	r.PUT("/:bucket/*key", s3.Put)
	r.DELETE("/:bucket/*key", s3.Delete)
	r.GET("/:bucket/*key", file.Get)
	r.GET("/tiles/1", tiles.Tiles1)

	fmt.Println("START GIN on port:", config.Port)
	r.Run(":" + strconv.Itoa(config.Port))
}
