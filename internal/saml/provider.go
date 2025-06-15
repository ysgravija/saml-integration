package saml

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"

	"saml-poc/internal/config"
)

// Provider wraps SAML service provider functionality
type Provider struct {
	SP     *samlsp.Middleware
	config *config.Config
}

// NewProvider creates a new SAML provider
func NewProvider(cfg *config.Config) (*Provider, error) {
	// Load IdP metadata
	idpMetadata, err := loadIdpMetadata(cfg.SAML.IdPMetadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load IdP metadata: %w", err)
	}

	// Load SP key pair
	keyPair, err := tls.LoadX509KeyPair(cfg.SAML.CertFile, cfg.SAML.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load SP key pair: %w", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse SP certificate: %w", err)
	}

	// Cast private key to *rsa.PrivateKey for crypto.Signer interface
	rsaPrivateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	rootURL, err := url.Parse(fmt.Sprintf("http://%s", cfg.ServerAddress()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse root URL: %w", err)
	}

	// Configure SAML middleware
	samlSP, err := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         rsaPrivateKey,
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
		EntityID:    cfg.SAML.EntityID,
		SignRequest: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SAML SP: %w", err)
	}

	return &Provider{
		SP:     samlSP,
		config: cfg,
	}, nil
}

// GetMiddleware returns the SAML middleware
func (p *Provider) GetMiddleware() *samlsp.Middleware {
	return p.SP
}

// loadIdpMetadata loads IdP metadata from file
func loadIdpMetadata(path string) (*saml.EntityDescriptor, error) {
	metadataXML, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read IdP metadata file: %w", err)
	}

	var metadata saml.EntityDescriptor
	if err := xml.Unmarshal(metadataXML, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IdP metadata: %w", err)
	}

	return &metadata, nil
}
