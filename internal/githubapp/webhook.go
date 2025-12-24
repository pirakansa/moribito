package githubapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifyWebhookSignature validates GitHub webhook signatures (sha256).
func VerifyWebhookSignature(secret string, body []byte, signatureHeader string) bool {
	if strings.TrimSpace(secret) == "" {
		return false
	}
	prefix := "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return false
	}
	signatureHex := strings.TrimPrefix(signatureHeader, prefix)
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	return hmac.Equal(signature, expected)
}

func DebugWebhookSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}
