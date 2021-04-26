package models

import (
	"fmt"
	"time"
)

type UnauthotizedUsers interface {
	Authenticate(string, string) (int, error)
}

type Users interface {
	UnauthotizedUsers
	Insert(string, string, string, []int) error
	Get(int) (*User, error)
	GetAll() ([]*User, error)
	ChangePassword(int, string, string) error
	ResetPassword(int, string) error
	GetRoleTypes() ([]*RoleType, error)
	GetRoles(int) (*[]string, error)
}

type APIUnauthotizedUsers interface {
	Authenticate(email, password string) (token string, err error)
}

type APIUsers interface {
	APIUnauthotizedUsers
	// the first string parameter is a valid token for the API
	Insert(string, string, string, string, []int) error
	Get(string, int) (*User, error)
	GetAll(string) ([]*User, error)
	ChangePassword(string, int, string, string) error
	GetRoleTypes(string) ([]*RoleType, error)
	GetRoles(string, int) (*[]string, error)
}

const (
	// AministratorRole - users with this role have granted permission
	AministratorRole = "administrator"
	// SelfRole is used to specify the permission where the current user is equal to the ID in the request
	SelfRole = "self"
)

// User defines the structure for an API user
// swagger:model
type User struct {
	// the id for the user
	//
	// required: false
	// min: 1
	ID int `json:"id" validate:"min=1"`

	// the name for this user
	//
	// required: true
	// max length: 255
	Name string `json:"name" validate:"required,max=255"`

	// the email for this user
	//
	// required: true
	// max length: 255
	Email string `json:"email" validate:"required,email,max=255"`

	// the authorization roles for this user
	//
	// required: false
	Roles []string `json:"roles"`

	HashedPassword []byte `json:"-"`

	// the created dateTime for this user
	//
	// required: false
	Created time.Time `json:"created"`

	// the active state of this user
	//
	// required: false
	Active bool `json:"active"`
}

// RoleType defines the structure for role types of an user in the API
// swagger:model
type RoleType struct {
	// the id for the role type
	//
	// required: false
	// min: 1
	ID int `json:"id" validate:"min=1"`

	// the role name
	//
	// required: true
	// max length: 45
	Role string `json:"role" validate:"required,max=45"`

	// the role description
	//
	// required: true
	// max length: 45
	Description string `json:"description" validate:"required,max=45"`

	// the created dateTime for this role type
	//
	// required: false
	Created time.Time `json:"created"`
}

// UserRoleDetail defines the structure for roles of an user
// swagger:model
type UserRoleDetail struct {
	// the id for the role detail
	//
	// required: false
	// min: 1
	ID int `json:"id"`

	// the id of the user assigned for this role detail
	//
	// required: false
	// min: 1
	IDUser int `json:"iduser" validate:"min=1"`

	// the if of the role specified role for this user
	//
	// required: true
	// max length: 45
	IDRole string `json:"idrole" validate:"required,max=45"`

	// the created dateTime for this role detail
	//
	// required: false
	Created time.Time `json:"created"`
}

// CheckPermission returns nil if the userRoles array has the permission or if it has AdministratorRole
func CheckUserPermission(userRoles *[]string, permission string) error {
	if userRoles == nil {
		return fmt.Errorf("checkPermission: No user roles supplied")
	}

	if permission == "" {
		return fmt.Errorf("checkPermission: You must supply a permission to check against.")
	}

	for _, r := range *userRoles {
		// Administrator role is always granted permission
		if r == AministratorRole || permission == r {
			return nil
		}
	}

	return fmt.Errorf("checkPermission: User not authorized")
}

// LoginUser defines the structure for login of an user
// swagger:model
type LoginUser struct {
	// the username for this user
	//
	// required: true
	// max length: 255
	Username string `json:"username" validate:"required,email,max=255"`
	// the username for this user
	//
	// required: true
	// max length: 60
	// min length: 10
	Password string `json:"password" validate:"required,min=10,max=60"`
}

// CreateUser defines the structure for creating an user
// swagger:model
type CreateUser struct {
	// the name for this user
	//
	// required: true
	// max length: 255
	// min length: 10
	Name string `json:"name" validate:"required,min=10,max=255"`
	// the username for this user
	//
	// required: true
	// max length: 255
	Email string `json:"email" validate:"required,email"`
	// the username for this user
	//
	// required: true
	// max length: 60
	// min length: 10
	Password string `json:"password" validate:"required,min=10,max=60"`
	// the int slice of role ID's for this user
	//
	// required: false
	Roles []int `json:"roles"`
}

// ChangeUserPassword defines the structure for change of an user password
// swagger:model
type ChangeUserPassword struct {
	// the old password for this user
	//
	// required: true
	// max length: 60
	// min length: 10
	OldPassword string `json:"oldPassword" validate:"required,min=10,max=60"`
	// the new password for this user
	//
	// required: true
	// max length: 60
	// min length: 10
	NewPassword string `json:"newPassword" validate:"required,min=10,max=60"`
}
