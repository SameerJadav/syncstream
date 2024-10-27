package main

import (
	"net/http"

	"github.com/SameerJadav/syncstream/internal/handler"
	"github.com/SameerJadav/syncstream/internal/logger"
	"github.com/SameerJadav/syncstream/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", handler.ServeStaticFiles())

	mux.HandleFunc("POST /rooms", handler.CreateRoom)
	mux.HandleFunc("/rooms/{id}", handler.JoinRoom)
	mux.HandleFunc("/ws/{id}", handler.UpgradeConnection)

	logger.Info.Println("http://localhost:8080")
	if err := http.ListenAndServe(":8080", middleware.LogRequest(mux)); err != nil {
		logger.Error.Fatal(err)
	}
}
