package storage

import (
	"bytes"
	"encoding/xml"
	"errors"

	xrv "github.com/mattermost/xml-roundtrip-validator"

	"github.com/crewjam/saml"
)

type ServiceProvider struct {
	ID       string                 `json:"id,omitempty"`
	Metadata *saml.EntityDescriptor `json:"-"`
}

type ServiceProviderDetailed struct {
	ID          string `json:"id,omitempty"`
	EntityID    string `json:"entityId,omitempty"`
	MetadataURL string `json:"metadataUrl,omitempty"`
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
