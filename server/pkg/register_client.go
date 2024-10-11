package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) RegisterPlayer(w http.ResponseWriter, r *http.Request) {
	var player Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		log.Println("bad req")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("register player: %+v", player)
	s.playersMutex.Lock()
	s.players[player.Name] = &player
	s.playersMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(player)
}
