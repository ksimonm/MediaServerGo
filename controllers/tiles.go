package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Tiles struct{}

func (h Tiles) Tiles1(c *gin.Context) {
	html := `
		<div id="openseadragon1" style="width: 800px; height: 600px;"></div>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/openseadragon/2.4.2/openseadragon.min.js"></script>
		<script type="text/javascript">
			const viewer = OpenSeadragon({
				id: "openseadragon1",
				tileSources: '/artefact/slon_20mb.jpg.dzi'
			});
		</script>
	`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
