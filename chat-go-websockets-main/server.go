package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	publisher         *EventPublisher
	logger            *LoggerObserver
	moderationObserver *ModerationObserver
	observerMap       map[string]*ConnectionObserver
	mutex             sync.RWMutex
	nextObserverID    int64
}

func NewServer() *Server {
	publisher := NewEventPublisher()
	logger := NewLoggerObserver()
	statsObserver := NewStatsObserver()
	
	// Crear ModerationObserver con estrategia de reemplazo de malas palabras
	moderationObserver := NewModerationObserver(NewBadWordReplacementStrategy())
	
	// Suscribir observadores a todos los eventos
	publisher.Subscribe(logger)
	publisher.Subscribe(statsObserver)
	publisher.Subscribe(moderationObserver)
	
	// Iniciar el timer de estadísticas cada 30 segundos
	statsObserver.StartStatsTimer(30 * time.Second)
	
	s := &Server{
		publisher:         publisher,
		logger:           logger,
		moderationObserver: moderationObserver,
		observerMap:      make(map[string]*ConnectionObserver),
		nextObserverID:   1,
	}
	
	return s
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

	// Crear un nuevo observador de conexión
	observerID := fmt.Sprintf("obs_%d", s.nextObserverID)
	s.nextObserverID++
	
	observer := NewConnectionObserver(observerID, conn)
	observer.StartListening()
	
	// Registrar el observador
	s.mutex.Lock()
	s.observerMap[observerID] = observer
	s.publisher.Subscribe(observer)
	s.mutex.Unlock()
	
	// Limpiar cuando se desconecte
	defer func() {
		s.mutex.Lock()
		delete(s.observerMap, observerID)
		observer.Stop()
		s.publisher.Unsubscribe(observer)
		s.mutex.Unlock()
	}()

	// Publicar evento de conexión
	s.publisher.PublishEvent(UserJoinEvent, "Usuario conectado", "", map[string]interface{}{
		"observer_id": observerID,
	})

	// Manejar mensajes del cliente
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		
		// Parsear el mensaje JSON del cliente
		var chatMsg ChatMessage
		if err := json.Unmarshal(msg, &chatMsg); err != nil {
			// Si no es JSON válido, tratar como mensaje de texto simple
			chatMsg = ChatMessage{
				Username: observer.GetUsername(),
				Message:  string(msg),
			}
		}
		
		// Actualizar username del observer si se proporciona
		if chatMsg.Username != "" {
			observer.SetUsername(chatMsg.Username)
		}
		
		// Usar la estrategia de moderación centralizada del servidor
		moderationResult := s.moderateMessage(chatMsg.Message)
		
		// Usar el mensaje moderado si fue modificado
		finalMessage := chatMsg.Message
		if moderationResult.Action == "modify" {
			finalMessage = moderationResult.ModifiedMessage
			fmt.Printf("[SERVER] Mensaje moderado: '%s' -> '%s'\n", 
				moderationResult.OriginalMessage, moderationResult.ModifiedMessage)
		} else if moderationResult.Action == "block" {
			// Si el mensaje fue bloqueado, enviar mensaje de sistema al usuario
			s.publisher.PublishEvent(SystemEvent, "Tu mensaje fue bloqueado: "+moderationResult.Reason, "", map[string]interface{}{
				"blocked_message": true,
				"sender_id": observerID,
			})
			continue // No procesar este mensaje
		}
		
		// Publicar evento de mensaje con el contenido final
		s.publisher.PublishEvent(MessageEvent, finalMessage, observer.GetUsername(), map[string]interface{}{
			"chat_message": chatMsg,
			"sender_id": observerID,
			"moderation_result": moderationResult,
		})
	}
	
	// Publicar evento de desconexión
	s.publisher.PublishEvent(UserLeave, "Usuario desconectado", observer.GetUsername(), map[string]interface{}{
		"observer_id": observerID,
	})
}

// Método auxiliar para obtener estadísticas del servidor
func (s *Server) GetConnectionCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.observerMap)
}

// Método para moderar mensajes usando la estrategia centralizada
func (s *Server) moderateMessage(message string) ModerationResult {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if s.moderationObserver == nil {
		// Si no hay moderación, permitir el mensaje
		return ModerationResult{
			OriginalMessage: message,
			ModifiedMessage: message,
			Action:          "allow",
			Reason:          "No moderation configured",
			Confidence:      0.0,
			Timestamp:       time.Now(),
			StrategyUsed:    "none",
		}
	}
	
	// Usar la estrategia del ModerationObserver
	return s.moderationObserver.Moderator.ModerateMessage(message)
}

// Método para cambiar la estrategia de moderación
func (s *Server) SetModerationStrategy(strategy ModerationStrategy) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.moderationObserver.SetStrategy(strategy)
	fmt.Printf("[SERVER] Estrategia de moderación cambiada a: %s\n", strategy.GetName())
}

// Método para obtener estadísticas de moderación
func (s *Server) GetModerationStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.moderationObserver.GetStats()
}

// Método para enviar mensajes de sistema a todos los usuarios conectados
func (s *Server) BroadcastSystemMessage(message string) {
	s.publisher.PublishEvent(SystemEvent, message, "", map[string]interface{}{
		"system_message": true,
	})
}
