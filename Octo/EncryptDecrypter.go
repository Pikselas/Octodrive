package Octo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type Encrypter interface {
	Encrypt(io.Reader) (io.Reader, error)
}

type Decrypter interface {
	Decrypt(io.Reader) (io.Reader, error)
}

type EncryptDecrypter interface {
	Encrypter
	Decrypter
}

type AesEncDec struct {
	key []byte
	iv  []byte
}

func (aesED *AesEncDec) baseAesFunc(data io.Reader) (io.Reader, error) {
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

func (aesED *AesEncDec) Encrypt(data io.Reader) (io.Reader, error) {
	return aesED.baseAesFunc(data)
}

func (aesED *AesEncDec) Decrypt(data io.Reader) (io.Reader, error) {
	return aesED.baseAesFunc(data)
}

func (aesED *AesEncDec) GetKey() []byte {
	return append(aesED.key, aesED.iv...)
}

func NewAesEncDecFrom(key []byte) EncryptDecrypter {
	return &AesEncDec{key: key[:32], iv: key[32:48]}
}

func NewAesEncDec() (*AesEncDec, error) {
	key, err := GenerateKey(32)
	if err != nil {
		return nil, err
	}
	iv, err := GenerateKey(16)
	if err != nil {
		return nil, err
	}
	return &AesEncDec{key: key, iv: iv}, nil
}

func GenerateKey(len uint) ([]byte, error) {
	key := make([]byte, len)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

/*
 * Nil Encrypt Decrypter
   A dummy encrypt decrypter that does nothing
*/

type nilEncDec struct{}

func (ned *nilEncDec) Encrypt(data io.Reader) (io.Reader, error) {
	return data, nil
}

func (ned *nilEncDec) Decrypt(data io.Reader) (io.Reader, error) {
	return data, nil
}

// A dummy encrypt decrypter that does nothing
func NewNilEncDec() EncryptDecrypter {
	return &nilEncDec{}
}
