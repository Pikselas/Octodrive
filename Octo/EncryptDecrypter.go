package Octo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type EncryptDecrypter interface {
	Encrypt(io.Reader) (io.Reader, error)
	Decrypt(io.Reader) (io.Reader, error)
}

type aesEncDec struct {
	key []byte
	iv  []byte
}

func (aesED *aesEncDec) baseAesFunc(data io.Reader) (io.Reader, error) {
	// Create a new AES cipher block using the key
	block, err := aes.NewCipher(aesED.key)
	if err != nil {
		return nil, err
	}

	// Create a new CBC mode cipher using the block and IV
	stream := cipher.NewCTR(block, aesED.iv)

	// Create a new Stream Reader that encrypts data as it is read from the input file
	encryptedReader := &cipher.StreamReader{S: stream, R: data}

	return encryptedReader, nil
}

func (aesED *aesEncDec) Encrypt(data io.Reader) (io.Reader, error) {
	return aesED.baseAesFunc(data)
}

func (aesED *aesEncDec) Decrypt(data io.Reader) (io.Reader, error) {
	return aesED.baseAesFunc(data)
}

func newAesEncDec(key []byte, iv []byte) EncryptDecrypter {
	return &aesEncDec{key: key, iv: iv}
}

func generateKey(len uint) ([]byte, error) {
	key := make([]byte, len)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
