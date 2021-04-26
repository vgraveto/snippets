package handlers

import (
	"context"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
	"time"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("X-XSS-Protection", "1; mode=block")
		rw.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(rw, r)
	})
}

func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		app.InfoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(rw, r)
	})
}

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			//panic or not. If there has...
			if err := recover(); err != nil {
				app.ErrorLog.Println("recoverPanic: recover ongoing")
				// Set a "Connection: close" header on the response.
				rw.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				//Internal Server response.
				app.serverError(rw, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

func (app *Application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		if !app.isAuthenticated(r) {
			app.Session.Put(r, KeySessionFlash, "Operation requires authentication, please logon!")
			app.Session.Put(r, KeySessionRedirectPath, r.URL.Path)
			http.Redirect(rw, r, "/user/login", http.StatusFound)
			return
		}

		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		rw.Header().Add("Cache-Control", "no-store")

		// And call the next handler in the chain.
		next.ServeHTTP(rw, r)
	})
}

// Create a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly flags set.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Check if a user value exists in the session. If this *isn't
		// present* then call the next handler in the chain as normal.
		exists := app.Session.Exists(r, KeySessionTokenMessage)
		if !exists {
			next.ServeHTTP(rw, r)
			return
		}
		tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
		if !ok {
			app.serverError(rw, fmt.Errorf("authenticate: no user available on session"))
			next.ServeHTTP(rw, r)
			return
		}

		// verify expiration date of token
		tokenClaims, err := models.GetClaimsFromToken(&tokenMsg.Token)
		if err != nil {
			app.serverError(rw, err)
			return
		}
		expiresAt := tokenClaims.StandardClaims.ExpiresAt
		delta := time.Unix(time.Now().Unix(), 0).Sub(time.Unix(expiresAt, 0))
		if delta > 0 {
			app.Session.Remove(r, KeySessionTokenMessage)
			app.Session.Put(r, KeySessionFlash, "You've been logged out - authentication expired, please logon!")
			next.ServeHTTP(rw, r)
			return
		}

		/*
			// Fetch the details of the current user from the api database. If no matching
			// record is found, or the current user is has been deactivated, remove the
			// (invalid) user value from their session and call the next
			// handler in the chain as normal.
			user, err := app.Users.Get(tokenMsg.Token, tokenMsg.User.ID)
			if  errors.Is(err, models.ErrNoRecord) ||
				errors.Is(err, models.ErrUnauthorizedToken) ||
				errors.Is(err, models.ErrForbiddenToken) ||
				(user!=nil && !user.Active) {
				app.Session.Remove(r, KeySessionTokenMessage)
				app.Session.Put(r, KeySessionFlash, "You've been logged out - user not found or invalid credentials!")
				next.ServeHTTP(rw, r)
				return
			} else if err != nil {
				app.serverError(rw, err)
				return
			}
		*/
		// Otherwise, we know that the request is coming from an authenticated user, with valid token.
		// We create a new copy of the request, with a true boolean value
		// added to the request context to indicate this,
		// and call the next handler in the chain *using this new copy of the request*.
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}
