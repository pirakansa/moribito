package githubapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifyWebhookSignature validates a GitHub webhook signature.
// GitHub sends signatures in the X-Hub-Signature-256 header.
// Returns false if:
//   - secret is empty
//   - signature doesn't have "sha256=" prefix
//   - signature doesn't match the expected HMAC
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

// DebugWebhookSignature generates the expected signature for debugging.
// Useful for testing webhook handlers locally.
func DebugWebhookSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}
