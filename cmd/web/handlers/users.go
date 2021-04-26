package handlers

import (
	"errors"
	"fmt"
	"github.com/vgraveto/snippets/pkg/forms"
	"github.com/vgraveto/snippets/pkg/models"
	"net/http"
)

func (app *Application) createUserForm(rw http.ResponseWriter, r *http.Request) {

	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		// Required logged in user, redirect to login page
		app.Session.Put(r, KeySessionRedirectPath, r.URL.Path)
		app.Session.Put(r, KeySessionFlash, "Operation not allowed. Please log in")
		http.Redirect(rw, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Get possible user role types from database
	roles, err := app.Users.GetRoleTypes(tokenMsg.Token)
	if err != nil {
		app.serverError(rw, err)
		return
	}
	app.render(rw, r, "signup.page.tmpl",
		&TemplateData{
			Form:  forms.New(nil),
			Roles: roles,
		})
}

func (app *Application) createUser(rw http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		app.clientError(rw, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		app.render(rw, r, "signup.page.tmpl", &TemplateData{Form: form})
		return
	}

	if app.DebugOn {
		app.ErrorLog.Printf("createUser: roles: %#v - %#v\n", form.GetString("roles"), form.GetInt("roles"))
	}

	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		// Required logged in user, redirect to login page
		app.Session.Put(r, KeySessionRedirectPath, r.URL.Path)
		app.Session.Put(r, KeySessionFlash, "Operation not allowed. Please log in")
		http.Redirect(rw, r, "/user/login", http.StatusSeeOther)
		//		app.serverError(rw, fmt.Errorf("createUser: no user available on session"))
		return
	}

	// Try to create a new user record in the database. If the email already exists
	// add an error message to the form and re-display it.
	err = app.Users.Insert(tokenMsg.Token,
		form.Get("name"),
		form.Get("email"),
		form.Get("password"),
		form.GetInt("roles"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(rw, r, "signup.page.tmpl", &TemplateData{Form: form})
		} else if errors.Is(err, models.ErrUnauthorizedToken) || errors.Is(err, models.ErrForbiddenToken) {
			app.Session.Put(r, KeySessionFlash, "Operation not allowed by this user")
			http.Redirect(rw, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(rw, err)
		}
		return
	}

	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked and asking them to log in.
	app.Session.Put(r, KeySessionFlash, "Your signup was successful.")

	// And redirect the user to the users list page.
	http.Redirect(rw, r, "/users", http.StatusSeeOther)
}

func (app *Application) loginUserForm(rw http.ResponseWriter, r *http.Request) {
	app.render(rw, r, "login.page.tmpl",
		&TemplateData{
			Form: forms.New(nil),
		})
}

func (app *Application) loginUser(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(rw, http.StatusBadRequest)
		return
	}
	// Check whether the credentials are valid. If they're not, add a generic error
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)

	token, err := app.Users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(rw, r, "login.page.tmpl", &TemplateData{Form: form})
		} else {
			app.serverError(rw, err)
		}
		return
	}

	// Add new tokenUser to the session, so that they are now 'logged in'.
	tu, err := models.GetUserFromToken(&token)
	if err != nil {
		app.serverError(rw, err)
		return
	}
	tm := models.TokenMessage{
		User:  *tu,
		Token: token,
	}
	app.Session.Put(r, KeySessionTokenMessage, tm)

	app.Session.Put(r, KeySessionFlash, "You've been logged in successfully!")
	path := app.Session.PopString(r, KeySessionRedirectPath)
	if path != "" {
		http.Redirect(rw, r, path, http.StatusSeeOther)
		return
	}

	// Redirect the user to the list snippets page.
	http.Redirect(rw, r, "/snippets", http.StatusSeeOther)
}

func (app *Application) logoutUser(rw http.ResponseWriter, r *http.Request) {
	// Remove the authenticatedUser from the session data so that the user is 'logged out'.
	app.Session.Remove(r, KeySessionTokenMessage)

	// Add a flash message to the session to confirm to the user that they've been logged out.
	app.Session.Put(r, KeySessionFlash, "You've been logged out successfully!")

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

func (app *Application) listUsers(rw http.ResponseWriter, r *http.Request) {

	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		app.serverError(rw, fmt.Errorf("listUsers: no user available on session"))
		return
	}

	u, err := app.Users.GetAll(tokenMsg.Token)
	if err != nil {
		if errors.Is(err, models.ErrUnauthorizedToken) || errors.Is(err, models.ErrForbiddenToken) {
			if app.DebugOn {
				app.ErrorLog.Printf("listUsers: %v\n", err)
			}
			app.Session.Put(r, KeySessionFlash, "Operation not allowed by this user")
			http.Redirect(rw, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(rw, err)
		}
		return
	}

	// Use the new render helper.
	app.render(rw, r, "users.page.tmpl",
		&TemplateData{
			Users: u,
			ID:    tokenMsg.User.ID})
}

func (app *Application) userGet(rw http.ResponseWriter, r *http.Request) {
	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		app.serverError(rw, fmt.Errorf("userGet: no user available on session"))
		return
	}

	// get ID from the URL
	id, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("userGet: user %d:  %v\n", id, err)
		app.serverError(rw, err)
		return
	}

	user, err := app.Users.Get(tokenMsg.Token, id)
	if err != nil {
		if errors.Is(err, models.ErrUnauthorizedToken) || errors.Is(err, models.ErrForbiddenToken) {
			if app.DebugOn {
				app.ErrorLog.Printf("listUsers: %v\n", err)
			}
			app.Session.Put(r, KeySessionFlash, "Operation not allowed by this user")
			http.Redirect(rw, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(rw, err)
		}
		return
	}

	app.render(rw, r, "user.page.tmpl",
		&TemplateData{
			User: user})
}

func (app *Application) userProfile(rw http.ResponseWriter, r *http.Request) {
	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		app.serverError(rw, fmt.Errorf("userProfile: no user available on session"))
		return
	}

	user, err := app.Users.Get(tokenMsg.Token, tokenMsg.User.ID)
	if err != nil {
		app.serverError(rw, err)
		return
	}

	app.render(rw, r, "profile.page.tmpl",
		&TemplateData{
			User: user})
}

func (app *Application) changePasswordForm(rw http.ResponseWriter, r *http.Request) {
	app.render(rw, r, "password.page.tmpl", &TemplateData{
		Form: forms.New(nil),
	})
}

func (app *Application) changePassword(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(rw, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper
	form := forms.New(r.PostForm)
	form.Required("currentPassword", "newPassword", "newPasswordConfirmation")
	form.MinLength("newPassword", 10)
	if form.Get("newPassword") != form.Get("newPasswordConfirmation") {
		form.Errors.Add("newPasswordConfirmation", "Passwords do not match")
	}
	if !form.Valid() {
		app.render(rw, r, "password.page.tmpl", &TemplateData{Form: form})
		return
	}

	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		app.serverError(rw, fmt.Errorf("currentPassword: no user available on session"))
		return
	}

	err = app.Users.ChangePassword(tokenMsg.Token, tokenMsg.User.ID, form.Get("currentPassword"), form.Get("newPassword"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("currentPassword", "Current password is not valid")
			app.render(rw, r, "password.page.tmpl", &TemplateData{Form: form})
		} else if errors.Is(err, models.ErrValidation) {
			form.Errors.Add("generic", "bad request invalid data provided")
			app.render(rw, r, "password.page.tmpl", &TemplateData{Form: form})
		} else if err != nil {
			app.serverError(rw, err)
		}
		return
	}

	app.Session.Put(r, KeySessionFlash, "Your password has been updated!")
	http.Redirect(rw, r, "/user/profile", http.StatusSeeOther)
}

func (app *Application) resetPasswordForm(rw http.ResponseWriter, r *http.Request) {
	// get ID from the URL
	id, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("resetPasswordForm: user %d:  %v\n", id, err)
		app.serverError(rw, err)
		return
	}

	app.render(rw, r, "resetPassword.page.tmpl", &TemplateData{
		Form: forms.New(nil),
		ID:   id,
	})
}

func (app *Application) resetPassword(rw http.ResponseWriter, r *http.Request) {
	// get ID from the URL
	id, err := getID(r)
	if err != nil {
		// should never happen as router blocks invalid URL request
		app.ErrorLog.Printf("resetPassword: user %d:  %v\n", id, err)
		app.serverError(rw, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(rw, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper
	form := forms.New(r.PostForm)
	form.Required("newPassword", "newPasswordConfirmation")
	form.MinLength("newPassword", 10)
	if form.Get("newPassword") != form.Get("newPasswordConfirmation") {
		form.Errors.Add("newPasswordConfirmation", "Passwords do not match")
	}
	if !form.Valid() {
		app.render(rw, r, "resetPassword.page.tmpl",
			&TemplateData{
				Form: form,
				ID:   id})
		return
	}

	tokenMsg, ok := app.Session.Get(r, KeySessionTokenMessage).(models.TokenMessage)
	if !ok {
		app.serverError(rw, fmt.Errorf("resetPassword: no user available on session"))
		return
	}

	// "0123498765" dummy oldpassword just to satisfy min length required
	err = app.Users.ChangePassword(tokenMsg.Token, id, "0123498765", form.Get("newPassword"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) || errors.Is(err, models.ErrValidation) {
			app.Session.Put(r, KeySessionFlash, "Operation not allowed by this user")
			http.Redirect(rw, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(rw, err)
		}
		return
	}

	app.Session.Put(r, KeySessionFlash, fmt.Sprintf("Password of user #%d has been updated!", id))
	http.Redirect(rw, r, fmt.Sprintf("/user/%d", id), http.StatusSeeOther)
}
