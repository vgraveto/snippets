package models

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strings"
	"time"
)

var (
	ErrUnauthorizedToken = errors.New("models: unauthorized JWT used")
	ErrForbiddenToken    = errors.New("models: forbidden JWT used")
	ErrExpiredToken      = errors.New("models: expired JWT used")
)

type Tokens interface {
	CreateToken(*User) (string, error)
	VerifyToken(*string) error
}

type TokenModel struct {
	td *TokenData
}

func NewTokenModel(d *TokenData) *TokenModel {
	return &TokenModel{td: d}
}

// api JWT token data from config file
type TokenData struct {
	TokenIssuerName string
	TokenValidTime  time.Duration // number of hours
	TokenSigningKey string
}

// Define the token user structure
type TokenUser struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// IsAdmin returns true is the TokenUser has AdministratorRole in its Roles and false otherwise
func (t *TokenUser) IsAdmin() bool {
	isAdmin := false
	for _, r := range t.Roles {
		if r == AministratorRole {
			isAdmin = true
			break
		}
	}
	return isAdmin
}

type TokenMessage struct {
	User  TokenUser `json:"user"`
	Token string    `json:"token"`
}

// Define the token claims structure
type MyCustomClaims struct {
	User *TokenUser `json:"user,omitempty"`
	jwt.StandardClaims
}

// CreateToken creates token with TokenUser data inside claims
func (t *TokenModel) CreateToken(user *User) (string, error) {
	if user == nil {
		return "", errors.New("createToken: user not defined")
	}

	tokenUser := &TokenUser{
		ID:    user.ID,
		Name:  user.Name,
		Roles: user.Roles,
	}
	// Create the Claims
	claims := MyCustomClaims{
		tokenUser,
		jwt.StandardClaims{
			Subject:   "User JWT",
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Duration(t.td.TokenValidTime)).Unix(),
			Issuer:    t.td.TokenIssuerName,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with key
	tokenString, err := token.SignedString([]byte(t.td.TokenSigningKey))
	if err != nil {
		return "", errors.New("createToken: failed to sign token")
	}

	return tokenString, nil
}

// VerifyToken verifies if token is valid
func (t *TokenModel) VerifyToken(tokenString *string) error {

	// If the token is empty...
	if *tokenString == "" {
		// If we get here, the required token is missing
		return fmt.Errorf("VerifyToken: empty token")
	}

	// Now parse the token
	token, err := jwt.Parse(*tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("VerifyToken: Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.td.TokenSigningKey), nil
	})
	if err != nil {
		return fmt.Errorf("VerifyToken: Invalid Token: %v", err)
	}

	// Check if token is valid
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Everything worked. Token valid and claims ok!
		return nil
	}

	return fmt.Errorf("VerifyToken: Invalid Token")
}

// GetClaimsFromToken verifies if token format is valid and extract claims data
func GetClaimsFromToken(tokenString *string) (*MyCustomClaims, error) {

	// If the token is empty...
	if *tokenString == "" {
		// If we get here, the required token is missing
		return nil, fmt.Errorf("GetClaimsFromToken: empty token")
	}

	// Get claims part of the token
	txt := strings.SplitAfter(*tokenString, ".")
	if len(txt) != 3 {
		return nil, fmt.Errorf("GetClaimsFromToken: Invalid token string")
	}
	txt1 := txt[1]
	if len(txt1) < 1 {
		return nil, fmt.Errorf("GetClaimsFromToken: Invalid token claims string")
	}
	txt2 := txt1[:len(txt1)-1]
	padding := len(txt2) % 4
	if padding > 1 { // add the required base64 padding
		for i := padding; i < 4; i++ {
			txt2 = txt2 + "="
		}
	}
	jsonClaims, err := base64.URLEncoding.DecodeString(txt2)
	if err != nil {
		return nil, fmt.Errorf("GetClaimsFromToken: base64: %v", err)
	}
	// log.Printf("GetClaimsFromToken: Token is OK. Claims:%s\n",string(jsonClaims))

	// obtain struct from json string of claims
	claims := &MyCustomClaims{}
	err = FromJSON(claims, strings.NewReader(string(jsonClaims)))
	if err != nil {
		return nil, fmt.Errorf("GetClaimsFromToken: json decoding: %v", err)
	}
	return claims, nil

}

// GetUserFromToken verifies if token format is valid and extract TokenUser data from claims
func GetUserFromToken(tokenString *string) (*TokenUser, error) {

	claims, err := GetClaimsFromToken(tokenString)
	if err != nil {
		return nil, err
	}

	return claims.User, nil

}
