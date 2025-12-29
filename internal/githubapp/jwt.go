// Package githubapp provides GitHub App authentication utilities.
// It handles JWT creation for App authentication and installation token management.
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
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// jwtSkewLeeway accounts for clock skew between servers.
	jwtSkewLeeway = 60 * time.Second
	// jwtExpiry is the JWT validity period (GitHub allows up to 10 minutes).
	jwtExpiry = 9 * time.Minute
)

// CreateAppJWT generates a signed JWT for GitHub App authentication.
// The JWT is used to authenticate as the App itself (not an installation).
// Parameters:
//   - appID: GitHub App ID
//   - privateKeyPath: path to RSA private key in PEM format
//   - now: current time (for testing, use time.Now() in production)
func CreateAppJWT(appID int64, privateKeyPath string, now time.Time) (string, error) {
	if appID == 0 {
		return "", fmt.Errorf("app id is required")
	}
	key, err := readPrivateKey(privateKeyPath)
	if err != nil {
		return "", err
	}

	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}
	payload := map[string]any{
		"iat": now.Add(-jwtSkewLeeway).Unix(),
		"exp": now.Add(jwtExpiry).Unix(),
		"iss": appID,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	enc := base64.RawURLEncoding
	headerPart := enc.EncodeToString(headerJSON)
	payloadPart := enc.EncodeToString(payloadJSON)
	unsigned := headerPart + "." + payloadPart

	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	return unsigned + "." + enc.EncodeToString(signature), nil
}

// readPrivateKey loads an RSA private key from a PEM file.
// Supports both PKCS#1 (RSA PRIVATE KEY) and PKCS#8 (PRIVATE KEY) formats.
func readPrivateKey(path string) (*rsa.PrivateKey, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("private key path is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("decode private key: invalid PEM")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs1 private key: %w", err)
		}
		return key, nil
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs8 private key: %w", err)
		}
		key, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return key, nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}
