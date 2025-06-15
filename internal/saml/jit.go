package saml

import (
	"fmt"
	"log"

	"saml-poc/internal/config"
	"saml-poc/internal/database"
	"saml-poc/internal/models"
)

// JITService handles Just-In-Time user creation
type JITService struct {
	userRepo *database.UserRepository
	config   *config.JITConfig
}

// NewJITService creates a new JIT service
func NewJITService(userRepo *database.UserRepository, jitConfig *config.JITConfig) *JITService {
	return &JITService{
		userRepo: userRepo,
		config:   jitConfig,
	}
}

// AuthorizeUserWithJIT checks if a user is authorized, creating them if JIT is enabled
func (j *JITService) AuthorizeUserWithJIT(attrs UserAttributes) (bool, *models.User, error) {
	// First, try to find existing user
	user, err := j.userRepo.GetByEmail(attrs.Email)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// If user exists, check if they're active
	if user != nil {
		if !user.IsAuthorized() {
			log.Printf("User is inactive: %s", attrs.Email)
			return false, user, nil
		}
		log.Printf("Existing user authorized: %s", attrs.Email)
		return true, user, nil
	}

	// User doesn't exist - check if JIT is enabled
	if !j.config.Enabled {
		log.Printf("User not found and JIT is disabled: %s", attrs.Email)
		return false, nil, nil
	}

	// JIT is enabled - validate required attributes
	if j.config.RequiredAttributesMode {
		if attrs.FirstName == "" || attrs.LastName == "" {
			log.Printf("JIT creation failed - missing required attributes for user: %s (firstName: '%s', lastName: '%s')",
				attrs.Email, attrs.FirstName, attrs.LastName)
			return false, nil, fmt.Errorf("missing required attributes for JIT user creation")
		}
	}

	// Set default values for missing attributes
	firstName := attrs.FirstName
	lastName := attrs.LastName
	if firstName == "" {
		firstName = "Unknown"
	}
	if lastName == "" {
		lastName = "User"
	}

	// Create new user via JIT
	log.Printf("Creating new user via JIT: %s (%s %s)", attrs.Email, firstName, lastName)
	newUser, err := j.userRepo.Create(attrs.Email, firstName, lastName, j.config.DefaultUserActive)
	if err != nil {
		log.Printf("JIT user creation failed for %s: %v", attrs.Email, err)
		return false, nil, fmt.Errorf("JIT user creation failed: %w", err)
	}

	log.Printf("JIT user creation successful: %s", attrs.Email)
	return true, newUser, nil
}
