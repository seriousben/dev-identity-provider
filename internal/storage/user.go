package storage

import (
	"crypto/rsa"
)

type User struct {
	ID            string   `json:"id,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
	Firstname     string   `json:"firstname,omitempty"`
	Lastname      string   `json:"lastname,omitempty"`
	Groups        []string `json:"groups,omitempty"`
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"emailVerified,omitempty"`
	/*
		PreferredLanguage language.Tag
		CommonName        string   `json:"common_name,omitempty"`
		Surname           string   `json:"surname,omitempty"`
		GivenName         string   `json:"given_name,omitempty"`
		ScopedAffiliation string   `json:"scoped_affiliation,omitempty"`
	*/
}

type Service struct {
	keys map[string]*rsa.PublicKey
}
