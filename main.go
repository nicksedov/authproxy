// main.go
package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

var (
	oauthConfig    *oauth2.Config
	destinationURL *url.URL
	appConfig      Config
	staticDir      string
	welcomePage    string
)

func main() {
	destinationURLStr := flag.String("destination", "", "Destination URL")
	port := flag.String("port", "8080", "Port to listen on")
	configFile := flag.String("config", "config.json", "Path to config file")
	static := flag.String("static", "", "Directory to serve static files from")
	welcomeFile := flag.String("welcome", "", "Path to welcome page HTML file (overrides config)")
	flag.Parse()

	staticDir = *static

	log.Printf("Starting IAM Proxy: port=%s, config=%s, destination=%s, static=%s, welcome=%s",
		*port, *configFile, *destinationURLStr, staticDir, *welcomeFile)

	if *destinationURLStr == "" && staticDir == "" {
		log.Fatal("Either destination URL or static directory must be specified")
	}

	if *destinationURLStr != "" {
		var err error
		destinationURL, err = url.Parse(*destinationURLStr)
		if err != nil {
			log.Fatalf("Invalid destination URL: %v", err)
		}
	}

	if err := loadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Обработка welcome page (флаг имеет приоритет над конфигом)
	if *welcomeFile != "" {
		if _, err := os.Stat(*welcomeFile); err == nil {
			welcomePage = *welcomeFile
			log.Printf("Using welcome page from flag: %s", *welcomeFile)
		} else {
			log.Printf("Warning: welcome page not found at %s", *welcomeFile)
		}
	}

	initOAuthConfig()

	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/logout", handleLogout)

	if staticDir != "" {
		log.Printf("Serving static files from: %s", staticDir)
		http.Handle("/", authMiddleware(staticHandler()))
	} else {
		log.Println("Proxy mode enabled")
		http.HandleFunc("/", handleProxyRoot)
	}

	log.Printf("IAM Proxy listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}