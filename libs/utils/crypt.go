package utils

import (
	"encoding/base64"

	"github.com/forgoer/openssl"
)

func Encrypt(plainText []byte, key []byte) (string, error) {
	cipherText, err := openssl.AesCBCEncrypt(plainText, key, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(cipherText string, key []byte) (string, error) {
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
