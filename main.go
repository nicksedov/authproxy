package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"gopkg.in/yaml.v3"
	"os"
)

type ProfileConfig struct {
	Name        string `yaml:"name"`
	Port        string `yaml:"port"`
	PublicURL   string `yaml:"public_url"`
	OAuthConfig string `yaml:"oauth_config"`
	Destination string `yaml:"destination"`
	StaticDir   string `yaml:"static_dir"`
	WelcomePage string `yaml:"welcome_page"`
}

type AppConfig struct {
	Profiles []ProfileConfig `yaml:"profiles"`
}

func main() {
	configFile := flag.String("config", "profiles.yaml", "Path to YAML config file")
	flag.Parse()

	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var appConfig AppConfig
	if err := yaml.Unmarshal(configData, &appConfig); err != nil {
		log.Fatalf("Failed to parse YAML config: %v", err)
	}

	var wg sync.WaitGroup
	for _, profile := range appConfig.Profiles {
		wg.Add(1)
		go func(p ProfileConfig) {
			defer wg.Done()
			startProfileServer(p)
		}(profile)
	}
	wg.Wait()
}

func startProfileServer(profile ProfileConfig) {
	log.Printf("Starting profile '%s' on port %s", profile.Name, profile.Port)
	
	// Загружаем OAuth конфиг из JSON-файла
	oauthConfig, err := loadOAuthConfig(profile.OAuthConfig)
	if err != nil {
		log.Fatalf("Profile %s: failed to load OAuth config: %v", profile.Name, err)
	}

	// Создаем сервер для профиля
	server := NewProfileServer(profile, oauthConfig)
	
	if err := http.ListenAndServe(":"+profile.Port, server.Mux()); err != nil {
		log.Fatalf("Profile '%s': server failed: %v", profile.Name, err)
	}
}