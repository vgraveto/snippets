package dbmysql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/vgraveto/snippets/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

// UserModel type which wraps a sql.DB connection pool.
type UserModel struct {
	db *sql.DB
}

// NewUserModel creates a new UserModel
func NewUserModel(d *sql.DB) *UserModel {
	return &UserModel{db: d}
}

// Insert method used to add a new record to the users table and its roles to useRoleDetails table
func (m *UserModel) Insert(name, email, password string, roles []int) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// begin a new transaction to impose that user is only inserted if everything is runs ok
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES(?, ?, ?, UTC_TIMESTAMP())`
	// Use the Exec() method to insert the user details and hashed password
	//into the users table.
	result, err := tx.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// If this returns an error, we use the errors.As() function to check
		// whether the error has the type *dbmysql.MySQLError. If it does, the
		// error will be assigned to the mySQLError variable. We can then check
		// whether or not the error relates to our users_uc_email key by
		// checking the contents of the message string. If it does, we return
		// an ErrDuplicateEmail error.
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				tx.Rollback()
				return models.ErrDuplicateEmail
			}
		}
		tx.Rollback()
		return err
	}

	idUser, _ := result.LastInsertId()
	ok := true
	if roles != nil {
		// insert the user roles in userRoledetails table
		stmt := `INSERT INTO userRolesDetails (iduser, idrole, created) VALUES(?, ?, UTC_TIMESTAMP())`
		for _, role := range roles {
			_, err = tx.Exec(stmt, idUser, role)
			if err != nil {
				ok = false
				break
			}
		}
	}
	if !ok {
		err1 := tx.Rollback()
		if err1 != nil {
			return fmt.Errorf("Insert: Rollback: %v: %v", err1, err)
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Insert: Commit: %v", err)
	}
	return nil
}

// Authenticate method to verify whether a user exists with the provided email address and password.
// This will return the relevant user ID if they do.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Retrieve the id and hashed password associated with the given email. If no
	// matching email exists, or the user is not active, we return the
	// ErrInvalidCredentials error.
	var id int
	var hashedPassword []byte
	stmt := "SELECT id, hashed_password FROM users WHERE email = ? AND active = TRUE"
	row := m.db.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Otherwise, the password is correct. Return the user ID.
	return id, nil
}

// GetAll will return all the created users.
func (m *UserModel) GetAll() ([]*models.User, error) {
	stmt := "SELECT id, name, email, created, active FROM users ORDER BY id DESC"
	rows, err := m.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		// Create a pointer to a new zeroed User struct.
		u := &models.User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
		if err != nil {
			return nil, err
		}
		// get user Roles
		uRoles, err := m.GetRoles(u.ID)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				// no roles defined for this user
				u.Roles = nil
			} else {
				return nil, err
			}
		}
		u.Roles = *uRoles

		// Append it to the slice of snippets.
		users = append(users, u)
	}
	// check for any errors on rows
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the users slice.
	return users, nil
}

// Get method used to fetch details for a specific user based on their user ID.
func (m *UserModel) Get(id int) (*models.User, error) {
	u := &models.User{}
	stmt := `SELECT id, name, email, created, active FROM users WHERE id = ?`
	err := m.db.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	// get user roles
	uRoles, err := m.GetRoles(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			// no roles defined for this user
			u.Roles = nil
		} else {
			return nil, err
		}
	}
	u.Roles = *uRoles

	return u, nil
}

// ChangePassword given the user ID, the current and the new passwords
// Verify current password to allow password change
func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte
	row := m.db.QueryRow("SELECT hashed_password FROM users WHERE id = ?", id)
	err := row.Scan(&currentHashedPassword)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return models.ErrInvalidCredentials
		} else {
			return err
		}
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	stmt := "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.db.Exec(stmt, string(newHashedPassword), id)
	return err
}

// ResetPassword given the user ID and the new passwords
// Only used for administrator purpose
func (m *UserModel) ResetPassword(id int, newPassword string) error {

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	stmt := "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.db.Exec(stmt, string(newHashedPassword), id)
	return err
}

// GetRoleTypes obtains the existing role types from the database
func (m *UserModel) GetRoleTypes() ([]*models.RoleType, error) {

	roles := []*models.RoleType{}
	stmt := "SELECT id, role, description, created FROM roleTypes"
	rows, err := m.db.Query(stmt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		rt := &models.RoleType{}
		err = rows.Scan(&rt.ID, &rt.Role, &rt.Description, &rt.Created)
		if err != nil {
			return nil, err
		}
		roles = append(roles, rt)
	}
	// check for any errors on rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK then return the roles slice.
	return roles, nil
}

// GetRoles obtains the roles of the user with the given id
func (m *UserModel) GetRoles(id int) (*[]string, error) {

	userRoles := []string{}
	stmt := "SELECT role FROM roleTypes,userRolesDetails WHERE roleTypes.id=userRolesDetails.idrole AND userRolesDetails.iduser=?;"
	rows, err := m.db.Query(stmt, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		err = rows.Scan(&role)
		if err != nil {
			return nil, err
		}
		userRoles = append(userRoles, role)
	}
	// check for any errors on rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK then return the userRoles slice.
	return &userRoles, nil
}
