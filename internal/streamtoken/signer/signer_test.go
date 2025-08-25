package signer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/htranq/vortech-ome/pkg/config"
)

const (
	_privateKey  = "" // PEM encoded ed25519 private key
	_issuer      = "stream_management"
	_type        = "access_token"
	_expiresTime = 84600 // seconds
	_userID      = "lucas"
)

func TestSign(t *testing.T) {
	t.Skip("Manual generate new token from private key")

	signing := &config.JwtSigning{
		PrivateKey:  _privateKey,
		ExpiresTime: _expiresTime,
		Issuer:      _issuer,
	}
	signer, err := New(_type, signing)
	assert.NoError(t, err)

	token, err := signer.Create(_userID, "default")
	assert.NoError(t, err)

	fmt.Printf("JTI: %s\n", token.Claims.Id)
	fmt.Printf("Token: %s\n", token.Raw)
	fmt.Printf("ExpiresAt: %v\n", token.Claims.ExpiresAt)

	tk, err := signer.Parse(token.Raw)
	assert.NoError(t, err)
	assert.True(t, tk.Valid() == nil)
	assert.Equal(t, _type, token.Claims.TokenType)
	assert.Equal(t, token.ExpiresAt, tk.Claims.ExpiresAt)
}

func TestSignSimple(t *testing.T) {
	t.Skip()

	var (
		_privateKey = ""
		_expiry     = 100 * 365 * 24 * time.Hour
		_audience   = ""
		_subject    = ""
		_issuer     = ""
	)

	privateKey, err := jwt.ParseEdPrivateKeyFromPEM([]byte(_privateKey))
	assert.NoError(t, err)

	now := time.Now()
	iat := now.UTC().Unix()
	exp := now.Add(_expiry)
	claims := jwt.StandardClaims{
		Id:        uuid.New().String(),
		ExpiresAt: exp.Unix(),
		IssuedAt:  iat,
		NotBefore: iat,
		Audience:  _audience,
		Subject:   _subject,
		Issuer:    _issuer,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(privateKey)
	assert.NoError(t, err)

	println(token)
}

func TestParse(t *testing.T) {
	signing := &config.JwtSigning{
		PrivateKey:  _privateKey,
		ExpiresTime: _expiresTime,
		Issuer:      _issuer,
	}
	signer, err := New(_type, signing)
	assert.NoError(t, err)

	token, err := signer.Parse("")
	if err != nil {
		fmt.Println(err)
	}

	v, _ := json.Marshal(token)

	fmt.Println(string(v))
}
