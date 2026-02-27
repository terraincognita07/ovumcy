package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	secureCookieVersion       = "v1"
	secureCookiePurposePrefix = "ovumcy.cookie."
)

var errInvalidSecureCookieValue = errors.New("invalid secure cookie value")

type secureCookieCodec struct {
	aead cipher.AEAD
}

func newSecureCookieCodec(secretKey []byte) (*secureCookieCodec, error) {
	if len(secretKey) == 0 {
		return nil, errors.New("secure cookie secret key is required")
	}

	derivedKey := deriveSecureCookieKey(secretKey)
	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return nil, fmt.Errorf("init secure cookie cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init secure cookie aead: %w", err)
	}
	return &secureCookieCodec{aead: aead}, nil
}

func deriveSecureCookieKey(secretKey []byte) [32]byte {
	label := []byte("ovumcy.secure-cookie.v1")
	material := make([]byte, 0, len(label)+len(secretKey))
	material = append(material, label...)
	material = append(material, secretKey...)
	return sha256.Sum256(material)
}

func (codec *secureCookieCodec) seal(purpose string, plaintext []byte) (string, error) {
	trimmedPurpose := strings.TrimSpace(purpose)
	if trimmedPurpose == "" {
		return "", errors.New("secure cookie purpose is required")
	}
	if codec == nil || codec.aead == nil {
		return "", errors.New("secure cookie codec is not initialized")
	}

	nonce := make([]byte, codec.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate secure cookie nonce: %w", err)
	}

	aad := []byte(secureCookiePurposePrefix + trimmedPurpose)
	ciphertext := codec.aead.Seal(nil, nonce, plaintext, aad)
	payload := make([]byte, 0, len(nonce)+len(ciphertext))
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)

	return secureCookieVersion + "." + base64.RawURLEncoding.EncodeToString(payload), nil
}

func (codec *secureCookieCodec) open(purpose string, rawValue string) ([]byte, error) {
	trimmedPurpose := strings.TrimSpace(purpose)
	if trimmedPurpose == "" {
		return nil, errors.New("secure cookie purpose is required")
	}
	if codec == nil || codec.aead == nil {
		return nil, errors.New("secure cookie codec is not initialized")
	}

	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return nil, errInvalidSecureCookieValue
	}

	version, encodedPayload, found := strings.Cut(rawValue, ".")
	if !found || version != secureCookieVersion || strings.TrimSpace(encodedPayload) == "" {
		return nil, errInvalidSecureCookieValue
	}

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, errInvalidSecureCookieValue
	}

	nonceSize := codec.aead.NonceSize()
	if len(payload) <= nonceSize {
		return nil, errInvalidSecureCookieValue
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	aad := []byte(secureCookiePurposePrefix + trimmedPurpose)
	plaintext, err := codec.aead.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, errInvalidSecureCookieValue
	}
	return plaintext, nil
}
