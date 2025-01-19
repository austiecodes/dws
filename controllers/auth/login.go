package auth

import (
	"net/http"

	services "github.com/austiecodes/dws/services/auth"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	err := services.LoginService(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "login"})
}
