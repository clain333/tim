package pkg

import (
	"cc.tim/client/config"
	"github.com/bwmarrin/snowflake"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	DeviceId uint64 `json:"device_id"`
	UserId   uint64 `json:"user_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(config.Config.Jwt.Secret)

// GenerateToken 生成 JWT
func generateToken(email string, uid, did uint64, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Email:    email,
		UserId:   uid,
		DeviceId: did,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "tim",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析 JWT
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
func GenerateLoginToken(email string, uid, did uint64) (string, uint64, error) {
	if uid == 0 {
		snowflakes, _ := snowflake.NewNode(1)
		uid = uint64(snowflakes.Generate())
	}
	token, err := generateToken(email, uid, did, 3*24*time.Hour)
	return token, did, err
}
