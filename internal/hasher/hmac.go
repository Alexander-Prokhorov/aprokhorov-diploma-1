package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()?|><.,][}{:`~"

type HMAC struct{}

func NewHMAC() HMAC {
	return HMAC{}
}

func (hm HMAC) GetHash(src string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (hm HMAC) RandomKey() (string, error) {
	var n = 16
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String(), nil
}

func (hm HMAC) GenerateToken() (string, error) {
	return hm.RandomKey()
}
