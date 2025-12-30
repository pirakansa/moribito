package githubapp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateAppJWT(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	tmpFile, err := os.CreateTemp(t.TempDir(), "key-*.pem")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmpFile.Write(pemBytes); err != nil {
		t.Fatalf("write key: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close key: %v", err)
	}

	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	token, err := CreateAppJWT(1234, tmpFile.Name(), now)
	if err != nil {
		t.Fatalf("create jwt: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token format: %s", token)
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode claims: %v", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}
	if claims["iss"].(float64) != 1234 {
		t.Fatalf("unexpected iss claim: %v", claims["iss"])
	}

	unsigned := strings.Join(parts[:2], ".")
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	hash := sha256.Sum256([]byte(unsigned))
	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hash[:], sig); err != nil {
		t.Fatalf("verify signature: %v", err)
	}
}
