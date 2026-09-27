package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

func SHA256(file multipart.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}
