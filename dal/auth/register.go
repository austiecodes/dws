package dal

import (
	"github.com/austiecodes/dws/libs/constants"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context, user *schema.User) error {
	db := resources.PGClient.WithContext(c).Table(constants.TableUsers)
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}
