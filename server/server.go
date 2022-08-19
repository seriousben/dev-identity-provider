package server

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/seriousben/dev-identity-provider/internal/oidc"
	"github.com/seriousben/dev-identity-provider/internal/saml"
	"github.com/seriousben/dev-identity-provider/internal/storage"
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
	storage := storage.NewStorage()
	oidcHandler := oidc.New(fmt.Sprintf("%s/oidc", serverRemoteAddr), storage)
	samlHandler := saml.New(fmt.Sprintf("%s/saml2", serverRemoteAddr), storage)
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.PathPrefix("/oidc").Handler(http.StripPrefix("/oidc", oidcHandler))
	r.PathPrefix("/saml2").Handler(http.StripPrefix("/saml2", samlHandler))
	r.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		users, err := storage.ListUsers()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		sps, err := storage.ListServiceProviders()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		v := map[string]interface{}{
			"users":             users,
			"service_providers": sps,
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		err = enc.Encode(v)
		if err != nil {
			log.Println("error encoding", err)
		}
	})
	r.HandleFunc("/config/saml_service_providers/{serviceID}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sp, err := storage.GetServiceProviderByID(vars["serviceID"])
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		enc := xml.NewEncoder(w)
		enc.Indent("", " ")
		err = enc.Encode(sp.Metadata)
		if err != nil {
			log.Println("error encoding", err)
		}
	})

	return r
}
