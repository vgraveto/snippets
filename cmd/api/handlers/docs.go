// Package classification Snippets API
//
// Documentation for Snippets API used for database manipulation
//
//	Schemes: https
//	BasePath: /
//	Version: 0.0.4
//	Contact: Vitor Graveto<vitor@wexcedo.com>
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//  SecurityDefinitions:
//  snippetskey:
//    type: apiKey
//    description: JSON Web Token (JWT) - token
//    name: Authentication
//    in: header
//
// swagger:meta
package handlers

import (
	"github.com/vgraveto/snippets/pkg/models"
)

//
// NOTE: Types defined here are purely for documentation purposes
// these types are not used by any of the handlers

// Text returned as a string
// swagger:response stringResponse
type stringResponseWrapper struct {
	// Text string
	// in: body
	Text string
}

// Generic message returned as a JSON string
// swagger:response messageResponse
type messageResponseWrapper struct {
	// Description of the message
	// in: body
	Body models.GenericMessage
}

// Validation errors defined as an array of strings
// swagger:response validationResponse
type errorValidationWrapper struct {
	// Collection of the errors
	// in: body
	Body models.ValidationMessagesError
}

// A list of Snippets
// swagger:response snippetsResponse
type snippetsResponseWrapper struct {
	// All current snippets
	// in: body
	Body []models.Snippet
}

// Data structure representing a single snippet
// swagger:response snippetResponse
type snippetResponseWrapper struct {
	// A snippet
	// in: body
	Body models.Snippet
}

// A list of users
// swagger:response usersResponse
type usersResponseWrapper struct {
	// All current users
	// in: body
	Body []models.User
}

// Data structure representing a single user
// swagger:response userResponse
type userResponseWrapper struct {
	// A user
	// in: body
	Body models.User
}

// No content is returned by this API endpoint
// swagger:response noContentResponse
type noContentResponseWrapper struct {
}

// Data structure representing the user token record
// swagger:response userTokenResponse
type userTokenResponseWrapper struct {
	// A snippet
	// in: body
	Body models.TokenMessage
}

// A list of role types
// swagger:response rolesResponse
type rolesResponseWrapper struct {
	// All current role types
	// in: body
	Body []models.RoleType
}

// swagger:parameters loginUser
type loginUserParamsWrapper struct {
	// Data structure to login with user credentials.
	// in: body
	// required: true
	Body models.LoginUser
}

// swagger:parameters createUser
type createUserParamsWrapper struct {
	// Data structure to create an user.
	// in: body
	// required: true
	Body models.CreateUser
}

// swagger:parameters changeUserPassword
type changeUserPasswordParamsWrapper struct {
	// The ID of the user to which the operation relates
	// in: path
	// required: true
	ID int `json:"id"`

	// Data structure to change the user password
	// in: body
	// required: true
	Body models.ChangeUserPassword
}

// swagger:parameters listSingleUser listSingleSnippet
type idParamsWrapper struct {
	// The ID for which the operation relates
	// in: path
	// required: true
	ID int `json:"id"`
}

// swagger:parameters createSnippet
type createSnippetParamsWrapper struct {
	// Data structure to create a new snippet
	// in: body
	// required: true
	Body models.SnippetCreate
}
