package impl

import (
	"Atlas/internal/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	jwt.StandardClaims
	Role string `json:"role"`
}

func (a *AuthService) CreateToken(user models.User) (string, error) {

	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: time.Now().Add(a.config.TokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		Role: user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.TokenSignedString))

}

func (a *AuthService) KeyFunc(token *jwt.Token) (any, error) {
	return []byte(a.config.TokenSignedString), nil
}
