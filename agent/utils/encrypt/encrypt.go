package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// getEncryptionKey 从环境变量获取加密密钥
func getEncryptionKey() []byte {
	// 优先从环境变量读取
	if key := os.Getenv("GPANEL_ENCRYPTION_KEY"); key != "" {
		return []byte(key)
	}

	// 如果没有设置环境变量，使用默认密钥（仅用于开发环境）
	// 生产环境必须设置环境变量
	return []byte("gpanel-agent-encrypt-key-32!") // 32字节密钥
}

// StringEncrypt 加密字符串
func StringEncrypt(text string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	key := getEncryptionKey()
	return StringEncryptWithKey(text, key)
}

// StringEncryptWithKey 使用指定密钥加密字符串
func StringEncryptWithKey(text string, key []byte) (string, error) {
	if len(text) == 0 || len(key) == 0 {
		return "", nil
	}
	pass := []byte(text)
	xpass, err := aesEncryptWithSalt(key, pass)
	if err == nil {
		pass64 := base64.StdEncoding.EncodeToString(xpass)
		return pass64, err
	}
	return "", err
}

// StringDecrypt 解密字符串
func StringDecrypt(text string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	key := getEncryptionKey()
	return StringDecryptWithKey(text, key)
}

// StringDecryptWithKey 使用指定密钥解密字符串
func StringDecryptWithKey(text string, key []byte) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			// 忽略panic
		}
	}()
	if len(text) == 0 {
		return "", nil
	}
	bytesPass, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	var tpass []byte
	tpass, err = aesDecryptWithSalt(key, bytesPass)
	if err == nil {
		result := string(tpass[:])
		return result, err
	}
	return "", err
}

func padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

func unPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func aesEncryptWithSalt(key, plaintext []byte) ([]byte, error) {
	plaintext = padding(plaintext, aes.BlockSize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[0:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func aesDecryptWithSalt(key, ciphertext []byte) ([]byte, error) {
	var block cipher.Block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)
	ciphertext = unPadding(ciphertext)
	return ciphertext, nil
}