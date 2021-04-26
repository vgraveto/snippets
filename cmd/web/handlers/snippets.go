package handlers

import (
	"errors"
	"fmt"
	"github.com/vgraveto/snippets/pkg/forms"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
)

func (app *Application) listSnippets(rw http.ResponseWriter, r *http.Request) {

	s, err := app.Snippets.Latest()
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(rw)
		} else {
			app.serverError(rw, err)
		}
		return
	}

	// Use the new render helper.
	app.render(rw, r, "snippets.page.tmpl",
		&TemplateData{Snippets: s})
}

func (app *Application) showSnippet(rw http.ResponseWriter, r *http.Request) {
	// get ID from the URL
	idValue, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("showSnippet: user %d:  %v\n", idValue, err)
		app.serverError(rw, err)
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a
	// specific record based on its ID. If no matching record is found,
	// return a 404 Not Found response.
	s, err := app.Snippets.Get(idValue)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(rw)
		} else {
			app.serverError(rw, err)
		}
		return
	}

	// Use the new render helper.
	app.render(rw, r, "show.page.tmpl",
		&TemplateData{
			Snippet: s,
		})
}

// Add a new createSnippetForm handler, which for now returns a placeholder response.
func (app *Application) createSnippetForm(rw http.ResponseWriter, r *http.Request) {
	app.ErrorLog.Println("createSnippetForm")
	app.render(rw, r, "create.page.tmpl", &TemplateData{
		// Pass a new empty forms.Form object to the template.
		Form: forms.New(nil)})
}

func (app *Application) createSnippet(rw http.ResponseWriter, r *http.Request) {
	app.ErrorLog.Println("createSnippet")

	// First we call r.ParseForm() which adds any data in POST request bodies
	// to the r.PostForm map. This also works in the same way for PUT and PATCH
	// requests. If there are any errors, we use our app.ClientError helper to send
	// a 400 Bad Request response to the user.
	err := r.ParseForm()
	if err != nil {
		app.clientError(rw, http.StatusBadRequest)
		return
	}

	// Create a new forms.Form struct containing the POSTed data from the
	// form, then use the validation methods to check the content.
	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	// If the form isn't valid, redisplay the template passing in the
	// form.Form object as the data.
	if !form.Valid() {
		app.render(rw, r, "create.page.tmpl", &TemplateData{Form: form})
		return
	}
	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		app.serverError(rw, fmt.Errorf("createSnippet: no user available on session"))
		return
	}

	// Because the form data (with type url.Values) has been anonymously embedded
	// in the form.Form struct, we can use the Get() method to retrieve
	// the validated value for a particular form field.
	id, err := app.Snippets.Insert(tokenMsg.Token, form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(rw, err)
		return
	}

	// Use the Put() method to add a string value ("Your snippet was saved
	// successfully!") and the corresponding key ("flash") to the session
	// data. Note that if there's no existing session for the current user
	// (or their session has expired) then a new, empty, session for them
	// will automatically be created by the session middleware.
	app.Session.Put(r, KeySessionFlash, "Snippet successfully created!")

	http.Redirect(rw, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}
