package database

import (
	"database/sql"
	"fmt"
	"log"

	"saml-poc/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail retrieves a user by email address
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, first_name, last_name, is_active, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	err := r.db.conn.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// Create creates a new user in the database
func (r *UserRepository) Create(email, firstName, lastName string, isActive bool) (*models.User, error) {
	query := `
		INSERT INTO users (email, first_name, last_name, is_active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, NOW(), NOW()) 
		RETURNING id, email, first_name, last_name, is_active, created_at, updated_at
	`

	user := &models.User{}
	err := r.db.conn.QueryRow(query, email, firstName, lastName, isActive).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Successfully created new user: %s (%s %s)", user.Email, user.FirstName, user.LastName)
	return user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET first_name = $2, last_name = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.conn.Exec(query, user.ID, user.FirstName, user.LastName, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete soft deletes a user (sets is_active to false)
func (r *UserRepository) Delete(id int) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`

	_, err := r.db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List returns all users with pagination
func (r *UserRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, first_name, last_name, is_active, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
