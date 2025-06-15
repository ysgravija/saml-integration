package handlers

import (
	"fmt"
	"net/http"

	"github.com/crewjam/saml/samlsp"

	"saml-poc/internal/saml"
)

// HomeHandler handles the home page
type HomeHandler struct{}

// NewHomeHandler creates a new home handler
func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

// ServeHTTP handles the home page request
func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get SAML session from context
	session := samlsp.SessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "No SAML session found", http.StatusUnauthorized)
		return
	}

	// Extract user attributes
	attrs := saml.ExtractUserAttributes(session, r)

	// Generate HTML response
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>SAML SSO - Home</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 800px; 
            margin: 50px auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .user-info {
            background: #ecf0f1;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .success {
            color: #27ae60;
            font-weight: bold;
        }
        .attribute {
            margin: 5px 0;
            padding: 5px 0;
        }
        .label {
            font-weight: bold;
            color: #34495e;
        }
        .value {
            color: #2c3e50;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="header">SAML SSO Authentication Successful</h1>
        
        <div class="success">
            Welcome! You have been successfully authenticated via SAML.
        </div>
        
        <div class="user-info">
            <h3>User Information:</h3>
            <div class="attribute">
                <span class="label">Email:</span> 
                <span class="value">%s</span>
            </div>
            <div class="attribute">
                <span class="label">First Name:</span> 
                <span class="value">%s</span>
            </div>
            <div class="attribute">
                <span class="label">Last Name:</span> 
                <span class="value">%s</span>
            </div>
        </div>
        
        <p>This page is protected and can only be accessed after successful SAML authentication and database validation.</p>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #bdc3c7;">
            <a href="/debug" style="color: #3498db; text-decoration: none;">View Debug Information</a>
        </div>
    </div>
</body>
</html>
    `, attrs.Email, attrs.FirstName, attrs.LastName)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
