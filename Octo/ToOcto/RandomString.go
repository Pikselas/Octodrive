package ToOcto

import (
	"encoding/base64"
	"math/rand"
)

func RandomString(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(randomBytes)[:length]
}
