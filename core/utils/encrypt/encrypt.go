package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// StringEncrypt 加密字符串
func StringEncrypt(text string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	key := []byte("gpanel-encrypt-key-32-bytes-length") // 32字节密钥
	return StringEncryptWithKey(text, key)
}

// StringEncryptWithKey 使用指定密钥加密字符串
func StringEncryptWithKey(text string, key []byte) (string, error) {
	if len(text) == 0 || len(key) == 0 {
		return "", nil
	}
	// 确保密钥长度为16、24或32字节（AES-128、AES-192、AES-256）
	if len(key) < 16 {
		// 如果密钥太短，填充到16字节
		paddedKey := make([]byte, 16)
		copy(paddedKey, key)
		key = paddedKey
	} else if len(key) > 32 {
		// 如果密钥太长，截断到32字节
		key = key[:32]
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
	// 尝试用多个可能的密钥解密（向后兼容）
	keys := [][]byte{
		[]byte("gpanel-encrypt-key-32-bytes-length"),     // 32字节（当前）
		[]byte("gpanel-encrypt-key-32-bytes-length!!"),   // 34字节（可能）
		[]byte("gpanel-encrypt-key-32-bytes-length!"),    // 33字节（可能）
		[]byte("gpanel-encrypt-key-32-bytes!!"),          // 29字节（旧）
		[]byte("gpanel-encrypt-key-32-bytes!"),           // 28字节（更旧）
		[]byte("gpanel-encrypt-key-32-bytes-length-extra"), // 38字节（可能）
	}
	
	for _, key := range keys {
		result, err := StringDecryptWithKey(text, key)
		if err == nil && result != "" {
			return result, nil
		}
	}
	
	return "", errors.New("failed to decrypt with all possible keys")
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
	
	// 尝试用原始密钥解密
	tpass, err = aesDecryptWithSalt(key, bytesPass)
	if err == nil {
		result := string(tpass[:])
		return result, nil
	}
	
	// 如果失败，尝试用调整后的密钥解密
	adjustedKey := adjustKeyLength(key)
	tpass, err = aesDecryptWithSalt(adjustedKey, bytesPass)
	if err == nil {
		result := string(tpass[:])
		return result, nil
	}
	
	return "", err
}

// adjustKeyLength 调整密钥长度以符合AES要求（16、24或32字节）
func adjustKeyLength(key []byte) []byte {
	keyLen := len(key)
	if keyLen <= 16 {
		// 填充到16字节
		adjusted := make([]byte, 16)
		copy(adjusted, key)
		return adjusted
	} else if keyLen <= 24 {
		// 填充到24字节
		adjusted := make([]byte, 24)
		copy(adjusted, key)
		return adjusted
	} else {
		// 截断到32字节
		adjusted := make([]byte, 32)
		copy(adjusted, key)
		return adjusted
	}
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