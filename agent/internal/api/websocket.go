package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"sysagent/internal/auth"
	"sysagent/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.RemoteAddr == "127.0.0.1" || r.Host == "localhost" || r.Host == "127.0.0.1"
	},
}

type WSServer struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	commandCh chan models.ControlCommand
}

func NewWSServer(commandCh chan models.ControlCommand) *WSServer {
	return &WSServer{
		clients:   make(map[*websocket.Conn]bool),
		commandCh: commandCh,
	}
}

func (server *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		tokenStr = r.Header.Get("Authorization")
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	_, err := auth.ValidateToken(tokenStr)
	if err != nil {
		log.Println("Unauthorized connection attempt:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	server.clientsMu.Lock()
	server.clients[conn] = true
	server.clientsMu.Unlock()

	defer func() {
		server.clientsMu.Lock()
		delete(server.clients, conn)
		server.clientsMu.Unlock()
		conn.Close()
	}()

	for {
		var cmd models.ControlCommand
		err := conn.ReadJSON(&cmd)
		if err != nil {
			log.Println("Reading error from WS client:", err)
			break
		}
		select {
		case server.commandCh <- cmd:
		case <-time.After(1 * time.Second):
			log.Println("Warning: Dropped command - processing channel is full")
		}
	}
}

func (server *WSServer) BroadcastMetrics(metrics []models.AppMetrics) {
	if len(metrics) == 0 {
		return
	}

	server.clientsMu.Lock()
	defer server.clientsMu.Unlock()

	data, err := json.Marshal(map[string]interface{}{
		"type":    "metrics",
		"payload": metrics,
	})
	if err != nil {
		return
	}

	for conn := range server.clients {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("Write error:", err)
			conn.Close()
			delete(server.clients, conn)
		}
	}
}
