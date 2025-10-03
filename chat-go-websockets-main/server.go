package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	connections map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection", err)
		return
	}
	defer conn.Close()

	addConnection(s, conn)
	defer removeConnection(s, conn)

	log.Println("Client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("Received message: %s\n", msg)
		broadcastMessage(s, msg, conn)
	}
}

func broadcastMessage(s *Server, message []byte, sender *websocket.Conn) {
	for conn := range s.connections {
		if sender != conn {
			go func(conn *websocket.Conn) {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("Error broadcasting message", err)
					conn.Close()
					delete(s.connections, conn)
				}
			}(conn)
		}
	}
}

func addConnection(s *Server, conn *websocket.Conn) {
	s.connections[conn] = true
}

func removeConnection(s *Server, conn *websocket.Conn) {
	delete(s.connections, conn)
}
