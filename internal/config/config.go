package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	SAML     SAMLConfig
	JIT      JITConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// SAMLConfig holds SAML-related configuration
type SAMLConfig struct {
	EntityID        string
	ACSURL          string
	IdPMetadataPath string
	CertFile        string
	KeyFile         string
}

// JITConfig holds Just-In-Time user creation configuration
type JITConfig struct {
	Enabled                bool
	DefaultUserActive      bool
	RequiredAttributesMode bool
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "saml_user"),
			Password: getEnv("DB_PASSWORD", "saml_password"),
			DBName:   getEnv("DB_NAME", "saml_sso"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		SAML: SAMLConfig{
			EntityID:        getEnv("SAML_ENTITY_ID", fmt.Sprintf("http://%s:%s/saml/metadata", getEnv("SERVER_HOST", "localhost"), getEnv("SERVER_PORT", "8080"))),
			ACSURL:          getEnv("SAML_ACS_URL", fmt.Sprintf("http://%s:%s/saml/acs", getEnv("SERVER_HOST", "localhost"), getEnv("SERVER_PORT", "8080"))),
			IdPMetadataPath: getEnv("SAML_IDP_METADATA_PATH", "configs/idp_metadata.xml"),
			CertFile:        getEnv("SAML_CERT_FILE", "sp.crt"),
			KeyFile:         getEnv("SAML_KEY_FILE", "sp.key"),
		},
		JIT: JITConfig{
			Enabled:                getBoolEnv("JIT_ENABLED", true),
			DefaultUserActive:      getBoolEnv("JIT_DEFAULT_USER_ACTIVE", true),
			RequiredAttributesMode: getBoolEnv("JIT_REQUIRED_ATTRIBUTES", true),
		},
	}

	return cfg, nil
}

// DatabaseConnectionString returns the database connection string
func (c *Config) DatabaseConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// ServerAddress returns the server address
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
