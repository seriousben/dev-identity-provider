# dev-identity-provider

dev-identity-provider is a SAML2 and OIDC (OpenID Connect) Provider with SCIM2 support.

## Testing

- https://samltest.id/
- https://openidconnect.net/

## Plan

- [x] OIDC-only server
- [x] SAML2-only server
- [x] OIDC + SAML2 single-server
- [] Unified configuration
- [] Simpified configuration
- [] Dynamic configuration

## Roadmap

- [] OIDC Support
- [] SAML2 Support
- [] Dynamic Config from a given URL (Specific GitHub repo)
- [] Mountable HTTP Handler
- [] Mount HTTP handler in Render
- [] SCIM2

## References

- Read user config from a remote file
- Run server with host based router
- https://github.com/crewjam/saml
- https://github.com/amdonov/lite-idp
- https://github.com/elimity-com/scim
- https://github.com/dexidp/dex
- https://github.com/zitadel/oidc
- https://github.com/zitadel/saml/blob/main/pkg/provider/provider.go
- https://samltest.id/start-idp-test/
- https://developer.okta.com/docs/guides/scim-provisioning-integration-prepare/main/#authentication
- https://github.com/coreos/go-oidc
