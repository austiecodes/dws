package auth

import (
	"fmt"

	"github.com/austiecodes/dws/lib/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-contrib/sessions"
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
		UUID:     uuid.New().String(),
		UserName: userName,
		UnixName: unixName,
		UserPswd: password,
	}
	err = CreateUser(c, user)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("create user failed: %v", err))
		return err
	}
	return nil
}

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

	fetchedUser, err := FetchUser(c, uuid)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("fetch user failed: %v", err))
		return err
	}

	if fetchedUser.UserName == userName && fetchedUser.UserPswd == password {
		session.Set("uuid", uuid)
		session.Save()
		resources.Logger.Info(fmt.Sprintf("user %s logged in", userName))
		return nil
	} else {
		resources.Logger.Info(fmt.Sprintf("user %s login failed", userName))
		return fmt.Errorf("login failed")
	}
}
