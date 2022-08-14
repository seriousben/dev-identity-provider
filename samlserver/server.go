package samlserver

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
	"github.com/seriousben/dev-identity-provider/samlserver/samlidp"
)

//go:embed sp_key.pem
var spKey []byte

//go:embed sp_cert.pem
var spCert []byte

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

type ServerTest struct {
	SPKey         *rsa.PrivateKey
	SPCertificate *x509.Certificate
	SP            saml.ServiceProvider

	Key         crypto.PrivateKey
	Certificate *x509.Certificate
	Server      *samlidp.Server
	Store       samlidp.MemoryStore
}

func New(remoteAddr string) http.Handler {
	test := ServerTest{}
	/*
		saml.TimeNow = func() time.Time {
			rv, _ := time.Parse("Mon Jan 2 15:04:05 MST 2006", "Mon Dec 1 01:57:09 UTC 2015")
			return rv
		}
		jwt.TimeFunc = saml.TimeNow
		saml.RandReader = &testRandomReader{}
	*/

	test.SPKey = mustParsePrivateKey(spKey).(*rsa.PrivateKey)
	test.SPCertificate = mustParseCertificate(spCert)
	test.SP = saml.ServiceProvider{
		Key:         test.SPKey,
		Certificate: test.SPCertificate,
		MetadataURL: mustParseURL("https://sp.example.com/saml2/metadata"),
		AcsURL:      mustParseURL("https://sp.example.com/saml2/acs"),
		IDPMetadata: &saml.EntityDescriptor{},
	}
	test.Key = mustParsePrivateKey(idpKey).(*rsa.PrivateKey)
	test.Certificate = mustParseCertificate(idpCert)

	test.Store = samlidp.MemoryStore{}

	var err error
	test.Server, err = samlidp.New(samlidp.Options{
		Certificate: test.Certificate,
		Key:         test.Key,
		Logger:      logger.DefaultLogger,
		Store:       &test.Store,
		URL:         mustParseURL(remoteAddr),
	})
	if err != nil {
		panic(err)
	}

	test.SP.IDPMetadata = test.Server.IDP.Metadata()
	test.Server.ServiceProviders["https://samltest.id/saml/sp"] = test.SP.Metadata()
	test.Server.InitializeHTTP()

	return test.Server
}
