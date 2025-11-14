package yookassa

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func VerifyWebhookSignature(payload []byte, signature, secret string) error {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if signature != expected {
		return errors.New("invalid webhook signature")
	}
	return nil
}
