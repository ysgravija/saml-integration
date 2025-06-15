# SAML SSO Integration PoC with Database Authentication

A Go-based SAML SSO integration proof-of-concept that authenticates users against a PostgreSQL database. Only users that exist in the database and are active will be allowed to authenticate via SAML SSO.

## Features

- **SAML 2.0 SSO Integration** with mocksaml.com (configurable for other IdPs)
- **PostgreSQL Database Integration** for user validation
- **Docker-based Development Environment**
- **User Authorization Control** - only database users can authenticate
- **Proper SAML Error Handling** for unauthorized users

## Architecture

```
User → SAML IdP → Application → Database Check → Allow/Deny Authentication
```

The application validates each SAML authentication against the PostgreSQL database:
1. User initiates SAML SSO flow
2. IdP authenticates user and sends SAML response
3. Application extracts user email from SAML attributes
4. Application checks if user exists and is active in database
5. If authorized: Complete authentication and show user info
6. If not authorized: Return SAML error response

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- OpenSSL (for certificate generation)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd saml-integration
```

### 2. Start PostgreSQL Database

```bash
# Start PostgreSQL container with sample data
docker-compose up -d

# Verify database is running
docker-compose ps
```

### 3. Generate Certificates (if not already done)

```bash
# Generate self-signed certificate for SAML SP
openssl req -x509 -newkey rsa:2048 -keyout sp.key -out sp.crt -days 365 -nodes \
  -subj "/C=US/ST=CA/L=San Francisco/O=Test/CN=localhost"
```

### 4. Install Dependencies and Run

```bash
# Install Go dependencies
go mod tidy

# Start the application
go run .
```

The application will start on `http://localhost:8080`

## Database Schema

The application uses a simple users table:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Sample Users

The database is pre-populated with test users:

| Email | Name | Status | Can Authenticate |
|-------|------|--------|------------------|
| jackson@example.com | Jackson Smith | Active | ✅ Yes |
| test@example.com | Test User | Active | ✅ Yes |
| admin@example.com | Admin User | Active | ✅ Yes |
| inactive@example.com | Inactive User | Inactive | ❌ No |

## Testing the Integration

### 1. Test with Authorized User

1. Navigate to `http://localhost:8080`
2. You'll be redirected to mocksaml.com
3. Use email: `jackson@example.com` (or any active user from database)
4. Complete SAML authentication
5. You should see a success page with user details

### 2. Test with Unauthorized User

1. Navigate to `http://localhost:8080`
2. You'll be redirected to mocksaml.com  
3. Use email: `unauthorized@example.com` (not in database)
4. Complete SAML authentication
5. You should see an "Access Denied" error

### 3. Test with Inactive User

1. Use email: `inactive@example.com`
2. Should be denied even though user exists in database

## Configuration

### Database Configuration

Edit `database.go` to modify connection settings:

```go
connStr := "host=localhost port=5432 user=saml_user password=saml_password dbname=saml_sso sslmode=disable"
```

### SAML Configuration

Edit `config.go` for SAML settings:

```go
const (
    SPEntityID      = "http://localhost:8080/saml/metadata"
    ACSURL         = "http://localhost:8080/saml/acs"
    IdpMetadataPath = "idp_metadata.xml"
)
```

## Adding New Users

### Via Database

```bash
# Connect to database
docker exec -it saml-postgres psql -U saml_user -d saml_sso

# Add new user
INSERT INTO users (email, first_name, last_name, is_active) 
VALUES ('newuser@example.com', 'New', 'User', true);
```

### Via Application (Future Enhancement)

The application can be extended to include user management endpoints for adding/removing users programmatically.

## Development

### Database Management

```bash
# View all users
docker exec saml-postgres psql -U saml_user -d saml_sso -c "SELECT * FROM users;"

# Add user
docker exec saml-postgres psql -U saml_user -d saml_sso -c "INSERT INTO users (email, first_name, last_name) VALUES ('new@example.com', 'New', 'User');"

# Deactivate user
docker exec saml-postgres psql -U saml_user -d saml_sso -c "UPDATE users SET is_active = false WHERE email = 'user@example.com';"

# Stop database
docker-compose down
```

### Logs and Debugging

The application logs database connections and authentication attempts:

```bash
# View application logs
go run . 2>&1 | grep -E "(database|auth|error)"

# View database logs
docker-compose logs postgres
```

## Security Considerations

1. **Database Security**: Use strong passwords and proper network isolation in production
2. **Certificate Management**: Use proper CA-signed certificates in production
3. **SAML Security**: Validate SAML signatures and assertions properly
4. **User Data**: Consider encrypting sensitive user data in database
5. **Access Logs**: Implement comprehensive audit logging for authentication events

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps
   
   # Check database logs
   docker-compose logs postgres
   ```

2. **SAML Authentication Fails**
   - Verify IdP metadata is correct
   - Check certificate validity
   - Ensure ACS URL matches configuration

3. **User Not Found in Database**
   - Verify user exists: `docker exec saml-postgres psql -U saml_user -d saml_sso -c "SELECT * FROM users WHERE email = 'user@example.com';"`
   - Check if user is active

4. **Port Already in Use**
   ```bash
   # Find and kill process using port 8080
   lsof -ti:8080 | xargs kill
   ```

## Next Steps

- [ ] Add user management API endpoints
- [ ] Implement role-based access control
- [ ] Add audit logging for authentication events
- [ ] Support for multiple SAML IdPs
- [ ] User profile management interface
- [ ] Integration with external user directories (LDAP/AD)

## License

This is a proof-of-concept for educational purposes. 