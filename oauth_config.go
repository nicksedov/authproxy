package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
)

type OAuthConfig struct {
	Web struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
	} `json:"web"`
}

func loadOAuthConfig(filename string) (*oauth2.Config, error) {
	log.Printf("Loading OAuth config: %s", filename)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config OAuthConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if config.Web.ClientID == "" {
		return nil, errors.New("missing client_id in config")
	}
	if config.Web.ClientSecret == "" {
		return nil, errors.New("missing client_secret in config")
	}
	if len(config.Web.RedirectURIs) == 0 {
		return nil, errors.New("missing redirect_uris in config")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.Web.ClientID,
		ClientSecret: config.Web.ClientSecret,
		RedirectURL:  config.Web.RedirectURIs[0],
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Web.AuthURI,
			TokenURL: config.Web.TokenURI,
		},
	}

	log.Printf("OAuth config loaded: ClientID=%s", config.Web.ClientID)
	return oauthConfig, nil
}