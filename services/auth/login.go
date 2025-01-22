package services

import (
	"fmt"

	dal "github.com/austiecodes/dws/dal/auth"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
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
	unixName, err := decryptUserFromForm(c, "unix_name")
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("get unix name from form failed: %v", err))
		return err
	}

	user := &schema.User{
		UUID:     uuid,
		UnixName: unixName,
	}
	fetchedUser, err := dal.FetchUser(c, user.UUID)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("fetch user failed: %v", err))
		return err
	}

	if fetchedUser.UnixName == user.UnixName && fetchedUser.Password == user.Password {
		session.Set("uuid", user.UUID)
		session.Save()
		resources.Logger.Info(fmt.Sprintf("user %s logged in", user.UnixName))
		return nil
	} else {
		resources.Logger.Info(fmt.Sprintf("user %s login failed", user.UnixName))
		return fmt.Errorf("login failed")
	}
}
