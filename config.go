package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
)

type Config struct {
	Web struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		
	} `json:"web"`
}

func loadConfig(filename string) error {
	log.Printf("Loading configuration: %s", filename)

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&appConfig); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	if appConfig.Web.ClientID == "" {
		return errors.New("missing client_id in config")
	}
	if appConfig.Web.ClientSecret == "" {
		return errors.New("missing client_secret in config")
	}
	if len(appConfig.Web.RedirectURIs) == 0 {
		return errors.New("missing redirect_uris in config")
	}

	log.Printf("OAuth config loaded: ClientID=%s", appConfig.Web.ClientID)
	return nil
}

func initOAuthConfig() {
	redirectURL := appConfig.Web.RedirectURIs[0]
	if len(appConfig.Web.RedirectURIs) > 1 {
		log.Printf("Multiple redirect URIs, using first: %s", redirectURL)
	}

	oauthConfig = &oauth2.Config{
		ClientID:     appConfig.Web.ClientID,
		ClientSecret: appConfig.Web.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  appConfig.Web.AuthURI,
			TokenURL: appConfig.Web.TokenURI,
		},
	}
    if oauthConfig.RedirectURL == "" {
        log.Fatal("OAuth RedirectURL is empty")
    }
    if oauthConfig.Endpoint.AuthURL == "" {
        log.Fatal("OAuth AuthURL is empty")
    }
    if oauthConfig.Endpoint.TokenURL == "" {
        log.Fatal("OAuth TokenURL is empty")
    }
	log.Printf("OAuth initialized: RedirectURL=%s", redirectURL)
}