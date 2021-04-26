package dbapi

import (
	"bytes"
	"fmt"
	"github.com/vgraveto/snippets/pkg/models"
	"io/ioutil"
	"net/http"
	"regexp"
)

// UserModel define type which wraps a API middleware connection to the database
type UserModel struct {
	Db API
}

func NewUserModel(d *API) *UserModel {
	return &UserModel{Db: *d}
}

// Authenticate method to verify whether a user exists with the provided email address and password.
// This will return a JSON Web Token (JWT) for the relevant user if they do.
func (m *UserModel) Authenticate(email, password string) (token string, err error) {
	// build the request URL
	urlRequest := fmt.Sprintf("%s/users/login", m.Db.Url)
	// build de request body
	body := models.LoginUser{
		Username: email,
		Password: password,
	}
	var bd bytes.Buffer
	err = models.ToJSON(body, &bd)
	if err != nil {
		return "", fmt.Errorf("AuthenticateJWT: Serialization: %v", err)
	}
	resp, err := http.Post(urlRequest, "application/json; charset=utf-8", &bd)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("AuthenticateJWT: ReadAll: %v", err)
		}
		bodyString := string(bodyBytes)
		//log.Printf("UserModel: Status %d (%s): %s\n", resp.StatusCode, resp.Status, bodyString)
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusUnprocessableEntity {
			return "", models.ErrInvalidCredentials
		} else {
			return "", fmt.Errorf("AuthenticateJWT: StatusCode %d -(%s): %s",
				resp.StatusCode, resp.Status, bodyString)
		}
	}
	tokenMsg := &models.TokenMessage{}
	err = models.FromJSON(tokenMsg, resp.Body)
	if err != nil {
		return "", fmt.Errorf("AuthenticateJWT: Deserialization: %v", err)
	}
	return tokenMsg.Token, nil
}

var reDuplicatedEmail = regexp.MustCompile(`duplicate email`)

// Insert method used to add a new record to the users table.
func (m *UserModel) Insert(token, name, email, password string, roles []int) error {
	// build the request URL
	url := fmt.Sprintf("%s/users", m.Db.Url)
	// build de request body
	body := models.CreateUser{
		Name:     name,
		Email:    email,
		Password: password,
		Roles:    roles,
	}
	var bd bytes.Buffer
	err := models.ToJSON(body, &bd)
	if err != nil {
		return fmt.Errorf("UserModel: Insert: Serialization: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, &bd)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authentication", token)
	// execute the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		//log.Printf("UserModel: Status %d (%): %s\n",resp.StatusCode,resp.Status,bodyString)
		if resp.StatusCode == http.StatusBadRequest {
			duplicatedEmail := reDuplicatedEmail.FindAllString(bodyString, -1)
			if duplicatedEmail == nil {
				return models.ErrBadRequest
			}
			return models.ErrDuplicateEmail
		} else if resp.StatusCode == http.StatusUnauthorized {
			return models.ErrUnauthorizedToken
		} else if resp.StatusCode == http.StatusForbidden {
			return models.ErrForbiddenToken
		} else if resp.StatusCode == http.StatusUnprocessableEntity {
			return models.ErrValidation
		} else {
			return fmt.Errorf("UserModel: Insert: StatusCode %d (%s): %s",
				resp.StatusCode, resp.Status, bodyString)
		}
	}
	/*
		// retrieve the message data from response body
		msg := &models.GenericMessage{}
		err = models.FromJSON(msg, resp.Body)
		if err!=nil {
			return err
		}
		log.Printf("UserModel: Insert: %q\n", msg.Message)
	*/
	return nil
}

// GetAll will return all the created users.
func (m *UserModel) GetAll(token string) ([]*models.User, error) {
	// build the request URL
	url := fmt.Sprintf("%s/users", m.Db.Url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authentication", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		bodyString := string(bodyBytes)
		// log.Printf("UserModel: Status %d (%): %s\n", resp.StatusCode, resp.Status, bodyString)
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, models.ErrUnauthorizedToken
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, models.ErrForbiddenToken
		} else {
			return nil, fmt.Errorf("UserModel: GetAll: Status: %s: %s",
				resp.Status, bodyString)
		}
	}

	// retrive the snippets data from response body
	var users []*models.User
	err = models.FromJSON(&users, resp.Body)
	if err != nil {
		//		log.Printf("SnippetUnauthenticatedModel: Deserializing: %v\n",err)
		return nil, err
	}
	return users, nil
}

// Get method used to fetch details for a specific user based on their user ID.
func (m *UserModel) Get(token string, id int) (*models.User, error) {
	// build the request URL
	url := fmt.Sprintf("%s/users/%d", m.Db.Url, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authentication", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		bodyString := string(bodyBytes)
		// log.Printf("UserModel: Status %d (%): %s\n", resp.StatusCode, resp.Status, bodyString)
		if resp.StatusCode == http.StatusNotFound {
			return nil, models.ErrNoRecord
		} else if resp.StatusCode == http.StatusUnauthorized {
			return nil, models.ErrUnauthorizedToken
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, models.ErrForbiddenToken
		} else {
			return nil, fmt.Errorf("UserModel: Get: StatusCode %d (%s): %s",
				resp.StatusCode, resp.Status, bodyString)
		}
	}

	// retrieve the user data from response body
	u := &models.User{}
	err = models.FromJSON(u, resp.Body)
	if err != nil {
		return nil, err
	}
	return u, nil
}

var reInvalidCredentials = regexp.MustCompile(`invalid credentials`)

// ChangePassword given the user ID, the current and the new passwords
// Verify current password to allow password change
func (m *UserModel) ChangePassword(token string, id int, currentPassword, newPassword string) error {
	// build the request URL
	url := fmt.Sprintf("%s/users/%d/change-password", m.Db.Url, id)
	// build de request body
	body := models.ChangeUserPassword{
		OldPassword: currentPassword,
		NewPassword: newPassword,
	}
	var bd bytes.Buffer
	err := models.ToJSON(body, &bd)
	if err != nil {
		return fmt.Errorf("UserModel: ChangePassword: Serialization: %v", err)
	}
	req, err := http.NewRequest(http.MethodPut, url, &bd)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authentication", token)
	// execute the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		//log.Printf("UserModel: Status %d (%): %s\n",resp.StatusCode,resp.Status,bodyString)
		if resp.StatusCode == http.StatusBadRequest {
			invalidCredentials := reInvalidCredentials.FindAllString(bodyString, -1)
			if invalidCredentials == nil {
				return models.ErrBadRequest
			}
			return models.ErrInvalidCredentials
		} else if resp.StatusCode == http.StatusUnauthorized {
			return models.ErrUnauthorizedToken
		} else if resp.StatusCode == http.StatusForbidden {
			return models.ErrForbiddenToken
		} else if resp.StatusCode == http.StatusUnprocessableEntity {
			return models.ErrValidation
		} else {
			return fmt.Errorf("UserModel: ChangePassword: StatusCode %d (%s): %s",
				resp.StatusCode, resp.Status, bodyString)
		}
	}
	return nil
}

// GetRoleTypes retrieves the existing role types from the database
func (m *UserModel) GetRoleTypes(token string) ([]*models.RoleType, error) {
	// build the request URL
	url := fmt.Sprintf("%s/users/role-types", m.Db.Url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authentication", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		bodyString := string(bodyBytes)
		// log.Printf("UserModel: Status %d (%): %s\n", resp.StatusCode, resp.Status, bodyString)
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, models.ErrUnauthorizedToken
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, models.ErrForbiddenToken
		} else {
			return nil, fmt.Errorf("UserModel: GetRoleTypes: Status: %s: %s",
				resp.Status, bodyString)
		}
	}

	// retrive the snippets data from response body
	var roles []*models.RoleType
	err = models.FromJSON(&roles, resp.Body)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoles obtains the roles of the user with the given id
func (m *UserModel) GetRoles(token string, id int) (*[]string, error) {
	// TODO implement request
	return nil, fmt.Errorf("UserModel: not implemented")
}
