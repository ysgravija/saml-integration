# Go SAML SSO Integration PoC with JumpCloud

This project is a proof-of-concept (PoC) for integrating SAML-based Single Sign-On (SSO) with JumpCloud using Go.

## Features
- SAML authentication with JumpCloud as the Identity Provider (IdP)
- Simple web server with SSO-protected endpoint
- Example configuration and metadata exchange

## Prerequisites
- Go 1.18+
- JumpCloud account with SAML application configured

## Setup
1. Clone this repository.
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Configure your JumpCloud SAML application:
   - Set the ACS (Assertion Consumer Service) URL to `http://localhost:8080/saml/acs`
   - Set the Entity ID to `http://localhost:8080/saml/metadata`
   - Download the IdP metadata XML from JumpCloud
4. Place the IdP metadata XML as `idp_metadata.xml` in the project root.

## Running the Project
```sh
go run main.go
```

- Visit `http://localhost:8080/` to start the SSO flow.

## Configuration
- Edit `config.go` to set your SAML settings (entity ID, ACS URL, etc.)
- Place your IdP metadata in `idp_metadata.xml`

## References
- [JumpCloud SAML Docs](https://jumpcloud.com/support/saml)
- [github.com/crewjam/saml](https://github.com/crewjam/saml) 