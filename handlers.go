package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
)

func (s *ProfileServer) handleProxyRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling root request: %s %s", r.Method, r.URL.Path)
	session, err := s.getSession(r)
	if err != nil {
		if s.config.WelcomePage != "" {
			s.showWelcomePage(w, r)
			return
		}
		s.startOAuthFlow(w, r)
		return
	}

	log.Printf("Valid session, proxying to %s", s.destination)
	ctx := context.WithValue(r.Context(), idTokenKey, session.IDToken)
	s.proxyRequest(w, r.WithContext(ctx))
}

func (s *ProfileServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling OAuth callback")

	if err := r.URL.Query().Get("error"); err != "" {
		errorDesc := r.URL.Query().Get("error_description")
		log.Printf("OAuth error: %s - %s", err, errorDesc)
		http.Error(w, "OAuth error: "+err, http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("Missing authorization code")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	log.Println("Exchanging authorization code for token")
	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Println("ID token missing in OAuth response")
		http.Error(w, "No id_token in OAuth response", http.StatusInternalServerError)
		return
	}

	session := Session{
		IDToken:   idToken,
		ExpiresAt: token.Expiry,
	}

	if err := s.saveSession(w, session); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	redirectPath := "/"
    if state := r.URL.Query().Get("state"); state != "" {
        log.Printf("Received state parameter: %s", state)
        if decodedState, err := base64.URLEncoding.DecodeString(state); err == nil {
            redirectPath = string(decodedState)
            log.Printf("Decoded redirect path: %s", redirectPath)
        } else {
            log.Printf("Error decoding state: %v", err)
        }
    } else {
        log.Println("State parameter is empty")
    }
    
    log.Printf("Redirecting to: %s", redirectPath)
    http.Redirect(w, r, redirectPath, http.StatusFound)
}

func (s *ProfileServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling logout request")
	s.clearSession(w)
	targetPage := s.config.WelcomePage;
	if targetPage == "" {
		targetPage = "/"
		http.Redirect(w, r, targetPage, http.StatusFound)
	} else {
		s.showWelcomePage(w, r)
	}
}

func (s *ProfileServer) staticHandler() http.Handler {
	return http.FileServer(http.Dir(s.config.StaticDir))
}

func (s *ProfileServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("Initiating OAuth flow from login page")
	s.startOAuthFlow(w, r)
}

func (s *ProfileServer) showWelcomePage(w http.ResponseWriter, r *http.Request) {
	// Динамическая загрузка welcome page
	if s.config.WelcomePage == "" {
		http.Error(w, "Welcome page not configured", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(s.config.WelcomePage)
	if err != nil {
		log.Printf("Error reading welcome page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}