package models

import "time"

type UnauthotizedSnippets interface {
	Get(int) (*Snippet, error)
	Latest() ([]*Snippet, error)
}

type Snippets interface {
	UnauthotizedSnippets
	Insert(string, string, string) (int, error)
}

type APISnippets interface {
	UnauthotizedSnippets
	Insert(string, string, string, string) (int, error)
}

// Snippet defines the structure for an API snippet
// swagger:model
type Snippet struct {
	// the id for the snippet
	//
	// required: false
	// min: 1
	ID int `json:"id" validate:"min=1"`

	// the title for this snippet
	//
	// required: true
	// max length: 100
	Title string `json:"title" validate:"required,max=100"`

	// the content for this snippet
	//
	// required: true
	// max length: 1000
	Content string `json:"content" validate:"required,max=1000"`

	// the created dateTime for this snippet
	//
	// required: false
	Created time.Time `json:"created"`

	// the expiration dateTime for this snippet
	//
	// required: false
	Expires time.Time `json:"expires"`
}

// SnippetCreate defines the structure for snippet creation
// swagger:model
type SnippetCreate struct {
	// the title for this snippet
	//
	// required: true
	// max length: 100
	Title string `json:"title" validate:"required,max=100"`

	// the content for this snippet
	//
	// required: true
	// max length: 1000
	Content string `json:"content" validate:"required,max=1000"`

	// the expiration number of days for this snippet (valid values 365, 7 or 1 days)
	//
	// required: true
	Expires string `json:"expires" validate:"required,expires"`
}
