package session

import (
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type SessionProvider struct {
	Config *config.Config
}

const (
	authorizationHeaderKey = "Authorization"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func (p *SessionProvider) ComparePasswords(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (p *SessionProvider) CreateToken(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID: userID,
	})

	signed, err := token.SignedString([]byte(*p.Config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

func (p *SessionProvider) ParseToken(c *gin.Context) (int64, error) {
	auth := c.GetHeader(authorizationHeaderKey)
	if len(auth) < 7 {
		return 0, errors.New("invalid token")
	}

	tokenString := auth[7:]
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(*p.Config.SecretKey), nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return 0, jwt.ErrTokenNotValidYet
	}

	if claims.UserID == 0 {
		return 0, jwt.ErrInvalidKey
	}

	return claims.UserID, nil
}

func (p *SessionProvider) CheckToken(c *gin.Context) error {
	auth := c.GetHeader(authorizationHeaderKey)
	if len(auth) < 7 {
		return errors.New("invalid token")
	}

	tokenString := auth[7:]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(*p.Config.SecretKey), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return jwt.ErrTokenNotValidYet
	}

	if claims.UserID == 0 {
		return jwt.ErrInvalidKey
	}

	return nil
}
