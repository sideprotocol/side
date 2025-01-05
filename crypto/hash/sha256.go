package hash

import (
	"crypto/sha256"
)

// Sha256 returns the SHA256 hash of the given data
func Sha256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}
