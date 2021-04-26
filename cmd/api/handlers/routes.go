package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
)

func (app *Application) Routes() http.Handler {

	mux := mux.NewRouter()
	// Wrap the existing chain with the logRequest middleware.
	mux.Use(app.recoverPanic, app.logRequest, secureHeaders)

	// GET handlers for API
	getR := mux.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/", home)
	getR.HandleFunc("/ping", ping)
	getR.HandleFunc("/snippets", app.listAllSnippets)
	getR.HandleFunc("/snippets/{id:[1-9][0-9]*}", app.getSimpleSnippet)
	getR.Handle("/users", AddMiddleware(http.HandlerFunc(app.listAllUsers),
		app.authorize("administrator"),
		app.authenticate))
	getR.Handle("/users/{id:[1-9][0-9]*}", AddMiddleware(http.HandlerFunc(app.getSimpleUser),
		app.authorize("self"),
		app.authenticate))
	getR.Handle("/users/role-types", AddMiddleware(http.HandlerFunc(app.listAllRoleTypes),
		app.authorize("administrator"),
		app.authenticate))

	// POST handlers for API
	postR := mux.Methods(http.MethodPost).Subrouter()
	postR.Handle("/snippets", AddMiddleware(http.HandlerFunc(app.createSnippet),
		app.ValidateJSONBody(&models.SnippetCreate{}, KeySnippetCreate{}),
		app.authorize("user"),
		app.authenticate))
	postR.Handle("/users/login", AddMiddleware(http.HandlerFunc(app.loginUser),
		app.ValidateJSONBody(&models.LoginUser{}, KeyLoginUser{})))
	postR.Handle("/users", AddMiddleware(http.HandlerFunc(app.createUser),
		app.ValidateJSONBody(&models.CreateUser{}, KeyCreateUser{}),
		app.authorize("administrator"),
		app.authenticate))

	// PUT handlers for API
	putR := mux.Methods(http.MethodPut).Subrouter()
	putR.Handle("/users/{id}/change-password", AddMiddleware(http.HandlerFunc(app.changeUserPassword),
		app.ValidateJSONBody(&models.ChangeUserPassword{}, KeyChangeUserPassword{}),
		app.authorize("self"),
		app.authenticate))

	// handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)

	getR.Handle("/docs", sh)
	getR.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	return mux
}
