package services

import (
	"fmt"

	dal "github.com/austiecodes/dws/dal/auth"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterService(c *gin.Context) error {
	userName, err := decryptUserFromForm(c, "user_name")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt user from form failed: %v", err))
		return err
	}
	unixName, err := decryptUserFromForm(c, "unix_name")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt unix name from form failed: %v", err))
		return err
	}
	password, err := decryptUserFromForm(c, "password")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt password from form failed: %v", err))
		return err
	}

	user := &schema.User{
		UUID:     uuid.New(),
		UserName: userName,
		UnixName: unixName,
		Password: password,
	}
	err = dal.CreateUser(c, user)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("create user failed: %v", err))
		return err
	}
	return nil
}
