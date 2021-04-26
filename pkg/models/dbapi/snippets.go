package dbapi

import (
	"bytes"
	"fmt"
	"github.com/vgraveto/snippets/pkg/models"
	"io/ioutil"
	"log"
	"net/http"
)

// SnippetModel define type which wraps a API middleware connection to the database
type SnippetModel struct {
	Db API
}

func NewSnippetModel(d *API) *SnippetModel {
	return &SnippetModel{Db: *d}
}

// Get will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {

	// build the request URL
	url := fmt.Sprintf("%s/snippets/%d", m.Db.Url, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			//			log.Printf("SnippetUnauthenticatedModel: Error body: %v\n",err)
			return nil, err
		}
		bodyString := string(bodyBytes)
		log.Printf("SnippetModel: Status %d (%s): %s\n", resp.StatusCode, resp.Status, bodyString)
		if resp.StatusCode == http.StatusNotFound {
			return nil, models.ErrNoRecord
		} else {
			return nil, fmt.Errorf("SnippetModel: Get: StatusCode %d (%s): %s",
				resp.StatusCode, resp.Status, bodyString)
		}
	}

	// retrive the snippet data from response body
	s := &models.Snippet{}
	err = models.FromJSON(s, resp.Body)
	if err != nil {
		//		log.Printf("SnippetUnauthenticatedModel: Deserializing: %v\n",err)
		return nil, err
	}
	return s, nil
}

// Latest will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {

	// build the request URL
	url := fmt.Sprintf("%s/snippets", m.Db.Url)
	resp, err := http.Get(url)
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
		//log.Printf("SnippetModel: Status %d: %s\n", resp.StatusCode, bodyString)
		return nil, fmt.Errorf("SnippetModel: Get: Status: %s: %s", resp.Status, bodyString)
	}

	// retrive the snippets data from response body
	var snippets []*models.Snippet
	err = models.FromJSON(&snippets, resp.Body)
	if err != nil {
		//		log.Printf("SnippetUnauthenticatedModel: Deserializing: %v\n",err)
		return nil, err
	}
	return snippets, nil
}

// Insert will insert a new snippet into the database and return its id
func (m *SnippetModel) Insert(token, title, content, expires string) (int, error) {
	// build the request URL
	url := fmt.Sprintf("%s/snippets", m.Db.Url)
	// build de request body
	body := models.SnippetCreate{
		Title:   title,
		Content: content,
		Expires: expires,
	}
	var bd bytes.Buffer
	err := models.ToJSON(body, &bd)
	if err != nil {
		return -1, fmt.Errorf("SnippetModel: Insert: Serialization: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, &bd)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authentication", token)
	// execute the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, err
		}
		bodyString := string(bodyBytes)
		log.Printf("SnippetModel: Status %d (%s): %s\n", resp.StatusCode, resp.Status, bodyString)
		if resp.StatusCode == http.StatusBadRequest {
			return -1, models.ErrBadRequest
		} else if resp.StatusCode == http.StatusUnauthorized {
			return -1, models.ErrUnauthorizedToken
		} else if resp.StatusCode == http.StatusForbidden {
			return -1, models.ErrForbiddenToken
		} else if resp.StatusCode == http.StatusUnprocessableEntity {
			return -1, models.ErrValidation
		} else {
			return -1, fmt.Errorf("SnippetModel: Insert: StatusCode %d (%s): %s",
				resp.StatusCode, resp.Status, bodyString)
		}
	}

	// retrieve the snippet data from response body
	s := &models.Snippet{}
	err = models.FromJSON(s, resp.Body)
	if err != nil {
		return -1, err
	}
	return s.ID, nil
}
