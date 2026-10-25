package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — что кладём в access-токен
type Claims struct {
	UserID string `json:"sub"`  // ID пользователя
	Role   string `json:"role"` // CLIENT / OFFICER / ADMIN
	//CustomerID string `json:"customer_id"` // ID клиента (если есть)
	Type string `json:"type"` // "access" или "refresh"
	jwt.RegisteredClaims
}

// Signer — подписывает JWT приватным ключом
type Signer struct {
	privateKey *rsa.PrivateKey
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewSigner(privateKeyPath, issuer string, accessTTL, refreshTTL time.Duration) (*Signer, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("decode PEM: invalid block")
	}

	// Пробуем PKCS#1 (как у OpenSSL genrsa)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Пробуем PKCS#8 (как у некоторых генераторов)
		parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not RSA private key")
		}
		key = rsaKey
	}

	return &Signer{
		privateKey: key,
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

// SignAccess — подписывает access-токен
func (s *Signer) SignAccess(userID, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTTL)

	claims := Claims{
		UserID: userID,
		Role:   role,
		//CustomerID: customerID,
		Type: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"api-gateway"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(), // уникальный ID токена (для blacklist)
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	return signed, int64(s.accessTTL.Seconds()), nil
}

// SignRefresh — подписывает refresh-токен
func (s *Signer) SignRefresh(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.refreshTTL)

	claims := Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"auth-service"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}

	return signed, nil
}

func generateJTI() string {
	// В реальном проекте — uuid.New().String()
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
