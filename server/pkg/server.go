package server

import (
	"bytes"
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

func (s *Server) MakeStepByGame(w http.ResponseWriter, r *http.Request) {
	// Получение параметров из запроса
	gameName := r.URL.Query().Get("game")

	if gameName == "" {
		http.Error(w, "передайте параметр game", http.StatusBadRequest)
		return
	}

	// Проверка наличия игры
	s.gamesMutex.Lock()
	game, ok := s.games[gameName]
	if !ok {
		s.gamesMutex.Unlock()
		http.Error(w, "игра не найдена", http.StatusNotFound)
		return
	}
	s.gamesMutex.Unlock()

	// Логика для выполнения шага
	var coord1, coord2 string
	fmt.Print("Введите ваш ход (например, '1 1'): ")
	fmt.Scanln(&coord1, &coord2)

	game.Moves = append(game.Moves, fmt.Sprintf("%s krest %s %s", game.Player.Name, coord1, coord2))

	// Отправка успешного ответа

	url := fmt.Sprintf("http://%s/do_step", game.Player.Address)

	// Формирование данных для запроса
	stepData := map[string]string{
		"step": fmt.Sprintf("%s krest %s %s", "Eddi", coord1, coord2),
	}

	jsonData, err := json.Marshal(stepData)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка маршалинга JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправка POST-запроса к игроку и ожидание ответа от него(что он тоже сделал шаг)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка отправки запроса к игроку: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	fmt.Print("получили ответ от игрока")
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			http.Error(w, fmt.Sprintf("неожиданный статус ответа от игрока: %v", resp.Status), http.StatusInternalServerError)
			return
		}

		errorMessage, ok := errorResponse["error"].(string)
		if !ok {
			http.Error(w, fmt.Sprintf("неожиданный статус ответа от игрока: %v", resp.Status), http.StatusInternalServerError)
			return
		}

		http.Error(w, fmt.Sprintf("ошибка от игрока: %s", errorMessage), http.StatusInternalServerError)
		return
	}

	// Чтение ответа от игрока
	var playerResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&playerResponse); err != nil {
		http.Error(w, fmt.Sprintf("ошибка декодирования ответа от игрока: %v", err), http.StatusInternalServerError)
		return
	}

	// Извлечение шага из ответа игрока
	playerStep, ok := playerResponse["step"].(string)
	if !ok {
		http.Error(w, "не удалось извлечь шаг из ответа игрока", http.StatusInternalServerError)
		return
	}
	game.Moves = append(game.Moves, playerStep)

	// Обновление игры в карте

	s.gamesMutex.Lock()
	s.games[gameName] = game
	s.gamesMutex.Unlock()

	response := map[string]string{
		"message": fmt.Sprintf("Шаг выполнен в игре %s игроком %s", gameName, game.Player.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}
