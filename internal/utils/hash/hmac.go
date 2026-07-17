package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GetHash returns the hex-encoded HMAC-SHA256 signature of body using the given key.
func GetHash(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
