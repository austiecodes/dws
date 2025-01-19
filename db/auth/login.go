package auth

import (
	"github.com/austiecodes/dws/libs/resources"
	"github.com/austiecodes/dws/models/schema"
	"github.com/gin-gonic/gin"
)

func FetchUser(c *gin.Context, unix_name string) (*schema.User, error) {
	db := resources.PGClient.WithContext(c).Table("users")
	var user schema.User
	if err := db.First(&user, "unix_name = ?", unix_name).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
