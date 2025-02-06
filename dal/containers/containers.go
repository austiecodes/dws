package dal

import (
	"fmt"
	"time"

	"github.com/austiecodes/dws/libs/constants"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/austiecodes/dws/models/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

func FetchContainerByID(c *gin.Context, containerID string) (*schema.Container, error) {
	var container schema.Container
	db := resources.PGClient.WithContext(c).Table(constants.TableContainers)
	err := db.First(&container, "container_id = ?", containerID).Error
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to fetch container: %v", err))
		return nil, err
	}
	return &container, nil
}

func FetchContainerByUUID(c *gin.Context, uuid string) ([]*schema.Container, error) {
	var containerList []*schema.Container
	db := resources.PGClient.WithContext(c).Table(constants.TableContainers)
	err := db.Where("uuid = ?", uuid).Find(&containerList).Error
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to fetch container: %v", err))
		return nil, err
	}
	return containerList, nil
}

func CreateContainer(c *gin.Context, uuid string, config *types.CreateContainerOptions) (*container.CreateResponse, error) {
	tx := resources.PGClient.WithContext(c).Begin()
	if tx.Error != nil {
		resources.Logger.Error(fmt.Sprintf("failed to begin transaction: %v", tx.Error))
		return nil, tx.Error
	}
	resp, err := resources.DockerClient.ContainerCreate(c, config.ContainerConfig, config.HostConfig, config.NetworkingConfig, config.Platform, config.ContainerName)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to create container: %v", err))
		tx.Rollback()
		RemoveContainerByID(c, resp.ID)
		return nil, err
	}
	if err := tx.Create(&schema.Container{
		UUID:        uuid,
		ContainerID: resp.ID,
		Name:        config.ContainerConfig.Hostname,
		Image:       config.ContainerConfig.Image,
	}).Error; err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to create container record: %v", err))
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &resp, nil
}

func RemoveContainerByID(c *gin.Context, containerID string) error {
	tx := resources.PGClient.WithContext(c).Begin()
	if tx.Error != nil {
		resources.Logger.Error(fmt.Sprintf("failed to begin transaction: %v", tx.Error))
		return tx.Error
	}

	err := resources.DockerClient.ContainerRemove(c, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to remove container: %v", err))
		tx.Rollback()
		return err
	}

	var container schema.Container
	err = tx.Table(constants.TableContainers).Delete(&container, "container_id = ?", containerID).Error
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to remove container record: %v", err))
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to commit transaction: %v", err))
		return err
	}

	return nil
}

func CommitContainerAsImage(c *gin.Context, uuid, userName, containerID, comment string) (string, error) {
	commitResp, err := resources.DockerClient.ContainerCommit(c, containerID, container.CommitOptions{
		Reference: fmt.Sprintf("user-task-%d", time.Now().Unix()), // 生成唯一镜像标签
		Comment:   comment,
		Author:    userName,
		Changes:   []string{"ENV TASK_MODE=production"},
		Pause:     true,
	})
	if err != nil {
		return "", fmt.Errorf("commit failed: %v", err)
	}
	return commitResp.ID, nil
}
