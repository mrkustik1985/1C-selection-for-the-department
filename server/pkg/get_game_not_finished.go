package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) GetGameNotFinished(w http.ResponseWriter, r *http.Request) {
	type GameInfo struct {
		GameName   string `json:"game_name"`
		PlayerName string `json:"player_name"`
	}

	unfinishedGames := []GameInfo{}

	s.gamesMutex.Lock()
	for gameName, game := range s.games {
		if !game.IsFinished {
			unfinishedGames = append(unfinishedGames, GameInfo{
				GameName:   gameName,
				PlayerName: game.Player.Name,
			})
		}
	}
	s.gamesMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(unfinishedGames); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}
