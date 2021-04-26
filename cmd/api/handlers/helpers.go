package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
	"runtime/debug"
	"strconv"
)

// swagger:route GET / global home
// Return the string "Snippets API"
//
//
//  Produces:
//  - text/plain
//
// responses:
//	200: stringResponse

// home handles GET requests and returns "Snippets API"
func home(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte("Snippets API"))
}

// swagger:route GET /ping global pingAPI
// Return the JSON message with string "OK"
//
// responses:
//	200: messageResponse

// ping handles GET requests and returns an "OK"
func ping(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	models.ToJSON(&models.GenericMessage{Message: "OK"}, rw)
}

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

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	app.clientError(w, http.StatusNotFound)
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

// AddMiddleware adds middleware to a Handler
func AddMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	//	log.Println("Add Middleware")
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}
