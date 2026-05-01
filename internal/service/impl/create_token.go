package impl

import (
	"Atlas/internal/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

// Claims represents the JWT claims structure including the user role.
type Claims struct {
	jwt.StandardClaims        // StandardClaims holds subject, expiry, issued at
	Role               string // Role is the user's permission level (admin, manager, viewer)
}

// CreateToken generates a JWT token for the given user using the configured signing key and TTL.
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

// KeyFunc returns the JWT signing key for token validation.
func (a *AuthService) KeyFunc(token *jwt.Token) (any, error) {
	return []byte(a.config.TokenSignedString), nil
}
