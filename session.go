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

func (s *ProfileServer) getSession(r *http.Request) (Session, error) {
	cookieName := "session_" + s.config.Name
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		log.Printf("Session not found: %v", err)
		return Session{}, err
	}
	return Session{IDToken: cookie.Value}, nil
}

func (s *ProfileServer) saveSession(w http.ResponseWriter, session Session) error {
	cookieName := "session_" + s.config.Name
	cookie := &http.Cookie{
		Name:     cookieName,
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

func (s *ProfileServer) clearSession(w http.ResponseWriter) {
	cookieName := "session_" + s.config.Name
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}