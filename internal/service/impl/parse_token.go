package impl

import (
	"Atlas/internal/errs"
	"strconv"

	"github.com/golang-jwt/jwt"
)

func (a *AuthService) ParseToken(tokenString string) (int64, error) {

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, a.keyFunc)
	if err != nil {
		return 0, errs.ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return 0, errs.ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, errs.ErrInvalidUserID
	}

	return userID, nil

}

func (a *AuthService) keyFunc(token *jwt.Token) (any, error) {
	return []byte(a.config.TokenSignedString), nil
}
