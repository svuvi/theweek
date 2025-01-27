package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/svuvi/theweek/db"
	"github.com/svuvi/theweek/middleware"
	"github.com/svuvi/theweek/routes"
)

func main() {
	db := db.ConnectDB()
	defer db.Close()

	h := routes.NewBaseHandler(db)
	router := middleware.NewLogger(h.NewRouter())

	port := flag.Int("port", 8080, "port number")

	server := http.Server{
		Addr:    fmt.Sprint(":", *port),
		Handler: router,
	}

	log.Printf("Запускаю сервер http://localhost%s", server.Addr)
	server.ListenAndServe()
}
