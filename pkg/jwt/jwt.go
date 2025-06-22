package jwt

import (
	"errors"
	"strings"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type JWT struct {
	key                   []byte
	expiresByAccessToken  time.Duration
	expiresByRefreshToken time.Duration
}

type CustomClaims struct {
	UserId uint64
	jwt.RegisteredClaims
}

func NewJwt(conf *viper.Viper) *JWT {
	return &JWT{
		key:                   []byte(conf.GetString("security.jwt.key")),
		expiresByAccessToken:  time.Duration(conf.GetInt64("security.jwt.expiresByAccessToken")),
		expiresByRefreshToken: time.Duration(conf.GetInt64("security.jwt.expiresByRefreshToken")),
	}
}

func (j *JWT) GetAckExpires() int64 {
	return int64(j.expiresByAccessToken.Seconds())
}

func (j *JWT) GetRefreshExpires() int64 {
	return int64(j.expiresByRefreshToken.Seconds())
}

func (j *JWT) genToken(userId uint64, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "",
			Subject:   "",
			ID:        "",
			Audience:  []string{},
		},
	})
	
	// Sign and get the complete encoded token as a string using the key
	tokenString, err := token.SignedString(j.key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) GenTokenPair(userId uint64) (accessToken, refreshToken string, err error) {
	accessToken, err = j.genToken(userId, time.Now().Add(time.Hour*j.expiresByAccessToken))
	if err != nil {
		return "", "", err
	}
	refreshToken, err = j.genToken(userId, time.Now().Add(time.Hour*j.expiresByRefreshToken))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
	
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("token is empty")
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.key, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
