package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling request: %s...", r.RequestURI)
		// Разрешаем доступ без аутентификации
		if r.URL.Path == "/callback" || r.URL.Path == "/logout" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Показываем welcome page если она задана и пользователь не аутентифицирован
		session, err := getSession(r)
		if err != nil {
			if welcomePage != "" {
				showWelcomePage(w, r)
				return
			}
			startOAuthFlow(w, r)
			return
		}
		log.Printf("Setting Authorization header with token: %s...", session.IDToken[:10])
		//r.Header.Set("Authorization", "Bearer "+session.IDToken)
		w.Header().Add("Authorization", "Bearer "+session.IDToken)
		next.ServeHTTP(w, r)
	})
}

func startOAuthFlow(w http.ResponseWriter, r *http.Request) {
	targetEndpoint := r.URL.RequestURI()
    log.Printf("Starting OAuth flow for request: %s", targetEndpoint)
	if strings.ToLower(targetEndpoint) == "/login" {
		targetEndpoint = "/"
	}
	state := base64.URLEncoding.EncodeToString([]byte(targetEndpoint))
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    log.Printf("Generated OAuth URL: %s", authURL)
    http.Redirect(w, r, authURL, http.StatusFound)
}