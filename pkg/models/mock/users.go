package mock

import (
	"github.com/vgraveto/snippets/pkg/models"
	"time"
)

var mockUser = &models.User{
	ID:      1,
	Name:    "Alice",
	Email:   "alice@example.com",
	Created: time.Now(),
	Active:  true,
}

type UserModel struct{}

func (m *UserModel) Authenticate(email, password string) (string, error) {
	tD := models.TokenData{
		TokenIssuerName: "Test Application",
		TokenValidTime:  1 * time.Hour,
		TokenSigningKey: "testKey",
	}
	tM := models.NewTokenModel(&tD)
	token, _ := tM.CreateToken(mockUser)
	switch email {
	case "alice@example.com":
		return token, nil
	default:
		return "", models.ErrInvalidCredentials
	}
}

func (m *UserModel) Insert(token, name, email, password string, roles []int) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Get(token string, id int) (*models.User, error) {
	switch id {
	case 1:
		return mockUser, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *UserModel) GetAll(string) ([]*models.User, error) {
	return nil, nil
}

func (m *UserModel) ChangePassword(string, int, string, string) error {
	return nil
}

func (m *UserModel) GetRoleTypes(string) ([]*models.RoleType, error) {
	return nil, nil
}

func (m *UserModel) GetRoles(string, int) (*[]string, error) {
	return nil, nil
}
