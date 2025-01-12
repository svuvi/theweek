package main

import (
	"log"
	"net/http"

	"github.com/svuvi/theweek/db"
	"github.com/svuvi/theweek/routes"
)

func main() {
	db := db.ConnectDB()
	defer db.Close()

	h := routes.NewBaseHandler(db)
	router := h.NewRouter()

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Printf("Запускаю сервер http://localhost%s", server.Addr)
	server.ListenAndServe()
}
