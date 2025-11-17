package helper

import (
	"errors"
	"time"
	"warehouse/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      uint   `json:"user_id"`
	WarehouseId uint   `json:"warehouse_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	config.LoadConfig()
	sec := config.Cfg.JWTSecret
	return []byte(sec)
}

func CreateJWT(userID, warehouseId uint, email, role string, ttlMinutes int) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Role:        role,
		WarehouseId: warehouseId,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlMinutes) * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func ParseJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
