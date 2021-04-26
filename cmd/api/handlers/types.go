package handlers

import (
	"github.com/vgraveto/snippets/pkg/models"
	"log"
)

type contextKey string

const (
	contextKeyIsAuthenticated = contextKey("isAuthenticated")
)

// Define an application struct to hold the application-wide dependencies for the
// API application. This will allow us to
// make the objects available to our handlers.
type Application struct {
	DebugOn  bool
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	Snippets models.Snippets
	Users    models.Users
	Tokens   models.Tokens
	Val      *models.Validation
}
