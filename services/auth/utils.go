package auth

import (
	"fmt"

	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/libs/utils"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func decryptUserFromForm(c *gin.Context) (schema.User, error) {
	encryptedUnixName := c.PostForm("unix_name")
	encryptedPassword := c.PostForm("password")

	unixName, err := utils.Decrypt(encryptedUnixName, []byte(resources.AESKey))
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt unix_name failed: %v", err))
		return schema.User{}, err
	}

	password, err := utils.Decrypt(encryptedPassword, []byte(resources.AESKey))
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt password failed: %v", err))
		return schema.User{}, err
	}

	return schema.User{
		UnixName: unixName,
		Password: password,
	}, nil
}
