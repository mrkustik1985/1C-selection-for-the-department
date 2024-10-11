package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Game struct {
	ID         string   `json:"id"`
	Moves      []string `json:"moves"`
	IsFinished bool     `json:"is_finished"`
}

type Client struct {
	AddressEddie string
	game         Game
	Name         string
	MyAddr       string
}

type Player struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Config struct {
	Name         string `json:"name"`
	EddieAddress string `json:"eddieAddress"`
	Port         int    `json:"port"`
}

func NewClient(cfg Config) *Client {
	return &Client{
		AddressEddie: cfg.EddieAddress,
		Name:         cfg.Name,
		MyAddr:       fmt.Sprintf("localhost:%d", cfg.Port),
	}
}

func (c *Client) GetReq(req string) string {
	return "http://" + c.AddressEddie + "/" + req
}

func (c *Client) RegisterPlayer() error {
	player := Player{Name: c.Name, Address: c.MyAddr}
	jsonData, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}
	log.Printf("addres request: %s", c.GetReq("register"))
	resp, err := http.Post(c.GetReq("register"), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный статус ответа: %v", resp.Status)
	}

	fmt.Printf("Игрок зарегистрирован: %+v\n", c)
	return nil
}

func (c *Client) GetGame(w http.ResponseWriter, r *http.Request) {
	gameName := r.URL.Query().Get("game")
	if gameName == "" {
		http.Error(w, "game parameter is required", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("%s?game=%s", c.GetReq("game"), gameName)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка отправки запроса: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("неожиданный статус ответа: %v", resp.Status), http.StatusInternalServerError)
		return
	}

	var gameProposal map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gameProposal); err != nil {
		http.Error(w, fmt.Sprintf("ошибка декодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(gameProposal); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

func (c *Client) GetSteps(w http.ResponseWriter, r *http.Request) {
	var steps string
	for _, move := range c.game.Moves {
		steps += move + " "
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(steps))
}
