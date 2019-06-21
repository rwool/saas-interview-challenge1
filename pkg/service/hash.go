package service

import (
	"crypto/sha256"
	"encoding/base64"
)

func createStringSHA256(s string) string {
	sha := sha256.New()
	sum := sha.Sum([]byte(s))
	b64SHA256 := base64.StdEncoding.EncodeToString(sum)
	return b64SHA256
}
