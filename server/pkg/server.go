package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Player struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Game struct {
	ID         string   `json:"id"`
	Player     *Player  `json:"player1"`
	Moves      []string `json:"moves"`
	IsFinished bool     `json:"is_finished"`
}

type Server struct {
	players      map[string]*Player
	games        map[string]*Game
	gamesMutex   sync.Mutex
	playersMutex sync.Mutex
}

func NewServer() *Server {
	return &Server{
		players: make(map[string]*Player),
		games:   make(map[string]*Game),
	}
}

func (s *Server) RegisterPlayer(w http.ResponseWriter, r *http.Request) {
	var player Player
	log.Println("prishel req")
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		log.Println("bad req")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("register player: %+v", player)
	s.playersMutex.Lock()
	s.players[player.Name] = &player
	s.playersMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(player)
}

func (s *Server) SuggestGame(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.gamesMutex.Lock()
	s.games[game.ID] = &game
	s.gamesMutex.Unlock()

	// Send game start request to the player
	player := game.Player
	if player != nil {
		jsonData, err := json.Marshal(game)
		if err != nil {
			log.Printf("error marshaling game data: %v", err)
			return
		}
		resp, err := http.Post("http://"+player.Address+"/start-game", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("error sending game start request: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("unexpected response status: %v", resp.Status)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(game)
}
