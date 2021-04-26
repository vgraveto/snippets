package handlers

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/justinas/alice"
)

// Update the signature for the routes() method so that it returns a
// http.Handler instead of *http.ServeMux.
func (app *Application) Routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes. For now, this chain will only contain
	// the session middleware but we'll add more to it later.
	dynamicMiddleware := alice.New(app.Session.Enable, noSurf, app.authenticate)

	mux := mux.NewRouter()
	mux.Handle("/", dynamicMiddleware.ThenFunc(app.home)).Methods("GET")
	mux.Handle("/about", dynamicMiddleware.ThenFunc(app.about)).Methods("GET")
	mux.Handle("/snippets", dynamicMiddleware.ThenFunc(app.listSnippets)).Methods("GET")
	mux.Handle("/snippet/{id:[1-9][0-9]*}", dynamicMiddleware.ThenFunc(app.showSnippet)).Methods("GET")
	mux.Handle("/snippet/create",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet)).Methods("POST")
	mux.Handle("/snippet/create",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm)).Methods("GET")

	mux.Handle("/user/signup",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createUserForm)).Methods("GET")
	mux.Handle("/user/signup",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createUser)).Methods("POST")
	mux.Handle("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm)).Methods("GET")
	mux.Handle("/user/login", dynamicMiddleware.ThenFunc(app.loginUser)).Methods("POST")
	mux.Handle("/user/logout",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser)).Methods("POST")
	mux.Handle("/users",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.listUsers)).Methods("GET")
	mux.Handle("/user/{id:[1-9][0-9]*}",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.userGet)).Methods("GET")
	mux.Handle("/user/{id:[1-9][0-9]*}/reset-password",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.resetPasswordForm)).Methods("GET")
	mux.Handle("/user/{id:[1-9][0-9]*}/reset-password",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.resetPassword)).Methods("POST")
	mux.Handle("/user/profile",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.userProfile)).Methods("GET")
	mux.Handle("/user/change-password",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.changePasswordForm)).Methods("GET")
	mux.Handle("/user/change-password",
		dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.changePassword)).Methods("POST")

	// Add a new GET /ping route.
	mux.Handle("/ping", http.HandlerFunc(ping)).Methods("GET")

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	// Wrap the existing chain with the logRequest middleware.
	return standardMiddleware.Then(mux)
}
