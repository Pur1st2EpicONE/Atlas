package impl

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

func (a *AuthService) CreateToken(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: time.Now().Add(a.config.TokenTTL).Unix(),
		IssuedAt:  time.Now().Unix()})
	return token.SignedString([]byte(a.config.TokenSignedString))
}
