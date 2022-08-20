package server

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/seriousben/dev-identity-provider/internal/oidc"
	"github.com/seriousben/dev-identity-provider/internal/saml"
	"github.com/seriousben/dev-identity-provider/internal/storage"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println("Request", r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func syncStorage(basePath string, s *storage.Storage) error {
	fmt.Println("Syncing storage")

	var config struct {
		ServiceProviders []struct {
			ID          string `json:"id"`
			MetadataURL string `json:"metadataUrl"`
		} `json:"service_providers"`
		Users []*storage.User `json:"users"`
	}

	resp, err := http.Get(fmt.Sprintf("%s/config.json", basePath))
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return err
	}

	for _, sp := range config.ServiceProviders {
		spResp, err := http.Get(fmt.Sprintf("%s/%s", basePath, strings.TrimPrefix(sp.MetadataURL, "/")))
		if err != nil {
			return err
		}
		b, err = io.ReadAll(spResp.Body)
		if err != nil {
			return err
		}

		meta, err := storage.NewMetadata(b)
		if err != nil {
			return err
		}

		if err := s.PutServiceProvider(sp.ID, &storage.ServiceProvider{
			ID:       sp.ID,
			Metadata: meta,
		}); err != nil {
			return err
		}
	}

	for i, u := range config.Users {
		if err := s.PutUser(u.ID, config.Users[i]); err != nil {
			return err
		}
	}

	return nil
}

func New(serverRemoteAddr string) http.Handler {
	storage := storage.NewStorage()

	if err := syncStorage("https://raw.githubusercontent.com/seriousben/dev-identity-provider-config/main/", storage); err != nil {
		panic(err)
	}
	go func() {
		for {
			// Reset/Sync config daily
			time.Sleep(24 * time.Hour)
			if err := syncStorage("https://raw.githubusercontent.com/seriousben/dev-identity-provider-config/main/", storage); err != nil {
				fmt.Println("error syncing", err)
			}
		}
	}()

	oidcHandler := oidc.New(fmt.Sprintf("%s/oidc", serverRemoteAddr), storage)
	samlHandler := saml.New(fmt.Sprintf("%s/saml2", serverRemoteAddr), storage)
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.PathPrefix("/oidc").Handler(http.StripPrefix("/oidc", oidcHandler))
	r.PathPrefix("/saml2").Handler(http.StripPrefix("/saml2", samlHandler))
	r.Methods("sdf").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	r.Path("/refresh-config").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if err := syncStorage("https://raw.githubusercontent.com/seriousben/dev-identity-provider-config/main/", storage); err != nil {
			log.Println("error", err)
			fmt.Fprintf(w, `{"error": %q}`, err.Error())
			return
		}
		w.Write([]byte(`{"success": "true"}`))
	})
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
