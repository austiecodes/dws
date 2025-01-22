package dal

import (
	"errors"

	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context, user *schema.User) error {
	db := resources.PGClient.WithContext(c).Table("users")
	if err := db.First(&user, "uuid = ?", user.UUID).Error; err == nil {
		return errors.New("user already exists")
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}
