package signer

import (
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/htranq/vortech-ome/pkg/config"
)

type Claims struct {
	jwt.StandardClaims

	TokenType string `json:"token_type"` // access_token, id_token
	// TODO add more fields if needed
}

type Token struct {
	*Claims
	Raw string
}

type Signer interface {
	Create(identity, audience string) (token *Token, err error)
	Parse(token string) (*Token, error)
}

type signer struct {
	tokenType       string
	privateKey      ed25519.PrivateKey
	expiry          time.Duration
	issuer          string
	defaultAudience string
}

func New(tokenType string, jwtSigning *config.JwtSigning) (Signer, error) {
	privateKey, err := jwt.ParseEdPrivateKeyFromPEM([]byte(jwtSigning.GetPrivateKey()))
	if err != nil {
		return nil, err
	}

	return &signer{
		tokenType:       tokenType,
		privateKey:      privateKey.(ed25519.PrivateKey),
		expiry:          time.Duration(jwtSigning.GetExpiresTime()) * time.Millisecond,
		issuer:          jwtSigning.GetIssuer(),
		defaultAudience: jwtSigning.GetDefaultAudience(),
	}, nil
}

func (t *signer) Create(identity, audience string) (*Token, error) {
	var (
		now = time.Now()
		iat = now.UTC().Unix()
		exp = now.Add(t.expiry)
		id  = uuid.New().String()
	)
	if audience == "" {
		audience = t.defaultAudience
	}

	claims := &Claims{
		TokenType: t.tokenType,
		StandardClaims: jwt.StandardClaims{
			Id:        id,
			ExpiresAt: exp.Unix(),
			IssuedAt:  iat,
			NotBefore: iat,
			Audience:  audience,
			Subject:   identity,
			Issuer:    t.issuer,
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(t.privateKey)

	return &Token{Raw: token, Claims: claims}, err
}

func (t *signer) Parse(token string) (*Token, error) {
	tk, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.privateKey.Public(), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := tk.Claims.(*Claims); ok && tk.Valid {
		return &Token{
			Claims: claims,
			Raw:    token,
		}, nil
	} else {
		return nil, err
	}
}
