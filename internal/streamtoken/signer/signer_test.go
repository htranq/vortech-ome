package signer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/htranq/vortech-ome/pkg/config"
)

const (
	// PEM encoded ed25519 private key
	_privateKey      = "-----BEGIN PRIVATE KEY-----\nMC4CAQAwBQYDK2VwBCIEIBPxxlFTu0gY2Y9nlXqer3RbdXY2/y1U9odS2JeYLs1r\n-----END PRIVATE KEY-----"
	_issuer          = "stream_management"
	_type            = "access_token"
	_expiresTime     = 5 * 60 * 1000 // 5 minute in (milliseconds)
	_userID          = "lucas"
	_defaultAudience = "my_signer"
)

func TestSign(t *testing.T) {
	//t.Skip("Manual generate new token from private key")

	signing := &config.JwtSigning{
		PrivateKey:      _privateKey,
		ExpiresTime:     _expiresTime,
		Issuer:          _issuer,
		DefaultAudience: _defaultAudience,
	}
	signer, err := New(_type, signing)
	assert.NoError(t, err)

	token, err := signer.Create(_userID, "")
	assert.NoError(t, err)

	fmt.Printf("JTI: %s\n", token.Claims.ID)
	fmt.Printf("Token: %s\n", token.Raw)
	fmt.Printf("ExpiresAt: %v\n", token.Claims.ExpiresAt)

	tk, err := signer.Parse(token.Raw)
	assert.NoError(t, err)
	assert.Equal(t, _type, token.Claims.TokenType)
	assert.Equal(t, token.ExpiresAt, tk.Claims.ExpiresAt)

	claims, _ := json.Marshal(token.Claims)
	fmt.Printf("Claims: %s\n", string(claims))
}

func TestSignSimple(t *testing.T) {
	//t.Skip()

	var (
		_expiry   = 100 * 365 * 24 * time.Hour
		_audience = "my_signer"
		_subject  = "lucas"
		_issuer   = "stream_management"
	)

	privateKey, err := jwt.ParseEdPrivateKeyFromPEM([]byte(_privateKey))
	assert.NoError(t, err)

	now := time.Now()
	iat := now
	exp := now.Add(_expiry)
	claims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(iat),
		NotBefore: jwt.NewNumericDate(iat),
		Audience:  []string{_audience},
		Subject:   _subject,
		Issuer:    _issuer,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(privateKey)
	assert.NoError(t, err)

	println(token)
}

func TestParse(t *testing.T) {
	t.Skip("Manual generate new token from private key")

	signing := &config.JwtSigning{
		PrivateKey:      _privateKey,
		ExpiresTime:     _expiresTime,
		Issuer:          _issuer,
		DefaultAudience: _defaultAudience,
	}
	signer, err := New(_type, signing)
	assert.NoError(t, err)

	token, err := signer.Parse("") // <- TODO put token here
	if err != nil {
		fmt.Println(err)
	}

	v, _ := json.Marshal(token)

	fmt.Println(string(v))
}
