package samlidp

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/seriousben/dev-identity-provider/internal/storage"
	"github.com/zenazn/goji/web"

	"github.com/crewjam/saml"
)

// GetServiceProvider returns the Service Provider metadata for the
// service provider ID, which is typically the service provider's
// metadata URL. If an appropriate service provider cannot be found then
// the returned error must be os.ErrNotExist.
func (s *Server) GetServiceProvider(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	service := storage.ServiceProvider{}
	err := s.Store.Get(fmt.Sprintf("/services-by-entity-id/%s", serviceProviderID), &service)
	if err != nil {
		return nil, err
	}
	return service.Metadata, nil
}

// HandleListServices handles the `GET /services/` request and responds with a JSON formatted list
// of service names.
func (s *Server) HandleListServices(c web.C, w http.ResponseWriter, r *http.Request) {
	services, err := s.Store.List("/services/")
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(struct {
		Services []string `json:"services"`
	}{Services: services})
}

// HandleGetService handles the `GET /services/:id` request and responds with the service
// metadata in XML format.
func (s *Server) HandleGetService(c web.C, w http.ResponseWriter, r *http.Request) {
	service := storage.ServiceProvider{}
	err := s.Store.Get(fmt.Sprintf("/services/%s", c.URLParams["id"]), &service)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	xml.NewEncoder(w).Encode(service.Metadata)
}

// HandlePutService handles the `PUT /shortcuts/:id` request. It accepts the XML-formatted
// service metadata in the request body and stores it.
func (s *Server) HandlePutService(c web.C, w http.ResponseWriter, r *http.Request) {
	service := storage.ServiceProvider{}

	metadata, err := GetSPMetadata(r.Body)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	service.Metadata = metadata

	err = s.Store.Put(fmt.Sprintf("/services/%s", c.URLParams["id"]), &service)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteService handles the `DELETE /services/:id` request.
func (s *Server) HandleDeleteService(c web.C, w http.ResponseWriter, r *http.Request) {
	service := storage.ServiceProvider{}
	err := s.Store.Get(fmt.Sprintf("/services/%s", c.URLParams["id"]), &service)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := s.Store.Delete(fmt.Sprintf("/services/%s", c.URLParams["id"])); err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
