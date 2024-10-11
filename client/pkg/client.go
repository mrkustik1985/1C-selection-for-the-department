package client

import (
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

func (c *Client) DoStep(w http.ResponseWriter, r *http.Request) {
	if c.game.IsFinished == true {
		http.Error(w, "Игра завершена", http.StatusBadRequest)
		return
	}
	// Извлечение и парсинг поля step из запроса
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		fmt.Print(err)
		http.Error(w, fmt.Sprintf("ошибка декодирования запроса: %v", err), http.StatusBadRequest)
		return
	}
	is_finished := requestData["is_finished"]
	if is_finished == "true" {
		c.game.IsFinished = true
		return
	}
	step, ok := requestData["step"]
	if !ok {
		http.Error(w, "поле step отсутствует в запросе", http.StatusBadRequest)
		return
	}
	c.game.Moves = append(c.game.Moves, step)

	var coord1, coord2 string
	fmt.Print("Введите ваш ход (например, '1 1'): ")
	fmt.Scanln(&coord1, &coord2)
	response := map[string]string{
		"step": c.Name + " zero " + coord1 + " " + coord2,
	}
	c.game.Moves = append(c.game.Moves, response["step"])
	winner := checkGame(c.game.Moves)
	log.Println("winner: ", winner)
	if winner == "try_add_where_stand" {
		fmt.Print("Попробуйте добавить куда-то в другое место")
	}
	if winner != "nobody" {
		c.game.IsFinished = true
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}

func (c *Client) GetSteps(w http.ResponseWriter, r *http.Request) {
	var steps string
	for _, move := range c.game.Moves {
		steps += move + "\n"
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(steps))
}
