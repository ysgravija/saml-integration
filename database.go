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

// CreateUser creates a new user in the database
func (d *Database) CreateUser(email, firstName, lastName string, isActive bool) (*User, error) {
	query := `
		INSERT INTO users (email, first_name, last_name, is_active) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, email, first_name, last_name, is_active
	`

	user := &User{}
	err := d.db.QueryRow(query, email, firstName, lastName, isActive).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Successfully created new user: %s (%s %s)", user.Email, user.FirstName, user.LastName)
	return user, nil
}

// IsUserAuthorizedWithJIT checks if a user exists and is active, with JIT user creation support
func (d *Database) IsUserAuthorizedWithJIT(email, firstName, lastName string) (bool, *User, error) {
	// First, try to find existing user
	user, err := d.GetUserByEmail(email)
	if err != nil {
		return false, nil, err
	}

	// If user exists, check if they're active
	if user != nil {
		if !user.IsActive {
			log.Printf("User is inactive: %s", email)
			return false, user, nil
		}
		log.Printf("Existing user authorized: %s", email)
		return true, user, nil
	}

	// User doesn't exist - check if JIT is enabled
	if !EnableJIT {
		log.Printf("User not found and JIT is disabled: %s", email)
		return false, nil, nil
	}

	// JIT is enabled - validate required attributes
	if RequiredAttributesForJIT {
		if firstName == "" || lastName == "" {
			log.Printf("JIT creation failed - missing required attributes for user: %s (firstName: '%s', lastName: '%s')",
				email, firstName, lastName)
			return false, nil, fmt.Errorf("missing required attributes for JIT user creation")
		}
	}

	// Set default values for missing attributes
	if firstName == "" {
		firstName = "Unknown"
	}
	if lastName == "" {
		lastName = "User"
	}

	// Create new user via JIT
	log.Printf("Creating new user via JIT: %s (%s %s)", email, firstName, lastName)
	newUser, err := d.CreateUser(email, firstName, lastName, DefaultUserActive)
	if err != nil {
		log.Printf("JIT user creation failed for %s: %v", email, err)
		return false, nil, fmt.Errorf("JIT user creation failed: %w", err)
	}

	log.Printf("JIT user creation successful: %s", email)
	return true, newUser, nil
}
