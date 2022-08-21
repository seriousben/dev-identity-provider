package server

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/seriousben/dev-identity-provider/internal/oidc"
	"github.com/seriousben/dev-identity-provider/internal/saml"
	"github.com/seriousben/dev-identity-provider/internal/storage"
)

var (
	//go:embed index.html
	indexHTML string
	indextmpl = template.Must(template.New("index").Funcs(template.FuncMap{
		"ToJSON": func(v interface{}) string {
			build := &strings.Builder{}
			enc := json.NewEncoder(build)
			enc.SetIndent("", " ")
			if err := enc.Encode(v); err != nil {
				return err.Error()
			}
			return build.String()
		},
	}).Parse(indexHTML))
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
			ID          string `json:"id,omitempty"`
			MetadataURL string `json:"metadataUrl,omitempty"`
		} `json:"service_providers"`
		Users   []*storage.User `json:"users"`
		Clients []struct {
			ClientID     string   `json:"clientId,omitempty"`
			ClientSecret string   `json:"clientSecret,omitempty"`
			RedirectURIs []string `json:"redirectUris,omitempty"`
		} `json:"clients"`
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

	for _, u := range config.Clients {
		cl := storage.WebClient(u.ClientID, u.ClientSecret, u.RedirectURIs...)
		if err := s.RegisterClient(cl.ID, cl); err != nil {
			return err
		}
	}

	return nil
}

func New(serverRemoteAddr string) http.Handler {
	stor := storage.NewStorage()

	if err := syncStorage("https://raw.githubusercontent.com/seriousben/dev-identity-provider-config/main/", stor); err != nil {
		panic(err)
	}
	go func() {
		for {
			// Reset/Sync config daily
			time.Sleep(24 * time.Hour)
			if err := syncStorage("https://raw.githubusercontent.com/seriousben/dev-identity-provider-config/main/", stor); err != nil {
				fmt.Println("error syncing", err)
			}
		}
	}()

	oidcHandler := oidc.New(fmt.Sprintf("%s/oidc", serverRemoteAddr), stor)
	samlHandler := saml.New(fmt.Sprintf("%s/saml2", serverRemoteAddr), stor)
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	r.PathPrefix("/oidc").Handler(http.StripPrefix("/oidc", oidcHandler))
	r.PathPrefix("/saml2").Handler(http.StripPrefix("/saml2", samlHandler))
	r.Methods("sdf").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	r.Path("/refresh-config").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if err := syncStorage("https://raw.githubusercontent.com/seriousben/dev-identity-provider-config/main/", stor); err != nil {
			log.Println("error", err)
			fmt.Fprintf(w, `{"error": %q}`, err.Error())
			return
		}
		w.Write([]byte(`{"success": "true"}`))
	})
	r.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		users, err := stor.ListUsers()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		sps, err := stor.ListServiceProviders()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		spds := make([]storage.ServiceProviderDetailed, len(sps))
		for i, sp := range sps {
			spds[i] = storage.ServiceProviderDetailed{
				ID:          sp.ID,
				EntityID:    sp.Metadata.EntityID,
				MetadataURL: fmt.Sprintf("%s/config/saml_service_providers/%s", serverRemoteAddr, sp.ID),
			}
		}
		v := map[string]interface{}{
			"users":             users,
			"service_providers": spds,
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
		sp, err := stor.GetServiceProviderByID(vars["serviceID"])
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
	r.Path("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clients, err := stor.ListClients()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		users, err := stor.ListUsers()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		sps, err := stor.ListServiceProviders()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		spds := make([]storage.ServiceProviderDetailed, len(sps))
		for i, sp := range sps {
			spds[i] = storage.ServiceProviderDetailed{
				ID:          sp.ID,
				EntityID:    sp.Metadata.EntityID,
				MetadataURL: fmt.Sprintf("%s/config/saml_service_providers/%s", serverRemoteAddr, sp.ID),
			}
		}

		version := "dev"
		if info, ok := debug.ReadBuildInfo(); ok {
			var vcsTime, vcsRev string
			for _, s := range info.Settings {
				if s.Key == "vcs.time" {
					vcsTime = s.Value
				}
				if s.Key == "vcs.revision" {
					vcsRev = s.Value
				}
				if vcsRev != "" && vcsTime != "" {
					version = fmt.Sprintf("Time=%s / Revision=%s", vcsTime, vcsRev)
				}
			}
		}

		v := map[string]interface{}{
			"Clients":          clients,
			"Users":            users,
			"ServiceProviders": spds,
			"Version":          version,
		}
		indextmpl.Execute(w, v)
	}))

	return r
}
