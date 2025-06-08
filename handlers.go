package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
)

func handleProxyRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling root request: %s %s", r.Method, r.URL.Path)
	session, err := getSession(r)
	if err != nil {
		if welcomePage != "" {
			showWelcomePage(w, r)
			return
		}
		startOAuthFlow(w, r)
		return
	}

	log.Printf("Valid session, proxying to %s", destinationURL)
	ctx := context.WithValue(r.Context(), idTokenKey, session.IDToken)
	proxyRequest(w, r.WithContext(ctx))
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
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
	token, err := oauthConfig.Exchange(context.Background(), code)
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

	if err := saveSession(w, session); err != nil {
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

func handleLogout(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling logout request")
	clearSession(w)
	targetPage := welcomePage;
	if targetPage == "" {
		targetPage = "/"
	}
	http.Redirect(w, r, targetPage, http.StatusFound)
}

func staticHandler() http.Handler {
	return http.FileServer(http.Dir(staticDir))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("Initiating OAuth flow from login page")
	startOAuthFlow(w, r)
}

func showWelcomePage(w http.ResponseWriter, r *http.Request) {
	// Динамическая загрузка welcome page
	if welcomePage == "" {
		http.Error(w, "Welcome page not configured", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(welcomePage)
	if err != nil {
		log.Printf("Error reading welcome page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}