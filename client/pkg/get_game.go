package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (c *Client) GetGame(w http.ResponseWriter, r *http.Request) {
	gameName := r.URL.Query().Get("game")

	if gameName == "" {
		http.Error(w, "game parameter is required", http.StatusBadRequest)
		return
	}
	startGame := r.URL.Query().Get("is_need_start")
	if startGame == "" {
		startGame = "0"
	}

	fmt.Print("Будем начинать игру с Эдди?(1 - да, 0 - нет): ")
	var strt int
	fmt.Scanln(&strt)
	if strt == 0 {
		log.Println("Игра не начата")
		http.Error(w, "Прости, пока не хочу начинать игру", http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"message":  fmt.Sprintf("Cоглашаюсь на начало игры %s", gameName),
		"is_start": "1",
		"step":     "",
	}

	if startGame == "1" {
		var coord1, coord2 string
		fmt.Print("Введите ваш ход (например, '1 1'): ")
		fmt.Scanln(&coord1, &coord2)
		response["step"] = c.Name + " zero " + coord1 + " " + coord2
		c.game.Moves = append(c.game.Moves, response["step"])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}
