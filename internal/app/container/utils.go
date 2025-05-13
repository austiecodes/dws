package container

import (
	"fmt"

	"github.com/austiecodes/dws/lib/resources"
	"github.com/gin-gonic/gin"
)

func checkContainerID(c *gin.Context, uuid, containerID string) error {
	fetchedContainer, err := FetchContainerByID(c, containerID)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to fetch container: %v", err))
		return err
	}
	if fetchedContainer.UUID != uuid {
		errMsg := fmt.Sprintf("containerID: %v does not match fetched containerID: %v", containerID, fetchedContainer.ID)
		resources.Logger.Error(errMsg)
		return fmt.Errorf("%s", errMsg)
	}
	return nil
}
