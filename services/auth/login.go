package auth

import (
	"fmt"

	"github.com/austiecodes/dws/db/auth"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func LoginService(c *gin.Context) error {
	session := sessions.Default(c)
	user, err := decryptUserFromForm(c)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("get user from form failed: %v", err))
		return err
	}

	fetchedUser, err := auth.FetchUser(c, user.UnixName)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("fetch user failed: %v", err))
		return err
	}

	if fetchedUser.UnixName == user.UnixName && fetchedUser.Password == user.Password {
		session.Set("unix_name", user.UnixName)
		session.Save()
		resources.Logger.Info(fmt.Sprintf("user %s logged in", user.UnixName))
		return nil
	} else {
		resources.Logger.Info(fmt.Sprintf("user %s login failed", user.UnixName))
		return fmt.Errorf("login failed")
	}
}
