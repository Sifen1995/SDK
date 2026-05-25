package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateHMAC generates a SHA256 HMAC for a message given a key.
func GenerateHMAC(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
