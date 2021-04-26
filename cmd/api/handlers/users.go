package handlers

import (
	"errors"
	"fmt"
	"github.com/gorilla/context"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
	"strings"
)

// KeyLoginUser is a key used for the LoginUser object in the context
type KeyLoginUser struct{}

// KeyTokenUser is a key used for the TokenUser object in the context
type KeyTokenUser struct{}

// KeyCreateUser is a key used for CreateUser object in the context
type KeyCreateUser struct{}

// KeyChangeUserPassword is a key used for ChangeUserPassword object in the context
type KeyChangeUserPassword struct{}

// swagger:route GET /users users listUsers
// Return a list of users from the database
//
//	Security:
//  - snippetskey:
//
// responses:
//	200: usersResponse
//  401: messageResponse
//  403: messageResponse
//	500: messageResponse

// listAllUsers handles GET requests and returns all current users
func (app *Application) listAllUsers(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	users, err := app.Users.GetAll()
	if err != nil {
		app.ErrorLog.Printf("listAllUsers: Unable to get users  %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Unable to get users list"}, rw)
		return
	}

	err = models.ToJSON(users, rw)
	if err != nil {
		// we should never be here but log the error just incase
		app.ErrorLog.Printf("listAllUsers: Unable to serializing users  %v\n", err)
	}
}

// swagger:route GET /users/{id} users listSingleUser
// Return a single user from the database
//
//	Security:
//  - snippetskey:
//
// responses:
//	200: userResponse
//	400: messageResponse
//	401: messageResponse
//	403: messageResponse
//	404: messageResponse
//  500: messageResponse

// getSingleUser handles GET requests of {id} user
func (app *Application) getSimpleUser(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	// get ID from the URL
	id, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("getSimpleUser: user %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusBadRequest)
		models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusBadRequest)}, rw)
		return
	}

	// get the user from the database
	u, err := app.Users.Get(id)
	switch err {
	case nil:
		// OK just return the user JSON
		break
	case models.ErrNoRecord:
		app.ErrorLog.Printf("getSimpleUser: user %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusNotFound)
		models.ToJSON(&models.GenericMessage{fmt.Sprintf("Unable to get user %d", id)}, rw)
		return
	default:
		app.ErrorLog.Printf("getSimpleUser: user %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{fmt.Sprintf("Unable to get user %d", id)}, rw)
		return
	}

	err = models.ToJSON(u, rw)
	if err != nil {
		// we should never be here but log the error just incase
		app.ErrorLog.Printf("getSimpleUser: Unable to serializing user %d:  %v\n", id, err)
	}
}

// swagger:route POST /users/login users loginUser
// Validate user credentials and return JWT when valid
//
// responses:
//	200: userTokenResponse
//  401: messageResponse
//	422: validationResponse
//	500: messageResponse

// loginUser handles POST requests to verify login and return JWT when user credentials are valid
func (app *Application) loginUser(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	// fetch the login user from the context
	user, ok := context.Get(r, KeyLoginUser{}).(*models.LoginUser)
	if !ok {
		app.ErrorLog.Printf("loginUser: No user data in the context\n")
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem with user data"}, rw)
		return
	}

	// Use credentials to obtain user ID
	id, err := app.Users.Authenticate(user.Username, user.Password)
	if err != nil {
		app.ErrorLog.Printf("loginUser: %v\n", err)
		rw.WriteHeader(http.StatusUnauthorized)
		models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusUnauthorized)}, rw)
		return
	}

	// get user data from the database
	u, err := app.Users.Get(id)
	if err != nil {
		app.ErrorLog.Printf("loginUser: get user: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"unable to get user"}, rw)
		return
	}

	// create a JWT for the user
	if app.DebugOn {
		app.InfoLog.Printf("loginUser: creating JWT for %q\n", user.Username)
	}
	token, err := app.Tokens.CreateToken(u)
	if err != nil {
		app.ErrorLog.Printf("loginUser: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"unable to create JWT"}, rw)
		return
	}
	if app.DebugOn {
		app.InfoLog.Printf("loginUser: created JWT\n")
	}

	// verify the token - redundant check as the token was just created
	err = app.Tokens.VerifyToken(&token)
	if err != nil {
		app.ErrorLog.Printf("loginUser: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"invalid JWT"}, rw)
		return
	}

	// get user data from token - redundant as the data was already availabre befr creating the token
	tUser, err := models.GetUserFromToken(&token)
	if err != nil {
		app.ErrorLog.Printf("loginUser: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"unable to get user from JWT"}, rw)
		return
	}

	//  create message with user data and token to reply back
	tokenmsg := models.TokenMessage{
		Token: token,
		User:  *tUser,
	}
	models.ToJSON(tokenmsg, rw)
}

// swagger:route POST /users users createUser
// Create and inserts a new user on the database
//
//	Security:
//  - snippetskey:
//
// responses:
//	200: messageResponse
//  400: messageResponse
//  401: messageResponse
//  403: messageResponse
//	422: validationResponse
//	500: messageResponse

// createUser handles POST requests to create a new user and insert it in the database
func (app *Application) createUser(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	// fetch the login user from the context
	user, ok := context.Get(r, KeyCreateUser{}).(*models.CreateUser)
	if !ok {
		app.ErrorLog.Printf("createUser: No user data in the context\n")
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem with user data"}, rw)
		return
	}

	// Insert user on the database
	err := app.Users.Insert(user.Name, user.Email, user.Password, user.Roles)
	if err != nil {
		app.ErrorLog.Printf("createUser: %v\n", err)
		if errors.Is(err, models.ErrDuplicateEmail) {
			rw.WriteHeader(http.StatusBadRequest)
			models.ToJSON(&models.GenericMessage{err.Error()}, rw)
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
		models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusBadRequest)}, rw)
		return
	}

	if app.DebugOn {
		app.InfoLog.Printf("createUser: created %q user\n", user.Name)
	}

	//  create message to reply back
	msg := fmt.Sprintf("User %q created with success", user.Name)
	models.ToJSON(&models.GenericMessage{msg}, rw)
}

// swagger:route PUT /users/{id}/change-password users changeUserPassword
// Change password for user {id} on the database
//
//	Security:
//  - snippetskey:
//
// responses:
//	200: messageResponse
//  400: messageResponse
//  401: messageResponse
//  403: messageResponse
//	422: validationResponse
//	500: messageResponse

// createUser handles POST requests to create a new user and insert it in the database
func (app *Application) changeUserPassword(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	// fetch the login user from the context
	user, ok := context.Get(r, KeyChangeUserPassword{}).(*models.ChangeUserPassword)
	if !ok {
		app.ErrorLog.Printf("changeUserPassword: No user data in the context\n")
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem with user data"}, rw)
		return
	}

	// get ID from the URL
	id, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("changeUserPassword: user %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusBadRequest)
		models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusBadRequest)}, rw)
		return
	}

	// fetch the token user from the context
	tuser, ok := context.Get(r, KeyTokenUser{}).(*models.TokenUser)
	if !ok {
		app.ErrorLog.Printf("changeUserPassword: No token user data in the context\n")
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem with user data"}, rw)
		return
	}

	// Change user's password on the database
	if tuser.IsAdmin() && !(tuser.ID == id) {
		// ignore old password if user is an administrator and not its own account
		err = app.Users.ResetPassword(id, user.NewPassword)
	} else {
		err = app.Users.ChangePassword(id, user.OldPassword, user.NewPassword)
	}
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			rw.WriteHeader(http.StatusBadRequest)
			models.ToJSON(&models.GenericMessage{err.Error()}, rw)
			return
		}
		app.ErrorLog.Printf("changeUserPassword: %v\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusBadRequest)}, rw)
		return
	}

	if app.DebugOn {
		app.InfoLog.Printf("changeUserPassword: password changed for user %d\n", id)
	}

	//  create message to reply back
	msg := fmt.Sprintf("Password changed for user %d with success", id)
	models.ToJSON(&models.GenericMessage{msg}, rw)
}

// Authenticate provides Authentication middleware for handlers
func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Get token from the Authorization header
		// format: Authentication: <token>
		tokens, ok := r.Header["Authentication"]
		// app.InfoLog.Printf("authenticate: tokens: %v\n",tokens)
		if ok && len(tokens) >= 1 {
			tokenString = tokens[0]
			// remove Bearer prefix if added to token - just for retro compatibility
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		// If the token is empty...
		if tokenString == "" {
			// If we get here, the required token is missing
			app.ErrorLog.Printf("authenticate: Missing JWT: %s %s\n", r.Method, r.URL.RequestURI())
			rw.WriteHeader(http.StatusUnauthorized)
			models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusUnauthorized)}, rw)
			return
		}

		// verify the token
		err := app.Tokens.VerifyToken(&tokenString)
		if err != nil {
			// If we get here, the required token is missing
			app.ErrorLog.Printf("authenticate: Invalid JWT: %v\n", err)
			rw.WriteHeader(http.StatusUnauthorized)
			models.ToJSON(&models.GenericMessage{"invalid JWT"}, rw)
			return
		}

		// get user data from token
		tUser, err := models.GetUserFromToken(&tokenString)
		if err != nil {
			app.ErrorLog.Printf("authenticate: %v\n", err)
			rw.WriteHeader(http.StatusInternalServerError)
			models.ToJSON(&models.GenericMessage{"unable to get user from JWT"}, rw)
			return
		}

		// Everything worked! Set the TokenUser in the context.
		context.Set(r, KeyTokenUser{}, tUser)
		next.ServeHTTP(rw, r)
	})
}

// authorize provides authorization middleware for handlers
// If the user has any of the required permissions or has AministrationRole than it is authorized
// when SelfRole is required the check is made between URL ID request and user ID
func (app *Application) authorize(permissions ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

			// fetch the token user from the context
			user, ok := context.Get(r, KeyTokenUser{}).(*models.TokenUser)
			if !ok {
				app.ErrorLog.Printf("authorize: No user data in the context\n")
				rw.WriteHeader(http.StatusInternalServerError)
				models.ToJSON(&models.GenericMessage{"Problem with user data"}, rw)
				return
			}

			if app.DebugOn {
				app.InfoLog.Printf("authorize: User: %q - %v - Required permission: %v\n", user.Name, user.Roles, permissions)
			}

			// the user needs to have any the requested permissions to be authorized
			isAuthorised := false
			for _, permission := range permissions {
				if permission == models.SelfRole {
					// authorize if token user is the same as id specified on URL and SerlRole permission exisys
					id, err := getID(r)
					if err != nil {
						app.ErrorLog.Printf("authorize: SelfRole check: No user ID specified on URL: %v\n", err)
					} else {
						// only fetch user profile if id corresponds to logged in user or an administrator
						if id == user.ID {
							// authorize OK
							isAuthorised = true
							break
						}
					}
				}
				if err := models.CheckUserPermission(&user.Roles, permission); err == nil {
					isAuthorised = true
					break
				} else if app.DebugOn {
					app.ErrorLog.Printf("authorize: %v\n", err)
				}
			}
			if !isAuthorised {
				app.ErrorLog.Printf("authorize: %s\n", http.StatusText(http.StatusForbidden))
				rw.WriteHeader(http.StatusForbidden)
				models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusForbidden)}, rw)
				return
			}
			if app.DebugOn {
				app.InfoLog.Println("authorize: Authorization is OK")
			}
			next.ServeHTTP(rw, r)
		})
	}
}

// swagger:route GET /users/role-types users listRoles
// Return a list of valid user role types from the database
//
//	Security:
//  - snippetskey:
//
// responses:
//	200: rolesResponse
//  401: messageResponse
//  403: messageResponse
//	500: messageResponse

// listAllRoleTypes handles GET requests and returns all role types
func (app *Application) listAllRoleTypes(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	roles, err := app.Users.GetRoleTypes()
	if err != nil {
		app.ErrorLog.Printf("listAllRoleTypes: Unable to get role types  %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Unable to get role types list"}, rw)
		return
	}

	err = models.ToJSON(roles, rw)
	if err != nil {
		// we should never be here but log the error just incase
		app.ErrorLog.Printf("listAllRoleTypes: Unable to serializing role types  %v\n", err)
	}
}
