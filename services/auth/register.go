package services

import (
	"fmt"

	"github.com/austiecodes/dws/db/auth"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/gin-gonic/gin"
)

func RegisterService(c *gin.Context) error {
	user, err := decryptUserFromForm(c)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("get user from form failed: %v", err))
		return err
	}

	err = auth.CreateUser(c, user)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("create user failed: %v", err))
		return err
	}
	return nil
}
