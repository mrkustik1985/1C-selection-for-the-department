package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) SuggestGame(w http.ResponseWriter, r *http.Request) {
	gameName := r.URL.Query().Get("game")
	playerName := r.URL.Query().Get("player")
	if gameName == "" || playerName == "" {
		http.Error(w, "game and player необходимы в запросе", http.StatusBadRequest)
		return
	}
	s.playersMutex.Lock()
	player, ok := s.players[playerName]
	if !ok {
		http.Error(w, "игрок не найден", http.StatusNotFound)
		return
	}
	s.playersMutex.Unlock()

	s.gamesMutex.Lock()
	if _, ok := s.games[gameName]; ok {
		http.Error(w, "игра уже существует", http.StatusConflict)
		return
	}
	s.gamesMutex.Unlock()

	fmt.Print(fmt.Sprintf("будем начинать игру с игроком %s вторыми?(1 - да, 0 - нет): ", player.Name))
	var is_need_start string
	fmt.Scanln(&is_need_start)
	// Формирование URL для запроса к игроку
	url := fmt.Sprintf("http://%s/get_game?game=%s&is_need_start=%s", player.Address, gameName, is_need_start)

	// Отправка запроса к игроку
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка отправки запроса к игроку: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			http.Error(w, fmt.Sprintf("неожиданный статус ответа от игрока с ошибкой: %+v", errorResponse), http.StatusBadRequest)
			return
		}

		errorMessage, ok := errorResponse["error"].(string)
		if !ok {
			http.Error(w, fmt.Sprintf("неожиданный статус ответа от игрока с ошибкой: %s", errorMessage), http.StatusBadRequest)
			return
		}

		http.Error(w, fmt.Sprintf("ошибка от игрока: %s", errorMessage), http.StatusBadRequest)
		return
	}

	// Чтение ответа от игрока
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		http.Error(w, fmt.Sprintf("ошибка декодирования ответа от игрока: %v", err), http.StatusInternalServerError)
		return
	}

	if response["is_start"] == "1" {
		s.gamesMutex.Lock()
		s.games[gameName] = &Game{
			ID:     gameName,
			Player: player,
		}
		if step, ok := response["step"].(string); ok {
			s.games[gameName].Moves = append(s.games[gameName].Moves, step)
		}
		s.gamesMutex.Unlock()
	} else {
		http.Error(w, "игра не начата", http.StatusBadRequest)
		return
	}

	// Проверка ответа от игрока(больше для дебагга)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}
