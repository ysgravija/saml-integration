package saml

import (
	"net/http"

	"github.com/crewjam/saml/samlsp"
)

// UserAttributes represents extracted user attributes from SAML
type UserAttributes struct {
	Email     string
	FirstName string
	LastName  string
}

// ExtractUserAttributes extracts user attributes from SAML session
func ExtractUserAttributes(session samlsp.Session, r *http.Request) UserAttributes {
	attrs := UserAttributes{}

	// Try to get attributes from session
	if sessionWithAttrs, ok := session.(samlsp.SessionWithAttributes); ok {
		samlAttrs := sessionWithAttrs.GetAttributes()

		// Extract email from various possible attribute names
		attrs.Email = extractAttribute(samlAttrs, []string{
			"email",
			"emailAddress",
			"mail",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		})

		// Extract first name from various possible attribute names
		attrs.FirstName = extractAttribute(samlAttrs, []string{
			"firstName",
			"givenName",
			"given_name",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
		})

		// Extract last name from various possible attribute names
		attrs.LastName = extractAttribute(samlAttrs, []string{
			"lastName",
			"surname",
			"sn",
			"family_name",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
		})
	}

	// If we couldn't get attributes from session attributes, try using AttributeFromContext helper
	if attrs.Email == "" {
		attrs.Email = extractAttributeFromContext(r, []string{
			"email",
			"emailAddress",
			"mail",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		})
	}

	if attrs.FirstName == "" {
		attrs.FirstName = extractAttributeFromContext(r, []string{
			"firstName",
			"givenName",
			"given_name",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
		})
	}

	if attrs.LastName == "" {
		attrs.LastName = extractAttributeFromContext(r, []string{
			"lastName",
			"surname",
			"sn",
			"family_name",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
		})
	}

	return attrs
}

// extractAttribute extracts an attribute value from SAML attributes by trying multiple attribute names
func extractAttribute(attrs samlsp.Attributes, names []string) string {
	for _, name := range names {
		if value := attrs.Get(name); value != "" {
			return value
		}
	}
	return ""
}

// extractAttributeFromContext extracts an attribute value from request context by trying multiple attribute names
func extractAttributeFromContext(r *http.Request, names []string) string {
	for _, name := range names {
		if value := samlsp.AttributeFromContext(r.Context(), name); value != "" {
			return value
		}
	}
	return ""
}
