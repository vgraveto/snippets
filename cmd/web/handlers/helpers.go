package handlers

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/nosurf"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

// The serverError helper writes an error message and stack trace to the errorLog,
//then sends a generic 500 Internal Server Error response to the user.
func (app *Application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Output(2, trace)

	if app.DebugOn {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
//to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *Application) clientError(w http.ResponseWriter, status int) {
	// VGG added infoLog to console
	app.InfoLog.Printf("clientError: %s\n", http.StatusText(status))
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
//the user.
func (app *Application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

//
// addDefaultData helper takes a pointer to a templateData struct,
// adds some data like current year, etc, and then returns the pointer.
func (app *Application) addDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}

	// Add the CSRF token to the templateData struct.
	td.CSRFToken = nosurf.Token(r)
	td.CurrentYear = time.Now().Year()
	// Add the flash message to the template data, if one exists.
	td.Flash = app.Session.PopString(r, KeySessionFlash)
	// Add the authentication status to the template data.
	td.IsAuthenticated = app.isAuthenticated(r)
	if td.IsAuthenticated {
		tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
		if !ok {
			app.ErrorLog.Printf("addDefaultData: no user available on session")
		} else {
			td.LoggedInName = tokenMsg.User.Name
			td.IsAdmin = tokenMsg.User.IsAdmin()
		}
	}

	return td
}

func (app *Application) render(w http.ResponseWriter, r *http.Request, name string, td *TemplateData) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := app.TemplateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError helper and then
	// return.
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write the contents of the buffer to the http.ResponseWriter. Again, this
	// is another time where we pass our http.ResponseWriter to a function that
	//takes an io.Writer.
	buf.WriteTo(w)
}

// Return true if the current request is from authenticated user, otherwise return false.
func (app *Application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

// getID returns the ID from the URL
func getID(r *http.Request) (int, error) {
	// parse the id from the url
	vars := mux.Vars(r)

	// convert the id into an integer and return
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		// should never happen
		return -1, err
	}

	return id, nil
}
