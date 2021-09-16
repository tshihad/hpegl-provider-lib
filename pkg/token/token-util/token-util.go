//(C) Copyright 2019 Hewlett Packard Enterprise Development LP

package tokenutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2"
)

// Token a jwt token format
type Token struct {
	Issuer           string `json:"iss"`
	Subject          string `json:"sub"`
	Expiry           int64  `json:"exp"`
	IssuedAt         int64  `json:"iat"`
	Type             string `json:"typ"`
	Nonce            string `json:"nonce"`
	AtHash           string `json:"at_hash"`
	ClientID         string `json:"cid,omitempty"`
	UserID           string `json:"uid,omitempty"`
	TenantID         string `json:"tenantId"`
	AuthorizedParty  string `json:"azp"`
	KeycloakClientID string `json:"clientId"`
	IsHPE            bool   `json:"isHPE"`
}

// DecodeAccessToken decodes the accessToken offline
func DecodeAccessToken(rawToken string) (Token, error) {
	_, err := jose.ParseSigned(rawToken)
	if err != nil {
		return Token{}, fmt.Errorf("oidc: malformed jwt: %v", err)
	}

	// Throw out tokens with invalid claims before trying to verify the token. This lets
	// us do cheap checks before possibly re-syncing keys.
	payload, err := parseJWT(rawToken)
	if err != nil {
		log.Errorf("oidc: malformed jwt: %v", err)
		return Token{}, fmt.Errorf("oidc: malformed jwt: %v", err)
	}
	var token Token
	if err := json.Unmarshal(payload, &token); err != nil {
		log.Errorf("oidc: failed to unmarshal claims: %v", err)
		return Token{}, fmt.Errorf("oidc: failed to unmarshal claims: %v", err)
	}

	if token.UserID != "" {
		// Okta User token
		token.Subject = "users/" + token.UserID
	} else if token.ClientID != "" || token.KeycloakClientID != "" {
		token.Subject = "clients/" + token.Subject
	} else {
		// TODO This is just so that Keycloak tokens continue to work. Remove after keycloak is gone
		token.Subject = "users/" + token.Subject
	}

	return token, nil
}

func parseJWT(p string) ([]byte, error) {
	parts := strings.Split(p, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("oidc: malformed jwt, expected 3 parts got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt payload: %v", err)
	}
	return payload, nil
}
