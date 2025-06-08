package main

import (
	"log"
	"net/http"
	"time"
)

type Session struct {
	IDToken   string    `json:"id_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func getSession(r *http.Request) (Session, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
        log.Printf("Session not found: %v", err)
        return Session{}, err
    }
    log.Printf("Session found: %s", cookie.Value[:10]+"...") // Логируем часть токена
    return Session{IDToken: cookie.Value}, nil
}

func saveSession(w http.ResponseWriter, session Session) error {
	log.Printf("Saving session, expires at: %v", session.ExpiresAt)
	cookie := &http.Cookie{
		Name:     "session",
		Value:    session.IDToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
	http.SetCookie(w, cookie)
	return nil
}

func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}