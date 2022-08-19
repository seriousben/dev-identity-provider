package saml

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"net/http"
	"net/url"

	"github.com/crewjam/saml/logger"
	"github.com/seriousben/dev-identity-provider/internal/saml/samlidp"
	"github.com/seriousben/dev-identity-provider/internal/storage"
)

//go:embed idp_key.pem
var idpKey []byte

//go:embed idp_cert.pem
var idpCert []byte

func mustParseURL(s string) url.URL {
	rv, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return *rv
}

func mustParsePrivateKey(pemStr []byte) crypto.PrivateKey {
	b, _ := pem.Decode(pemStr)
	if b == nil {
		panic("cannot parse PEM")
	}
	k, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		panic(err)
	}
	return k
}

func mustParseCertificate(pemStr []byte) *x509.Certificate {
	b, _ := pem.Decode(pemStr)
	if b == nil {
		panic("cannot parse PEM")
	}
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}

type Storage interface {
	ListUsers() ([]*storage.User, error)
	GetUserByID(string) (*storage.User, error)
	DeleteUser(string) error
	PutUser(string, *storage.User) error

	ListServiceProviders() ([]*storage.ServiceProvider, error)
	GetServiceProviderByID(string) (*storage.ServiceProvider, error)
	GetServiceProviderByEntityID(string) (*storage.ServiceProvider, error)
	DeleteServiceProvider(string) error
	PutServiceProvider(string, *storage.ServiceProvider) error
}

func New(remoteAddr string, stor Storage) http.Handler {
	key := mustParsePrivateKey(idpKey).(*rsa.PrivateKey)
	certificate := mustParseCertificate(idpCert)

	store := MemoryStore{
		storage: stor,
	}

	server, err := samlidp.New(samlidp.Options{
		Certificate: certificate,
		Key:         key,
		Logger:      logger.DefaultLogger,
		Store:       &store,
		URL:         mustParseURL(remoteAddr),
	})
	if err != nil {
		panic(err)
	}

	server.InitializeHTTP()

	return server
}
