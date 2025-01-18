package services

import (
	"github.com/austiecodes/dws/resources"
	"github.com/docker/docker/api/types/image"
	"github.com/gin-gonic/gin"
)

func ListImages(c *gin.Context) ([]image.Summary, error) {
	images, err := resources.DockerClient.ImageList(c, image.ListOptions{})
	return images, err
}
