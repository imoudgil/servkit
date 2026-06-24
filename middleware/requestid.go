package middleware

import (
	"crypto/rand"
	"encoding/hex"
)

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "fallback-request-id"
	}
	return hex.EncodeToString(b[:])
}
