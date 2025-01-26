package services

import (
	"fmt"

	dal "github.com/austiecodes/dws/dal/auth"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func LoginService(c *gin.Context) error {
	session := sessions.Default(c)
	uuid, err := decryptUserFromForm(c, "uuid")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("get user from form failed: %v", err))
		return err
	}
	userName, err := decryptUserFromForm(c, "userName")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("get unix name from form failed: %v", err))
		return err
	}
	password, err := decryptUserFromForm(c, "password")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("get password from form failed: %v", err))
		return err
	}

	fetchedUser, err := dal.FetchUser(c, uuid)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("fetch user failed: %v", err))
		return err
	}

	if fetchedUser.UserName == userName && fetchedUser.Password == password {
		session.Set("uuid", uuid)
		session.Save()
		resources.Logger.Info(fmt.Sprintf("user %s logged in", userName))
		return nil
	} else {
		resources.Logger.Info(fmt.Sprintf("user %s login failed", userName))
		return fmt.Errorf("login failed")
	}
}
