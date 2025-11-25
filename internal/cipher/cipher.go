package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

func GenerateRandomSalt() ([]byte, error) {
	salt := make([]byte, 16)
	n, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	if n != 16 {
		return nil, errors.New("failed to generate random 16-byte salt")
	}
	return salt, nil
}

func DeriveEncryptionKey32(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

func DeriveEncryptionKey64(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 64)
}

func EncryptAESGCM(cipherKey32, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(cipherKey32)
	if err != nil {
		return nil, err
	}
	aesBlock, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesBlock.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	cipherData := aesBlock.Seal(nil, nonce, data, nil)
	output := append(nonce, cipherData...)
	return output, nil
}

func DecryptAESGCM(cipherKey32, cipherData []byte) ([]byte, error) {
	block, err := aes.NewCipher(cipherKey32)
	if err != nil {
		return nil, err
	}
	aesBlock, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesBlock.NonceSize()
	if len(cipherData) < nonceSize {
		return nil, errors.New("invalid ciphertext (too short)")
	}

	nonce := cipherData[:nonceSize]
	data := cipherData[nonceSize:]

	plaintext, err := aesBlock.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
