package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/seriousben/dev-identity-provider/oidcserver"
	"github.com/seriousben/dev-identity-provider/samlserver"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func New(serverRemoteAddr string) http.Handler {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.PathPrefix("/oidc").Handler(http.StripPrefix("/oidc", oidcserver.New(fmt.Sprintf("%s/oidc", serverRemoteAddr))))
	r.PathPrefix("/saml2").Handler(http.StripPrefix("/saml2", samlserver.New(fmt.Sprintf("%s/saml2", serverRemoteAddr))))

	return r
}
