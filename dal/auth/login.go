package dal

import (
	"github.com/austiecodes/dws/libs/constants"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func FetchUser(c *gin.Context, uuid string) (*schema.User, error) {
	db := resources.PGClient.WithContext(c).Table(constants.TableUsers)
	var user schema.User
	if err := db.First(&user, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
