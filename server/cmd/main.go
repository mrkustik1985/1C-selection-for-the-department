package main

import (
	server "1C-selection-for-the-department/server/pkg"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	port = 8080
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	server := server.NewServer()
	
	// одключаем обработчики
	r.Post("/register", server.RegisterPlayer)
	r.Get("/suggest_game", server.SuggestGame)
	r.Get("/history_step_by_name", server.GetHistoryByName)
	r.Get("/get_game_not_finished", server.GetGameNotFinished)
	r.Get("/make_step", server.MakeStepByGame)
	// запускаем сервер на порту localhost 8080
	log.Printf("server start on port: %d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
