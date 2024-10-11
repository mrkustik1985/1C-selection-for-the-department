package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

	// Проверка на победителя
	winner := checkGame(game.Moves)
	if winner == "try_add_where_stand" {
		http.Error(w, "Попробуйте добавить куда-то в другое место", http.StatusBadRequest)
		return
	}
	if winner != "nobody" {
		game.IsFinished = true
	}

	// Отправка успешного ответа
	url := fmt.Sprintf("http://%s/do_step", game.Player.Address)

	// Формирование данных для запроса
	stepData := map[string]string{
		"step":        fmt.Sprintf("%s krest %s %s", "Eddi", coord1, coord2),
		"is_finished": strconv.FormatBool(game.IsFinished),
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

	fmt.Print("получили ответ от игрока playerstep")


	// Извлечение шага из ответа игрока
	playerStep, ok := playerResponse["step"].(string)
	if !ok {
		http.Error(w, "не удалось извлечь шаг из ответа игрока", http.StatusInternalServerError)
		return
	}
	game.Moves = append(game.Moves, playerStep)
	winner = checkGame(game.Moves)
	if winner == "try_add_where_stand" {
		http.Error(w, "Попробуйте добавить куда-то в другое место", http.StatusBadRequest)
		return
	}
	if winner != "nobody" {
		game.IsFinished = true
	}

	// Обновление игры в карте
	s.gamesMutex.Lock()
	if game.IsFinished {
		delete(s.games, gameName)
		s.historyGame[*game.Player] = append(s.historyGame[*game.Player], *game)
	} else {
		s.games[gameName] = game
	}
	s.gamesMutex.Unlock()

	response := map[string]string{
		"message": fmt.Sprintf("Шаг выполнен в игре %s игроком %s", gameName, game.Player.Name),
	}
	if game.IsFinished {
		response["winner"] = "игра закончена! победил:" + winner
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("ошибка кодирования ответа: %v", err), http.StatusInternalServerError)
		return
	}
}
