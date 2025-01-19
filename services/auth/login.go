package services

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/austiecodes/dws/libs/resources"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var aesKey = []byte("1234567890123456")

func encrypt(plainText, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(plainText)%aes.BlockSize != 0 {
		return "", errors.New("plaintext is not a multiple of the block size")
	}

	cipherText := make([]byte, len(plainText))
	iv := key[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plainText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decrypt(cipherText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedCipherText, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	if len(decodedCipherText)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	iv := key[:aes.BlockSize]
	plainText := make([]byte, len(decodedCipherText))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, decodedCipherText)

	return string(plainText), nil
}

// mocked data
var users = map[string]string{
	"admin": "password123",
}

func LoginService(c *gin.Context) error {
	session := sessions.Default(c)
	encryptedUsername := c.PostForm("username")
	encryptedPassword := c.PostForm("password")

	username, err := decrypt(encryptedUsername, aesKey)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt username failed: %v", err))
		return err
	}

	password, err := decrypt(encryptedPassword, aesKey)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("decrypt password failed: %v", err))
		return err
	}

	storedPassword, ok := users[username]
	if !ok || storedPassword != password {
		errMsg := fmt.Sprintf("invalid username or password for user: %s", username)
		resources.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	session.Set("user", username)
	session.Save()
	return nil
}
