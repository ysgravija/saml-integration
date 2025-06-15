package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
}

type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase() (*Database, error) {
	connStr := "host=localhost port=5432 user=saml_user password=saml_password dbname=saml_sso sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// GetUserByEmail retrieves a user by email address
func (d *Database) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, first_name, last_name, is_active 
		FROM users 
		WHERE email = $1
	`

	err := d.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// IsUserAuthorized checks if a user exists and is active
func (d *Database) IsUserAuthorized(email string) (bool, *User, error) {
	user, err := d.GetUserByEmail(email)
	if err != nil {
		return false, nil, err
	}

	if user == nil {
		log.Printf("User not found in database: %s", email)
		return false, nil, nil
	}

	if !user.IsActive {
		log.Printf("User is inactive: %s", email)
		return false, user, nil
	}

	log.Printf("User authorized: %s", email)
	return true, user, nil
}
