package queue

import (
	"crypto/sha256"
	"fmt"
)

func hashKey(key string) string {
	h := sha256.New()

	h.Write([]byte(key))
	return fmt.Sprintf("%x", h.Sum(nil))
}
