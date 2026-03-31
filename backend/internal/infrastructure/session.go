package infrastructure

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/your-org/webapp-template/internal/domain"
)

// JWTSessionService は JWT ベースのセッションサービスです。
type JWTSessionService struct {
	secretKey []byte
	duration  time.Duration
	issuer    string
}

// NewJWTSessionService は JWTSessionService を生成します。
func NewJWTSessionService(secret string, duration time.Duration) *JWTSessionService {
	return &JWTSessionService{
		secretKey: []byte(secret),
		duration:  duration,
		issuer:    "webapp-template",
	}
}

type sessionJWTClaims struct {
	jwt.RegisteredClaims
	Email string      `json:"email"`
	Role  domain.Role `json:"role"`
}

// IssueToken はセッション JWT を生成します。
func (s *JWTSessionService) IssueToken(claims *domain.SessionClaims) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &sessionJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.duration)),
		},
		Email: claims.Email,
		Role:  claims.Role,
	})

	signed, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// ValidateToken はセッション JWT を検証し、クレームを返します。
func (s *JWTSessionService) ValidateToken(tokenString string) (*domain.SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &sessionJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUnauthorized, err)
	}

	claims, ok := token.Claims.(*sessionJWTClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	return &domain.SessionClaims{
		UserID: claims.Subject,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
