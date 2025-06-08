package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
)

func proxyRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Proxying to backend: %s %s", r.Method, r.URL.Path)

	// Получаем токен из контекста
	idToken, ok := r.Context().Value(idTokenKey).(string)
	if !ok || idToken == "" {
		log.Println("ID token missing in context")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("Setting Authorization header with token: %s...", idToken[:10])
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			log.Printf("Proxying to: %s%s", destinationURL, req.URL.Path)
			req.URL.Scheme = destinationURL.Scheme
			req.URL.Host = destinationURL.Host
			req.Host = destinationURL.Host
			
			req.Header.Set("Authorization", "Bearer "+idToken)

			if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				req.Header.Set("X-Forwarded-For", clientIP)
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			log.Printf("Backend response: %d %s, headers: %v", resp.StatusCode, resp.Request.URL, resp.Header)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Backend error: %v", err)
			http.Error(w, "Backend unavailable", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(w, r)
}