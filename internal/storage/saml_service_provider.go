package storage

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	xrv "github.com/mattermost/xml-roundtrip-validator"

	"github.com/crewjam/saml"
)

type ServiceProvider struct {
	ID       string                 `json:"id,omitempty"`
	Metadata *saml.EntityDescriptor `json:"-"`
}

func (sp *ServiceProvider) MarshalJSON() ([]byte, error) {
	a := struct {
		ID          string `json:"id,omitempty"`
		EntityID    string `json:"entityId,omitempty"`
		MetadataURL string `json:"metadataUrl,omitempty"`
	}{
		ID:          sp.ID,
		EntityID:    sp.Metadata.EntityID,
		MetadataURL: fmt.Sprintf("/config/saml_service_providers/%s", sp.ID),
	}
	return json.Marshal(a)
}

func mustMetadata(s *saml.EntityDescriptor, err error) *saml.EntityDescriptor {
	if err != nil {
		panic(err)
	}
	return s
}

func NewMetadata(data []byte) (*saml.EntityDescriptor, error) {
	spMetadata := &saml.EntityDescriptor{}
	if err := xrv.Validate(bytes.NewBuffer(data)); err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, &spMetadata); err != nil {
		if err.Error() == "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
			entities := &saml.EntitiesDescriptor{}
			if err := xml.Unmarshal(data, &entities); err != nil {
				return nil, err
			}

			for _, e := range entities.EntityDescriptors {
				if len(e.SPSSODescriptors) > 0 {
					return &e, nil
				}
			}

			// there were no SPSSODescriptors in the response
			return nil, errors.New("metadata contained no service provider metadata")
		}

		return nil, err
	}

	return spMetadata, nil
}
