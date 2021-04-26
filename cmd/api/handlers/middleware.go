package handlers

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(w, r)
	})
}

func (app *Application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.InfoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			//panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				//Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ValidateJSONBody provides JSON body read and validation middleware for handlers
// reads the JSON body to a given data object and validates its content
// adds that data to the request context when valid and calls the next handler
// otherwise aborts the request returning the response message with information
func (app *Application) ValidateJSONBody(dataObj, contextKey interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Add("Content-Type", "application/json")

			// verify that parameters are not nil that represents a misuse use of this middleware
			if dataObj == nil || contextKey == nil {
				app.ErrorLog.Printf("ValidateJSONBody: Invalid parameters: dataObj - %#v contextKey - %#v\n", dataObj, contextKey)

				rw.WriteHeader(http.StatusInternalServerError)
				models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusInternalServerError)}, rw)
				return
			}
			if app.DebugOn {
				app.InfoLog.Printf("ValidateJSONBody: dataObj - %#v contextKey - %#v\n", dataObj, contextKey)
			}

			err := models.FromJSON(dataObj, r.Body)
			if err != nil {
				app.ErrorLog.Printf("ValidateJSONBody: Deserializing error: %v\n", err)

				rw.WriteHeader(http.StatusBadRequest)
				models.ToJSON(&models.GenericMessage{fmt.Sprintf("Deserializing error: %v", err)}, rw)
				return
			}

			// validate the product
			errs := app.Val.Validate(dataObj)
			if len(errs) != 0 {
				app.ErrorLog.Printf("ValidateJSONBody: Validating: %v\n", errs)

				// return the validation messages as an array
				rw.WriteHeader(http.StatusUnprocessableEntity)
				models.ToJSON(&models.ValidationMessagesError{errs.Errors()}, rw)
				return
			}

			// add the LoginUser to the context
			context.Set(r, contextKey, dataObj)

			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(rw, r)
		})
	}
}
