package images

import (
	"net/http"

	services "github.com/austiecodes/dws/services/images"
	"github.com/gin-gonic/gin"
)

func ListImages(c *gin.Context) {
	images, err := services.ListImages(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, images)
}
