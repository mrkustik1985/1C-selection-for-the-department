package main

import (
	client "1C-selection-for-the-department/client/pkg"
	"encoding/json"
	"fmt"
	"log"
	_ "embed"
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
}
