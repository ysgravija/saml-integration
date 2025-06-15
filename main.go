package main

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

var database *Database

func main() {
	// Initialize database connection
	var err error
	database, err = NewDatabase()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	// Load IdP metadata
	idpMetadata, err := loadIdpMetadata(IdpMetadataPath)
	if err != nil {
		log.Fatalf("failed to load IdP metadata: %v", err)
	}

	// Load SP key pair
	keyPair, err := tls.LoadX509KeyPair("sp.crt", "sp.key")
	if err != nil {
		log.Fatalf("failed to load SP key pair: %v", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		log.Fatalf("failed to parse SP certificate: %v", err)
	}

	// Cast private key to *rsa.PrivateKey for crypto.Signer interface
	rsaPrivateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("private key is not RSA")
	}

	rootURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatalf("failed to parse root URL: %v", err)
	}

	// Configure SAML middleware
	samlSP, err := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         rsaPrivateKey,
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
		EntityID:    SPEntityID,
		SignRequest: true,
	})
	if err != nil {
		log.Fatalf("failed to create SAML SP: %v", err)
	}

	// SAML endpoints - register with prefix pattern
	http.Handle("/saml/", samlSP)

	// Debug endpoint (unprotected)
	http.HandleFunc("/debug", debugHandler)

	// Protected home endpoint with database validation middleware
	http.Handle("/home", samlSP.RequireAccount(databaseValidationMiddleware(http.HandlerFunc(homeHandler))))

	// Root redirect to protected home - this will trigger SAML auth if not authenticated
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/home", http.StatusFound)
	})

	fmt.Println("Server started at http://localhost:8080")
	fmt.Println("Database connection established")
	fmt.Println("SAML endpoints:")
	fmt.Println("  - SSO: http://localhost:8080/saml/sso")
	fmt.Println("  - ACS: http://localhost:8080/saml/acs")
	fmt.Println("  - Metadata: http://localhost:8080/saml/metadata")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// databaseValidationMiddleware validates users against the database after SAML authentication
func databaseValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get SAML session from context (already authenticated by SAML)
		session := samlsp.SessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "No SAML session found", http.StatusUnauthorized)
			return
		}

		// Extract user email from SAML session
		userEmail := extractEmailFromSession(session, r)
		if userEmail == "" {
			log.Println("No email found in SAML session")
			http.Error(w, "No email found in SAML session", http.StatusBadRequest)
			return
		}

		log.Printf("Validating user from SAML session: %s", userEmail)

		// Validate user against database
		authorized, user, err := database.IsUserAuthorized(userEmail)
		if err != nil {
			log.Printf("Database error during user validation: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !authorized {
			log.Printf("User not authorized: %s", userEmail)
			http.Error(w, "Access denied: User not authorized for this application", http.StatusForbidden)
			return
		}

		log.Printf("User successfully validated: %s (%s %s)", user.Email, user.FirstName, user.LastName)

		// User is authorized, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// extractEmailFromSession extracts email from various possible SAML attributes
func extractEmailFromSession(session samlsp.Session, r *http.Request) string {
	var userEmail string

	// Try to get email from session attributes
	if sessionWithAttrs, ok := session.(samlsp.SessionWithAttributes); ok {
		attrs := sessionWithAttrs.GetAttributes()
		// Try different common attribute names for email
		if email := attrs.Get("email"); email != "" {
			userEmail = email
		} else if email := attrs.Get("emailAddress"); email != "" {
			userEmail = email
		} else if email := attrs.Get("mail"); email != "" {
			userEmail = email
		} else if email := attrs.Get("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"); email != "" {
			userEmail = email
		}
	}

	// If we couldn't get email from attributes, try using AttributeFromContext helper
	if userEmail == "" {
		userEmail = samlsp.AttributeFromContext(r.Context(), "email")
		if userEmail == "" {
			userEmail = samlsp.AttributeFromContext(r.Context(), "emailAddress")
		}
		if userEmail == "" {
			userEmail = samlsp.AttributeFromContext(r.Context(), "mail")
		}
		if userEmail == "" {
			userEmail = samlsp.AttributeFromContext(r.Context(), "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress")
		}
	}

	// If still no email, try to get it from the NameID (common in test environments)
	if userEmail == "" {
		if jwtSession, ok := session.(*samlsp.JWTSessionClaims); ok {
			userEmail = jwtSession.Subject
		}
	}

	return userEmail
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get SAML session from context
	session := samlsp.SessionFromContext(r.Context())
	if session == nil {
		fmt.Fprintln(w, "<h1>Not authenticated</h1>")
		return
	}

	// Extract user email from session
	userEmail := extractEmailFromSession(session, r)
	if userEmail == "" {
		log.Println("No email found in SAML session")
		http.Error(w, "No email found in SAML session", http.StatusBadRequest)
		return
	}

	// Get user details from database
	user, err := database.GetUserByEmail(userEmail)
	if err != nil {
		log.Printf("Error fetching user from database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Display welcome page with user information
	fmt.Fprintf(w, `
		<html>
		<head>
			<title>SAML SSO - Welcome</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; }
				.header { background: #4CAF50; color: white; padding: 20px; border-radius: 5px; }
				.content { margin: 20px 0; }
				.user-info { background: #f5f5f5; padding: 15px; border-radius: 5px; }
				.debug { background: #e8e8e8; padding: 10px; border-radius: 5px; font-family: monospace; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="header">
				<h1>Welcome, %s %s!</h1>
				<p>SAML Authentication Successful ✅</p>
			</div>
			<div class="content">
				<div class="user-info">
					<h3>User Information:</h3>
					<p><strong>Email:</strong> %s</p>
					<p><strong>User ID:</strong> %d</p>
					<p><strong>Status:</strong> %s</p>
					<p><strong>Account Created:</strong> Database user</p>
				</div>
				<div class="debug">
					<h4>Debug Information:</h4>
					<p><strong>SAML Session Type:</strong> %T</p>
					<p><strong>Authentication Method:</strong> SAML SSO with Database Validation</p>
				</div>
			</div>
		</body>
		</html>
	`, user.FirstName, user.LastName, user.Email, user.ID,
		map[bool]string{true: "Active", false: "Inactive"}[user.IsActive], session)
}

func loadIdpMetadata(path string) (*saml.EntityDescriptor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var entity saml.EntityDescriptor
	if err := xml.Unmarshal(data, &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get SAML session from context
	session := samlsp.SessionFromContext(r.Context())

	fmt.Fprintf(w, `
		<html>
		<head>
			<title>SAML Debug Information</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; }
				.section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
				.error { background: #ffebee; border-color: #f44336; }
				.success { background: #e8f5e8; border-color: #4caf50; }
				.info { background: #e3f2fd; border-color: #2196f3; }
				pre { background: #f5f5f5; padding: 10px; border-radius: 3px; overflow-x: auto; }
			</style>
		</head>
		<body>
			<h1>🔍 SAML Debug Information</h1>
	`)

	if session == nil {
		fmt.Fprintf(w, `
			<div class="section error">
				<h3>❌ No SAML Session</h3>
				<p>No SAML session found. User is not authenticated.</p>
				<p><a href="/">Click here to start SAML authentication</a></p>
			</div>
		`)
	} else {
		fmt.Fprintf(w, `
			<div class="section success">
				<h3>✅ SAML Session Found</h3>
				<p><strong>Session Type:</strong> %T</p>
			</div>
		`, session)

		// Try to extract email
		userEmail := extractEmailFromSession(session, r)
		if userEmail != "" {
			fmt.Fprintf(w, `
				<div class="section success">
					<h3>✅ Email Extracted</h3>
					<p><strong>Email:</strong> %s</p>
				</div>
			`, userEmail)

			// Check database authorization
			authorized, user, err := database.IsUserAuthorized(userEmail)
			if err != nil {
				fmt.Fprintf(w, `
					<div class="section error">
						<h3>❌ Database Error</h3>
						<p><strong>Error:</strong> %v</p>
					</div>
				`, err)
			} else if !authorized {
				fmt.Fprintf(w, `
					<div class="section error">
						<h3>❌ User Not Authorized</h3>
						<p>User exists in database: %t</p>
						<p>User is active: %t</p>
					</div>
				`, user != nil, user != nil && user.IsActive)
			} else {
				fmt.Fprintf(w, `
					<div class="section success">
						<h3>✅ User Authorized</h3>
						<p><strong>Name:</strong> %s %s</p>
						<p><strong>Email:</strong> %s</p>
						<p><strong>Status:</strong> %s</p>
					</div>
				`, user.FirstName, user.LastName, user.Email,
					map[bool]string{true: "Active", false: "Inactive"}[user.IsActive])
			}
		} else {
			fmt.Fprintf(w, `
				<div class="section error">
					<h3>❌ No Email Found</h3>
					<p>Could not extract email from SAML session.</p>
				</div>
			`)
		}

		// Show session attributes if available
		if sessionWithAttrs, ok := session.(samlsp.SessionWithAttributes); ok {
			attrs := sessionWithAttrs.GetAttributes()
			fmt.Fprintf(w, `
				<div class="section info">
					<h3>📋 SAML Attributes</h3>
					<pre>%+v</pre>
				</div>
			`, attrs)
		}

		// Show raw session data
		fmt.Fprintf(w, `
			<div class="section info">
				<h3>🔧 Raw Session Data</h3>
				<pre>%+v</pre>
			</div>
		`, session)
	}

	// Show database users
	fmt.Fprintf(w, `
		<div class="section info">
			<h3>👥 Database Users</h3>
			<p>Users authorized for this application:</p>
	`)

	// Query database for all users
	rows, err := database.db.Query("SELECT email, first_name, last_name, is_active FROM users ORDER BY email")
	if err != nil {
		fmt.Fprintf(w, `<p style="color: red;">Error querying database: %v</p>`, err)
	} else {
		defer rows.Close()
		fmt.Fprintf(w, `<table border="1" style="border-collapse: collapse; width: 100%%;">
			<tr><th>Email</th><th>Name</th><th>Status</th></tr>`)

		for rows.Next() {
			var email, firstName, lastName string
			var isActive bool
			if err := rows.Scan(&email, &firstName, &lastName, &isActive); err == nil {
				status := "Active"
				statusColor := "green"
				if !isActive {
					status = "Inactive"
					statusColor = "red"
				}
				fmt.Fprintf(w, `<tr><td>%s</td><td>%s %s</td><td style="color: %s;">%s</td></tr>`,
					email, firstName, lastName, statusColor, status)
			}
		}
		fmt.Fprintf(w, `</table>`)
	}

	fmt.Fprintf(w, `
		</div>
		<div class="section info">
			<h3>🔗 Useful Links</h3>
			<ul>
				<li><a href="/">Home (Start SAML Auth)</a></li>
				<li><a href="/saml/metadata">SAML Metadata</a></li>
				<li><a href="/debug">Refresh Debug Info</a></li>
			</ul>
		</div>
		</body>
		</html>
	`)
}
