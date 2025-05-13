package auth

import (
	"fmt"

	"github.com/austiecodes/dws/lib/resources"
	"github.com/austiecodes/dws/lib/utils"
	"github.com/gin-gonic/gin"
)

func decryptUserFromForm(c *gin.Context, field string) (string, error) {
	encryptedField := c.PostForm(field)
	decryptedField, err := utils.Decrypt(encryptedField, []byte(resources.AESKey))
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt %s failed: %v", field, err))
		return "", err
	}
	return decryptedField, nil
}
