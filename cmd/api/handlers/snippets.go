package handlers

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
)

// KeySnippet is a key used for the Snippet object in the context
type KeySnippetCreate struct{}

// swagger:route GET /snippets snippets listSnippets
// Return a list of snippets from the database
//
// responses:
//	200: snippetsResponse
//	500: messageResponse

// listAllSnippets handles GET requests and returns all current snippets
func (app *Application) listAllSnippets(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	sp, err := app.Snippets.Latest()
	if err != nil {
		app.ErrorLog.Printf("listAllSnippets: Unable to get snipplets  %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Unable to get snipplets list"}, rw)
		return
	}

	err = models.ToJSON(sp, rw)
	if err != nil {
		// we should never be here but log the error just incase
		app.ErrorLog.Printf("listAllSnippets: Unable to serializing snipplets  %v\n", err)
	}
}

// swagger:route GET /snippets/{id} snippets listSingleSnippet
// Return a single snippet from the database
//
// responses:
//	200: snippetResponse
//	400: messageResponse
//	404: messageResponse
//  500: messageResponse

// getSingleSnippet handles GET requests of {id} snipplet
func (app *Application) getSimpleSnippet(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	id, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("getSimpleSnippet: snippet %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusBadRequest)
		models.ToJSON(&models.GenericMessage{http.StatusText(http.StatusBadRequest)}, rw)
		return
	}

	sp, err := app.Snippets.Get(id)
	switch err {
	case nil:
		break
	case models.ErrNoRecord:
		app.ErrorLog.Printf("getSimpleSnippet: snipplet %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusNotFound)
		models.ToJSON(&models.GenericMessage{fmt.Sprintf("Unable to get snipplet %d", id)}, rw)
		return
	default:
		app.ErrorLog.Printf("getSimpleSnippet: snipplet %d:  %v\n", id, err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{fmt.Sprintf("Unable to get snipplet %d", id)}, rw)
		return
	}

	err = models.ToJSON(sp, rw)
	if err != nil {
		// we should never be here but log the error just incase
		app.ErrorLog.Printf("getSimpleSnippet: Unable to serializing snipplet %d:  %v\n", id, err)
	}
}

// swagger:route POST /snippets snippets createSnippet
// Create a new snippet
//
//	Security:
//  - snippetskey:
//
// responses:
//	200: snippetResponse
//	400: messageResponse
//	401: messageResponse
//	403: messageResponse
//  422: validationResponse
//	500: messageResponse

// Create handles POST requests to add new snippet
func (app *Application) createSnippet(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	// fetch the snippet from the context
	spc, ok := context.Get(r, KeySnippetCreate{}).(*models.SnippetCreate)
	if !ok {
		app.ErrorLog.Printf("createSnippet: No snippet in the context\n")
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem with snippet data"}, rw)
		return
	}

	if app.DebugOn {
		app.InfoLog.Printf("createSnippet: Inserting snippet: %#v\n", spc)
	}

	id, err := app.Snippets.Insert(spc.Title, spc.Content, spc.Expires)
	if err != nil {
		app.ErrorLog.Printf("createSnippet: inserting: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem inserting snippet data"}, rw)
		return
	}
	sp, err := app.Snippets.Get(id)
	if err != nil {
		app.ErrorLog.Printf("createSnippet: geting: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		models.ToJSON(&models.GenericMessage{"Problem geting snippet data"}, rw)
		return
	}

	models.ToJSON(sp, rw)
}
