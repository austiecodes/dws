package services

import (
	"encoding/base64"
	"fmt"

	"github.com/austiecodes/dws/db/auth"
	"github.com/austiecodes/dws/libs/resources"
	"github.com/forgoer/openssl"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AES密钥 (16, 24, 32字节长度)
var aesKey = []byte("1234567890123456")

// 加密函数
func encrypt(plainText []byte, key []byte) (string, error) {
	cipherText, err := openssl.AesCBCEncrypt(plainText, key, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// 解密函数
func decrypt(cipherText string, key []byte) (string, error) {
	decodedCipherText, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	plainText, err := openssl.AesCBCDecrypt(decodedCipherText, key, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func LoginService(c *gin.Context) error {
	session := sessions.Default(c)

	encryptedUUID := c.PostForm("uuid")
	encryptedUnixName := c.PostForm("unix_name")
	encryptedPassword := c.PostForm("password")

	unixName, err := decrypt(encryptedUnixName, aesKey)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt unix_name failed: %v", err))
		return err
	}
	uuid, err := decrypt(encryptedUUID, aesKey)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt uuid failed: %v", err))
		return err
	}
	password, err := decrypt(encryptedPassword, aesKey)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt password failed: %v", err))
		return err
	}

	user, err := auth.FetchUser(c, uuid)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("fetch user failed: %v", err))
		return err
	}

	if user.Password == password {
		session.Set("uuid", uuid)
		session.Save()
		resources.Logger.Info(fmt.Sprintf("user %s logged in", unixName))
		return nil
	} else {
		resources.Logger.Info(fmt.Sprintf("user %s login failed", unixName))
		return fmt.Errorf("login failed")
	}
}
