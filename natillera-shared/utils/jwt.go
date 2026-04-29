package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtSecret = []byte("tu-clave-secreta-muy-segura") // En producción, usa variables de entorno

type Claims struct {
	Id    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Email string             `json:"email"`
	RolId primitive.ObjectID `json:"rolId"`
	jwt.RegisteredClaims
}

// GenerateToken genera un token JWT
func GenerateToken(userID primitive.ObjectID, username, email string, rolId primitive.ObjectID) (string, error) {
	claims := Claims{
		Id:    userID,
		Name:  username,
		Email: email,
		RolId: rolId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 24 horas
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken valida y decodifica un token JWT
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}
