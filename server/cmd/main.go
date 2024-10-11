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
	// запускаем сервер на порту localhost 8080
	log.Printf("server start on port: %d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
