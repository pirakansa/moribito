package githubapp

import "testing"

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "supersecret"
	body := []byte("payload")
	signature := DebugWebhookSignature(secret, body)

	if !VerifyWebhookSignature(secret, body, signature) {
		t.Fatalf("expected signature to be valid")
	}
	if VerifyWebhookSignature(secret, body, "sha256=deadbeef") {
		t.Fatalf("expected signature to be invalid")
	}
}
