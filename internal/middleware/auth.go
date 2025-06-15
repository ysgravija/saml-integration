package middleware

import (
	"log"
	"net/http"

	"github.com/crewjam/saml/samlsp"

	"saml-poc/internal/saml"
)

// AuthMiddleware handles SAML authentication and user validation
type AuthMiddleware struct {
	jitService *saml.JITService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jitService *saml.JITService) *AuthMiddleware {
	return &AuthMiddleware{
		jitService: jitService,
	}
}

// DatabaseValidation validates users against the database after SAML authentication
func (m *AuthMiddleware) DatabaseValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get SAML session from context (already authenticated by SAML)
		session := samlsp.SessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "No SAML session found", http.StatusUnauthorized)
			return
		}

		// Extract user attributes from SAML session
		attrs := saml.ExtractUserAttributes(session, r)
		if attrs.Email == "" {
			log.Println("No email found in SAML session")
			http.Error(w, "No email found in SAML session", http.StatusBadRequest)
			return
		}

		log.Printf("Validating user from SAML session: %s (firstName: '%s', lastName: '%s')",
			attrs.Email, attrs.FirstName, attrs.LastName)

		// Validate user against database with JIT support
		authorized, user, err := m.jitService.AuthorizeUserWithJIT(attrs)
		if err != nil {
			log.Printf("Database error during user validation: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !authorized {
			log.Printf("User not authorized: %s", attrs.Email)
			http.Error(w, "Access denied: User not authorized for this application", http.StatusForbidden)
			return
		}

		log.Printf("User successfully validated: %s (%s %s)", user.Email, user.FirstName, user.LastName)

		// User is authorized, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
