package jwt

import (
	"errors"
	"rebu/shared/env"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var secret = []byte(env.GetString("JWT_SECRET", "jwt-secret"))

type UserType string

const (
	UserTypeGuest      UserType = "guest"
	UserTypeRegistered UserType = "registered"
)

type Claims struct {
	UserID   string `json:"userID"`
	UserType string `json:"userType"`
	jwt.RegisteredClaims
}

func NewGuestToken() (string, error) {
	return generate(uuid.New().String(), UserTypeGuest, 30*24*60) // 30 days * 24 hour/days * 60 minutes/hour
}

func NewUserToken(userID string) (string, error) {
	return generate(userID, UserTypeRegistered, 15) // 15minutes?
}

func generate(userID string, userType UserType, ttlMinutes int32) (string, error) {
	claims := Claims{
		UserID:   userID,
		UserType: string(userType),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "domino",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttlMinutes) * time.Minute)),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(secret)
}

// Parse valiudates the token's signature and expiring date and return its claimls
func Parse(tokenString string) (*Claims, error) {
	// NOTE: If you provide a custom claim implementation that embeds one of the standard claims (such as RegisteredClaims), make sure that a) you either embed a non-pointer version of the claims or b) if you are using a pointer, allocate the proper memory for it before passing in the overall claims, otherwise you might run into a panic.

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		//When a client sends a JWT, the library first decodes the Header of the token. The header contains unverified information, such as the algorithm used to sign it (e.g., HS256, RS256). The library does not trust this header yet. It passes the unverified token structure (t) into your callback function so you can run safety checks before it attempts to verify the signature.

		func(t *jwt.Token) (any, error) {
			// Check if signing-method is Hmac or also known as HS256
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
