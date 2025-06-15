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

func main() {
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

	// SAML endpoints
	http.Handle("/saml/", samlSP)

	// Protected endpoint
	http.Handle("/", samlSP.RequireAccount(http.HandlerFunc(homeHandler)))

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get SAML attributes from context
	session := samlsp.SessionFromContext(r.Context())
	if session != nil {
		fmt.Fprintf(w, "<h1>Welcome!</h1><p>SAML session: %+v</p>", session)
	} else {
		fmt.Fprintln(w, "<h1>Not authenticated</h1>")
	}
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
