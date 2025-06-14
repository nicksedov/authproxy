package main

import (
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

type ProfileServer struct {
	config      ProfileConfig
	oauthConfig *oauth2.Config
	destination *url.URL
	mux         *http.ServeMux
}

func NewProfileServer(profile ProfileConfig, oauthConfig *oauth2.Config) *ProfileServer {
	server := &ProfileServer{
		config:      profile,
		oauthConfig: oauthConfig,
		mux:         http.NewServeMux(),
	}

	// Парсим URL назначения
	if profile.Destination != "" {
		var err error
		server.destination, err = url.Parse(profile.Destination)
		if err != nil {
			log.Fatalf("Profile %s: invalid destination URL: %v", profile.Name, err)
		}
	}

	// Регистрируем обработчики
	server.registerHandlers()

	return server
}

func (s *ProfileServer) Mux() *http.ServeMux {
	return s.mux
}

func (s *ProfileServer) registerHandlers() {
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/callback", s.handleCallback)
	s.mux.HandleFunc("/logout", s.handleLogout)

	if s.config.StaticDir != "" {
		log.Printf("Serving static files from: %s", s.config.StaticDir)
		s.mux.Handle("/", s.authMiddleware(http.FileServer(http.Dir(s.config.StaticDir))))
	} else if s.destination != nil {
		log.Println("Proxy mode enabled")
		s.mux.HandleFunc("/", s.handleProxyRoot)
	} else {
		log.Fatalf("Profile %s: must specify either static_dir or destination", s.config.Name)
	}
}