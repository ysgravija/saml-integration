package main

const (
	SPEntityID      = "http://localhost:8080/saml/metadata"
	ACSURL          = "http://localhost:8080/saml/acs"
	IdpMetadataPath = "idp_metadata.xml"
)

// JIT (Just-In-Time) User Creation Configuration
const (
	// EnableJIT controls whether to automatically create users on first SAML login
	EnableJIT = true

	// DefaultUserActive sets whether JIT-created users are active by default
	DefaultUserActive = true

	// RequiredAttributes defines which SAML attributes are required for JIT user creation
	// If any of these are missing, JIT creation will fail
	RequiredAttributesForJIT = true // Set to false to allow JIT creation with minimal attributes
)
