package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) GetHistoryByName(w http.ResponseWriter, r *http.Request) {
	playerName := r.URL.Query().Get("player")
	if playerName == "" {
		http.Error(w, "имя игрока необходимо(параметр player)", http.StatusBadRequest)
		return
	}

	s.playersMutex.Lock()
	player, ok := s.players[playerName]
	if !ok {
		http.Error(w, "игрок не найден", http.StatusNotFound)
		return
	}
	s.playersMutex.Unlock()

	history := s.historyGame[*player]

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}
