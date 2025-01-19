package auth

import (
	"errors"

	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context, user schema.User) error {
	db := resources.PGClient.WithContext(c).Table("users")
	if err := db.First(&user, "unix_name = ?", user.UnixName).Error; err == nil {
		return errors.New("user already exists")
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}
