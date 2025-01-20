package containers

import (
	"fmt"

	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func FetchContainerByID(c *gin.Context, containerID string) (*schema.Container, error) {
	var container schema.Container
	db := resources.PGClient.WithContext(c).Table("containers")
	err := db.First(&container, "container_id = ?", containerID).Error
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to fetch container: %v", err))
		return nil, err
	}
	return &container, nil
}
