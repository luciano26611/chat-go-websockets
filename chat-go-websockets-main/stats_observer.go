package main

import (
	"fmt"
	"sync"
	"time"
)

// StatsObserver implementa Observer para recopilar estadísticas del chat
type StatsObserver struct {
	id                     string
	totalMessages          int64
	totalUsers             int64
	userConnections        map[string]int64 // contador de conexiones por usuario
	hourlyMessageCount     map[int]int64   // mensajes por hora
	mutex                  sync.RWMutex
	startTime              time.Time
}

func NewStatsObserver() *StatsObserver {
	return &StatsObserver{
		id:                  "stats_observer",
		totalMessages:       0,
		totalUsers:          0,
		userConnections:     make(map[string]int64),
		hourlyMessageCount:  make(map[int]int64),
		startTime:           time.Now(),
	}
}

func (so *StatsObserver) Update(event Event) {
	so.mutex.Lock()
	defer so.mutex.Unlock()

	switch event.Type {
	case MessageEvent:
		so.totalMessages++
		hour := time.Now().Hour()
		so.hourlyMessageCount[hour]++
		
	case UserJoinEvent:
		so.totalUsers++
		if event.Username != "" {
			so.userConnections[event.Username]++
		}
		
	case UserLeave:
		// No decrementamos usuarios, mantenemos el historial
		break
		
	case SystemEvent:
		fmt.Printf("[STATS] System message: %s\n", event.Message)
	}
}

func (so *StatsObserver) GetID() string {
	return so.id
}

// GetStats retorna estadísticas en tiempo real
func (so *StatsObserver) GetStats() map[string]interface{} {
	so.mutex.RLock()
	defer so.mutex.RUnlock()

	return map[string]interface{}{
		"total_messages":      so.totalMessages,
		"total_unique_users": so.totalUsers,
		"uptime_minutes":      time.Since(so.startTime).Minutes(),
		"most_active_users":   so.getMostActiveUsers(),
		"messages_per_hour":   so.hourlyMessageCount,
	}
}

func (so *StatsObserver) getMostActiveUsers() map[string]int64 {
	// Retorna una copia de los usuarios más activos
	result := make(map[string]int64)
	for user, count := range so.userConnections {
		if count > 0 {
			result[user] = count
		}
	}
	return result
}

// PrintStats imprime estadísticas cada cierto tiempo
func (so *StatsObserver) StartStatsTimer(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for range ticker.C {
			stats := so.GetStats()
			fmt.Println("\n=== ESTADÍSTICAS DEL CHAT ===")
			fmt.Printf("Total de mensajes: %v\n", stats["total_messages"])
			fmt.Printf("Usuarios únicos: %v\n", stats["total_unique_users"])
			fmt.Printf("Tiempo activo: %.1f minutos\n", stats["uptime_minutes"])
			fmt.Printf("Usuarios activos: %v\n", stats["most_active_users"])
			fmt.Println("=============================\n")
		}
	}()
}
