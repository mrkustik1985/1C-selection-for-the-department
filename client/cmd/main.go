package main

import (
	client "1C-selection-for-the-department/client/pkg"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

//go:embed config.json
var configFile []byte

func loadConfig() (client.Config, error) {
	var config client.Config

	if err := json.Unmarshal(configFile, &config); err != nil {
		return config, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	return config, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("ошибка загрузки конфига: %s", err)
	}
	fmt.Printf("Имя: %s, Адрес: %s\n", cfg.Name, cfg.EddieAddress)
	client := client.NewClient(cfg)
	err = client.RegisterPlayer()
	if err != nil {
		log.Fatalf("error in registration player: %s", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/get_game", client.GetGame)
	r.Get("/steps", client.GetSteps)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %s", err)
	}
}
