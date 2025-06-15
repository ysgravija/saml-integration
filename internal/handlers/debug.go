package handlers

import (
	"fmt"
	"net/http"

	"saml-poc/internal/config"
)

// DebugHandler handles debug information display
type DebugHandler struct {
	config *config.Config
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler(cfg *config.Config) *DebugHandler {
	return &DebugHandler{
		config: cfg,
	}
}

// ServeHTTP handles the debug page request
func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle POST request to clear cookies
	if r.Method == "POST" && r.FormValue("action") == "clear_cookies" {
		h.clearCookies(w, r)
		return
	}

	// Show debug page
	h.showDebugPage(w, r)
}

// clearCookies clears SAML session cookies
func (h *DebugHandler) clearCookies(w http.ResponseWriter, r *http.Request) {
	// Clear the SAML session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "saml_" + h.config.SAML.EntityID,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})

	// Also try to clear common SAML cookie names
	cookieNames := []string{
		"saml_session",
		"SAML_SESSION",
		"samlsession",
		"token",
	}

	for _, name := range cookieNames {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
		})
	}

	// Redirect back to debug page with success message
	http.Redirect(w, r, "/debug?cleared=true", http.StatusSeeOther)
}

// showDebugPage displays the debug information page
func (h *DebugHandler) showDebugPage(w http.ResponseWriter, r *http.Request) {
	// Check if cookies were just cleared
	clearedMessage := ""
	if r.URL.Query().Get("cleared") == "true" {
		clearedMessage = `
		<div class="success-message">
			Cookies cleared successfully! You can now test the SAML authentication flow.
		</div>`
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>SAML SSO - Debug Information</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 1000px; 
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
            border-bottom: 2px solid #e74c3c;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .section {
            margin: 20px 0;
            padding: 15px;
            background: #ecf0f1;
            border-radius: 5px;
        }
        .section h3 {
            margin-top: 0;
            color: #34495e;
        }
        .config-item {
            margin: 8px 0;
            padding: 5px 0;
        }
        .label {
            font-weight: bold;
            color: #2c3e50;
            display: inline-block;
            width: 200px;
        }
        .value {
            color: #27ae60;
        }
        .enabled {
            color: #27ae60;
            font-weight: bold;
        }
        .disabled {
            color: #e74c3c;
            font-weight: bold;
        }
        .warning {
            background: #f39c12;
            color: white;
            padding: 10px;
            border-radius: 5px;
            margin: 10px 0;
        }
        .success-message {
            background: #27ae60;
            color: white;
            padding: 10px;
            border-radius: 5px;
            margin: 10px 0;
        }
        .cookie-section {
            background: #3498db;
            color: white;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .cookie-section h3 {
            margin-top: 0;
            color: white;
        }
        .clear-button {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            margin-top: 10px;
        }
        .clear-button:hover {
            background: #c0392b;
        }
        .cookie-info {
            font-size: 14px;
            margin-top: 10px;
            opacity: 0.9;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="header">SAML SSO Debug Information</h1>
        
        <div class="warning">
            This page shows configuration details and should not be accessible in production.
        </div>
        
        %s
        
        <div class="cookie-section">
            <h3>Session Management</h3>
            <p>Clear SAML session cookies to test the authentication flow from the beginning.</p>
            <form method="POST" style="margin: 0;">
                <input type="hidden" name="action" value="clear_cookies">
                <button type="submit" class="clear-button">Clear Session Cookies</button>
            </form>
            <div class="cookie-info">
                This will clear SAML session cookies and redirect you back to this page.
                After clearing cookies, visiting the home page will trigger a new SAML authentication.
            </div>
        </div>
        
        <div class="section">
            <h3>Server Configuration</h3>
            <div class="config-item">
                <span class="label">Server Address:</span>
                <span class="value">%s</span>
            </div>
        </div>
        
        <div class="section">
            <h3>Database Configuration</h3>
            <div class="config-item">
                <span class="label">Host:</span>
                <span class="value">%s:%s</span>
            </div>
            <div class="config-item">
                <span class="label">Database:</span>
                <span class="value">%s</span>
            </div>
            <div class="config-item">
                <span class="label">User:</span>
                <span class="value">%s</span>
            </div>
        </div>
        
        <div class="section">
            <h3>SAML Configuration</h3>
            <div class="config-item">
                <span class="label">Entity ID:</span>
                <span class="value">%s</span>
            </div>
            <div class="config-item">
                <span class="label">ACS URL:</span>
                <span class="value">%s</span>
            </div>
            <div class="config-item">
                <span class="label">IdP Metadata Path:</span>
                <span class="value">%s</span>
            </div>
            <div class="config-item">
                <span class="label">Certificate File:</span>
                <span class="value">%s</span>
            </div>
            <div class="config-item">
                <span class="label">Key File:</span>
                <span class="value">%s</span>
            </div>
        </div>
        
        <div class="section">
            <h3>JIT (Just-In-Time) Configuration</h3>
            <div class="config-item">
                <span class="label">JIT Enabled:</span>
                <span class="%s">%s</span>
            </div>
            <div class="config-item">
                <span class="label">Default User Active:</span>
                <span class="%s">%s</span>
            </div>
            <div class="config-item">
                <span class="label">Required Attributes:</span>
                <span class="%s">%s</span>
            </div>
        </div>
        
        <div class="section">
            <h3>SAML Endpoints</h3>
            <div class="config-item">
                <span class="label">SSO:</span>
                <span class="value">http://%s/saml/sso</span>
            </div>
            <div class="config-item">
                <span class="label">ACS:</span>
                <span class="value">http://%s/saml/acs</span>
            </div>
            <div class="config-item">
                <span class="label">Metadata:</span>
                <span class="value">http://%s/saml/metadata</span>
            </div>
        </div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #bdc3c7;">
            <a href="/home" style="color: #3498db; text-decoration: none;">Back to Home</a> |
            <a href="/" style="color: #3498db; text-decoration: none;">Test SAML Flow</a>
        </div>
    </div>
</body>
</html>
    `,
		clearedMessage,
		h.config.ServerAddress(),
		h.config.Database.Host, h.config.Database.Port,
		h.config.Database.DBName,
		h.config.Database.User,
		h.config.SAML.EntityID,
		h.config.SAML.ACSURL,
		h.config.SAML.IdPMetadataPath,
		h.config.SAML.CertFile,
		h.config.SAML.KeyFile,
		boolToClass(h.config.JIT.Enabled), boolToString(h.config.JIT.Enabled),
		boolToClass(h.config.JIT.DefaultUserActive), boolToString(h.config.JIT.DefaultUserActive),
		boolToClass(h.config.JIT.RequiredAttributesMode), boolToString(h.config.JIT.RequiredAttributesMode),
		h.config.ServerAddress(),
		h.config.ServerAddress(),
		h.config.ServerAddress(),
	)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// boolToString converts boolean to enabled/disabled string
func boolToString(b bool) string {
	if b {
		return "ENABLED"
	}
	return "DISABLED"
}

// boolToClass converts boolean to CSS class
func boolToClass(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
