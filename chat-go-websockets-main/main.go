package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Websocket server started")
	server := NewServer()

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", server.handleWebSocket)
	
	// Endpoints para moderación
	http.HandleFunc("/moderation/badword", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			server.SetModerationStrategy(NewBadWordReplacementStrategy())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Estrategia cambiada a BadWordReplacement"))
		}
	})
	
	http.HandleFunc("/moderation/strict", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			server.SetModerationStrategy(NewStrictBlockingStrategy())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Estrategia cambiada a StrictBlocking"))
		}
	})
	
	http.HandleFunc("/moderation/warning", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			server.SetModerationStrategy(NewWarningStrategy())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Estrategia cambiada a Warning"))
		}
	})
	
	http.HandleFunc("/moderation/composite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			server.SetModerationStrategy(NewCompositeModerationStrategy())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Estrategia cambiada a Composite"))
		}
	})
	
	http.HandleFunc("/moderation/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			stats := server.GetModerationStats()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
		}
	})
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
