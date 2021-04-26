package handlers

import (
	"github.com/golangcollege/sessions"
	"github.com/vgraveto/snippets/pkg/models"
	"html/template"
	"log"
)

type contextKey string

const (
	contextKeyIsAuthenticated = contextKey("isAuthenticated")
)

const (
	KeySessionTokenMessage = "authenticatedTokenMessage"
	KeySessionFlash        = "flash"
	KeySessionRedirectPath = "redirectPathAfterLogin"
)

// Application Define a struct to hold the application-wide dependencies for the web application.
// This will allow us to make the objects available to our handlers.
type Application struct {
	DebugOn       bool
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	Session       *sessions.Session
	TemplateCache map[string]*template.Template
	Snippets      models.APISnippets
	Users         models.APIUsers
}
