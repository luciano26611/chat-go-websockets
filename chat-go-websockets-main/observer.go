package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// EventType representa los diferentes tipos de eventos en el chat
type EventType string

const (
	MessageEvent    EventType = "message"
	UserJoinEvent   EventType = "user_join"
	UserLeave EventType = "user_leave"
	SystemEvent     EventType = "system"
)

// Event representa un evento genérico en el sistema
type Event struct {
	Type      EventType           `json:"type"`
	Message   string              `json:"message,omitempty"`
	Username  string              `json:"username,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
}

// Observer es la interfaz que deben implementar todos los observadores
type Observer interface {
	Update(event Event)
	GetID() string
}

// Publisher es la interfaz para el sujeto observable
type Publisher interface {
	Subscribe(observer Observer)
	Unsubscribe(observer Observer)
	Notify(event Event)
}

// ChatMessage representa un mensaje del chat con información adicional
type ChatMessage struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Room     string    `json:"room,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ConnectionObserver implementa Observer para manejar conexiones WebSocket
type ConnectionObserver struct {
	id         string
	conn       *websocket.Conn
	username   string
	sendChan   chan Event
	closeChan  chan bool
}

func NewConnectionObserver(id string, conn *websocket.Conn) *ConnectionObserver {
	return &ConnectionObserver{
		id:        id,
		conn:      conn,
		sendChan:  make(chan Event, 100),
		closeChan: make(chan bool),
	}
}

func (co *ConnectionObserver) Update(event Event) {
	select {
	case co.sendChan <- event:
	default:
		// Si el canal está lleno, log pero no bloquea
		fmt.Printf("Warning: Notification queue full for observer %s\n", co.id)
	}
}

func (co *ConnectionObserver) GetID() string {
	return co.id
}

func (co *ConnectionObserver) SetUsername(username string) {
	co.username = username
}

func (co *ConnectionObserver) GetUsername() string {
	return co.username
}

// StartListening inicia el proceso de escucha para enviar mensajes al cliente
func (co *ConnectionObserver) StartListening() {
	go func() {
		for {
			select {
			case event := <-co.sendChan:
				co.sendEventToClient(event)
			case <-co.closeChan:
				return
			}
		}
	}()
}

func (co *ConnectionObserver) sendEventToClient(event Event) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		return
	}

	err = co.conn.WriteMessage(websocket.TextMessage, eventBytes)
	if err != nil {
		fmt.Printf("Error sending message to client: %v\n", err)
		co.Stop()
	}
}

// Stop cierra el observador de conexión
func (co *ConnectionObserver) Stop() {
	select {
	case co.closeChan <- true:
	default:
	}
	close(co.sendChan)
}

// EventPublisher implementa Publisher para manejar notificaciones
type EventPublisher struct {
	observers map[string]Observer
	eventChan chan Event
}

func NewEventPublisher() *EventPublisher {
	ep := &EventPublisher{
		observers: make(map[string]Observer),
		eventChan: make(chan Event, 1000),
	}
	
	// Inicia el proceso de publicación en una goroutine
	go ep.processEvents()
	
	return ep
}

func (ep *EventPublisher) Subscribe(observer Observer) {
	ep.observers[observer.GetID()] = observer
}

func (ep *EventPublisher) Unsubscribe(observer Observer) {
	delete(ep.observers, observer.GetID())
}

func (ep *EventPublisher) Notify(event Event) {
	select {
	case ep.eventChan <- event:
	default:
		fmt.Printf("Warning: Event queue full\n")
	}
}

// PublishEvent es un método conveniente para publicar eventos
func (ep *EventPublisher) PublishEvent(eventType EventType, message, username string, data map[string]interface{}) {
	event := Event{
		Type:      eventType,
		Message:   message,
		Username:  username,
		Data:      data,
		Timestamp: time.Now(),
	}
	ep.Notify(event)
}

func (ep *EventPublisher) processEvents() {
	for event := range ep.eventChan {
		for _, observer := range ep.observers {
			go observer.Update(event)
		}
	}
}

// LoggerObserver implementa Observer para logging de eventos
type LoggerObserver struct {
	id string
}

func NewLoggerObserver() *LoggerObserver {
	return &LoggerObserver{id: "logger"}
}

func (lo *LoggerObserver) Update(event Event) {
	formattedTime := event.Timestamp.Format("15:04:05")
	fmt.Printf("[%s] %s: %s", formattedTime, string(event.Type), event.Message)
	if event.Username != "" {
		fmt.Printf(" (de: %s)", event.Username)
	}
	fmt.Printf("\n")
}

func (lo *LoggerObserver) GetID() string {
	return lo.id
}
