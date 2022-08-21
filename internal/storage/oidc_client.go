package storage

import (
	"time"

	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/oidc/pkg/op"
)

var (
	//we use the default login UI and pass the (auth request) id
	defaultLoginURL = func(id string) string {
		return "/oidc/login/username?authRequestID=" + id
	}
)

//Client represents the internal model of an OAuth/OIDC client
//this could also be your database model
type Client struct {
	ID                             string             `json:"clientId,omitempty"`
	Secret                         string             `json:"clientSecret,omitempty"`
	ClientRedirectURIs             []string           `json:"redirectURIs,omitempty"`
	ClientApplicationType          op.ApplicationType `json:"applicationType,omitempty"`
	ClientAuthMethod               oidc.AuthMethod    `json:"authMethod,omitempty"`
	loginURL                       func(string) string
	ClientResponseTypes            []oidc.ResponseType `json:"responseTypes,omitempty"`
	ClientGrantTypes               []oidc.GrantType    `json:"grantTypes,omitempty"`
	ClientAccessTokenType          op.AccessTokenType  `json:"accessTokenType,omitempty"`
	devMode                        bool
	idTokenUserinfoClaimsAssertion bool
	clockSkew                      time.Duration
}

//GetID must return the client_id
func (c *Client) GetID() string {
	return c.ID
}

//RedirectURIs must return the registered redirect_uris for Code and Implicit Flow
func (c *Client) RedirectURIs() []string {
	return c.ClientRedirectURIs
}

//PostLogoutRedirectURIs must return the registered post_logout_redirect_uris for sign-outs
func (c *Client) PostLogoutRedirectURIs() []string {
	return []string{}
}

//ApplicationType must return the type of the client (app, native, user agent)
func (c *Client) ApplicationType() op.ApplicationType {
	return c.ClientApplicationType
}

//AuthMethod must return the authentication method (client_secret_basic, client_secret_post, none, private_key_jwt)
func (c *Client) AuthMethod() oidc.AuthMethod {
	return c.ClientAuthMethod
}

//ResponseTypes must return all allowed response types (code, id_token token, id_token)
//these must match with the allowed grant types
func (c *Client) ResponseTypes() []oidc.ResponseType {
	return c.ClientResponseTypes
}

//GrantTypes must return all allowed grant types (authorization_code, refresh_token, urn:ietf:params:oauth:grant-type:jwt-bearer)
func (c *Client) GrantTypes() []oidc.GrantType {
	return c.ClientGrantTypes
}

//LoginURL will be called to redirect the user (agent) to the login UI
//you could implement some logic here to redirect the users to different login UIs depending on the client
func (c *Client) LoginURL(id string) string {
	return c.loginURL(id)
}

//AccessTokenType must return the type of access token the client uses (Bearer (opaque) or JWT)
func (c *Client) AccessTokenType() op.AccessTokenType {
	return c.ClientAccessTokenType
}

//IDTokenLifetime must return the lifetime of the client's id_tokens
func (c *Client) IDTokenLifetime() time.Duration {
	return 1 * time.Hour
}

//DevMode enables the use of non-compliant configs such as redirect_uris (e.g. http schema for user agent client)
func (c *Client) DevMode() bool {
	return c.devMode
}

//RestrictAdditionalIdTokenScopes allows specifying which custom scopes shall be asserted into the id_token
func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

//RestrictAdditionalAccessTokenScopes allows specifying which custom scopes shall be asserted into the JWT access_token
func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

//IsScopeAllowed enables Client specific custom scopes validation
//in this example we allow the CustomScope for all clients
func (c *Client) IsScopeAllowed(scope string) bool {
	return scope == CustomScope
}

//IDTokenUserinfoClaimsAssertion allows specifying if claims of scope profile, email, phone and address are asserted into the id_token
//even if an access token if issued which violates the OIDC Core spec
//(5.4. Requesting Claims using Scope Values: https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims)
//some clients though require that e.g. email is always in the id_token when requested even if an access_token is issued
func (c *Client) IDTokenUserinfoClaimsAssertion() bool {
	return c.idTokenUserinfoClaimsAssertion
}

//ClockSkew enables clients to instruct the OP to apply a clock skew on the various times and expirations
//(subtract from issued_at, add to expiration, ...)
func (c *Client) ClockSkew() time.Duration {
	return c.clockSkew
}

//NativeClient will create a client of type native, which will always use PKCE and allow the use of refresh tokens
//user-defined redirectURIs may include:
// - http://localhost without port specification (e.g. http://localhost/auth/callback)
// - custom protocol (e.g. custom://auth/callback)
//(the examples will be used as default, if none is provided)
func NativeClient(id string, redirectURIs ...string) *Client {
	if len(redirectURIs) == 0 {
		redirectURIs = []string{
			"http://localhost/auth/callback",
			"custom://auth/callback",
		}
	}
	return &Client{
		ID:                             id,
		Secret:                         "", //no secret needed (due to PKCE)
		ClientRedirectURIs:             redirectURIs,
		ClientApplicationType:          op.ApplicationTypeNative,
		ClientAuthMethod:               oidc.AuthMethodNone,
		loginURL:                       defaultLoginURL,
		ClientResponseTypes:            []oidc.ResponseType{oidc.ResponseTypeCode},
		ClientGrantTypes:               []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		ClientAccessTokenType:          op.AccessTokenTypeBearer,
		devMode:                        false,
		idTokenUserinfoClaimsAssertion: false,
		clockSkew:                      0,
	}
}

//WebClient will create a client of type web, which will always use Basic Auth and allow the use of refresh tokens
//user-defined redirectURIs may include:
// - http://localhost with port specification (e.g. http://localhost:9999/auth/callback)
//(the example will be used as default, if none is provided)
func WebClient(id, secret string, redirectURIs ...string) *Client {
	if len(redirectURIs) == 0 {
		redirectURIs = []string{
			"http://localhost:9999/auth/callback",
		}
	}
	return &Client{
		ID:                             id,
		Secret:                         secret,
		ClientRedirectURIs:             redirectURIs,
		ClientApplicationType:          op.ApplicationTypeWeb,
		ClientAuthMethod:               oidc.AuthMethodBasic,
		loginURL:                       defaultLoginURL,
		ClientResponseTypes:            []oidc.ResponseType{oidc.ResponseTypeCode},
		ClientGrantTypes:               []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		ClientAccessTokenType:          op.AccessTokenTypeBearer,
		devMode:                        false,
		idTokenUserinfoClaimsAssertion: false,
		clockSkew:                      0,
	}
}
