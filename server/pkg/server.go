package server

import (
	"encoding/json"
	"fmt"
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
	historyGame  map[Player][]Game
}

func NewServer() *Server {
	return &Server{
		players:     make(map[string]*Player),
		games:       make(map[string]*Game),
		historyGame: make(map[Player][]Game),
	}
}

func (s *Server) RegisterPlayer(w http.ResponseWriter, r *http.Request) {
	var player Player
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
	gameName := r.URL.Query().Get("game")
	playerName := r.URL.Query().Get("player")
	if gameName == "" || playerName == "" {
		http.Error(w, "game and player parameters are required", http.StatusBadRequest)
		return
	}
	s.playersMutex.Lock()
	player, ok := s.players[playerName]
	if !ok {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}
	s.playersMutex.Unlock()

	s.gamesMutex.Lock()
	if _, ok := s.games[gameName]; ok {
		http.Error(w, "game already exists", http.StatusConflict)
		return
	}
	s.gamesMutex.Unlock()

	fmt.Print(fmt.Sprintf("будем начинать игру с %s вторыми?(1 - нет, 0 - да): ", player.Name))
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

// func (s *Server) MakeStepByGame(w http.ResponseWriter, r *http.Request) {

// }
