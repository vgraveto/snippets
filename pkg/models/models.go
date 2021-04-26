package models

import (
	"errors"
)

var (
	// ErrNoRecord error if no record if found for a specified request
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrBadRequest error if provided request has invalid data
	ErrBadRequest = errors.New("models: request data not valid")
	// ErrInvalidCredentials error if a user tries to login with an incorrect email address or password.
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail error if a user tries to signup with an email address that's already in use.
	ErrDuplicateEmail = errors.New("models: duplicate email")
	// ErrValidation error if a user tries to signup with an email address that's already in use.
	ErrValidation = errors.New("models: validation error")
)

// GenericMessage is a generic message returned by a server
type GenericMessage struct {
	Message string `json:"message"`
}

// ValidationError is a collection of validation error messages
type ValidationMessagesError struct {
	Messages []string `json:"messages"`
}
