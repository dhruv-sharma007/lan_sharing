package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// CalculateFileSHA256 computes the SHA-256 hash of a file.
func CalculateFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
