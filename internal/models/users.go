package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	IsActive       bool
	Features       any
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	AddFeatures(id int, features []string) error
	PasswordUpdate(id int, currentPassword, newPassword string) error
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, isActive, created) 
	VALUES(?, ?, ?, false, UTC_TIMESTAMP()) `

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (*User, error) {
	var u = &User{}

	stmt := `SELECT id, hashed_password, isActive FROM users WHERE email = ?`
	err := m.DB.QueryRow(stmt, email).Scan(&u.ID, &u.HashedPassword, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	err = bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	return u, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := `SELECT EXISTS(SELECT true FROM users WHERE id = ?)`
	err := m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) Get(id int) (*User, error) {
	var u = &User{}

	stmt := `SELECT name, email, hashed_password, isActive, features, created FROM users WHERE id = ?`
	row := m.DB.QueryRow(stmt, id)
	err := row.Scan(&u.Name, &u.Email, &u.HashedPassword, &u.IsActive, &u.Features, &u.Created)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	return u, nil
}

func (m *UserModel) AddFeatures(id int, features []string) error {
	stmt := "UPDATE users SET features = ? WHERE id = ?"

	featuresJSON, err := json.Marshal(features)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(stmt, featuresJSON, id)
	if err != nil {
		return err
	}

	query := `CREATE TABLE IF NOT EXISTS features (
		id INT AUTO_INCREMENT PRIMARY KEY,
		date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		loaves INT NOT NULL,`

	for i, feature := range features {
		query += fmt.Sprintf("feature%d_%s TINYINT(1) DEFAULT 0, ", i+1, feature)
	}

	// Remove trailing comma and space, then close the table definition
	query = query[:len(query)-2] + ");"

	_, err = m.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte

	stmt := "SELECT hashed_password FROM users WHERE id = ?"

	err := m.DB.QueryRow(stmt, id).Scan(&currentHashedPassword)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt = "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)
	return err
}
