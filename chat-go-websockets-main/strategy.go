package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ModerationStrategy define la interfaz para diferentes estrategias de moderación
type ModerationStrategy interface {
	Moderate(message string) ModerationResult
	GetName() string
}

// ModerationResult contiene el resultado del proceso de moderación
type ModerationResult struct {
	OriginalMessage string    `json:"original_message"`
	ModifiedMessage string    `json:"modified_message"`
	Action          string    `json:"action"` // "allow", "modify", "block", "warn"
	Reason          string    `json:"reason"`
	Confidence      float64   `json:"confidence"` // 0.0 - 1.0
	Timestamp       time.Time `json:"timestamp"`
	StrategyUsed    string    `json:"strategy_used"`
}

// ModerationContext maneja las estrategias de moderación
type ModerationContext struct {
	strategy ModerationStrategy
}

func NewModerationContext(strategy ModerationStrategy) *ModerationContext {
	return &ModerationContext{
		strategy: strategy,
	}
}

func (mc *ModerationContext) SetStrategy(strategy ModerationStrategy) {
	mc.strategy = strategy
}

func (mc *ModerationContext) ModerateMessage(message string) ModerationResult {
	if mc.strategy == nil {
		return ModerationResult{
			OriginalMessage: message,
			ModifiedMessage: message,
			Action:          "allow",
			Reason:          "No moderation strategy set",
			Confidence:      0.0,
			Timestamp:       time.Now(),
			StrategyUsed:    "none",
		}
	}
	return mc.strategy.Moderate(message)
}

// BadWordReplacementStrategy reemplaza malas palabras con asteriscos
type BadWordReplacementStrategy struct {
	badWords    []string
	replacement string
}

func NewBadWordReplacementStrategy() *BadWordReplacementStrategy {
	// Lista de palabras que queremos censurar
	badWords := []string{
		"malo", "feo", "tonto", "idiota", "estúpido", "imbécil",
		"odio", "asco", "basura", "mierda", "joder", "puta",
		"cabrón", "hijo de puta", "maldito", "desgraciado",
		// Agregar más palabras según sea necesario
	}
	
	return &BadWordReplacementStrategy{
		badWords:    badWords,
		replacement: "***",
	}
}

func (bwrs *BadWordReplacementStrategy) Moderate(message string) ModerationResult {
	originalMessage := message
	modifiedMessage := message
	action := "allow"
	reason := "No inappropriate content detected"
	confidence := 0.0
	wordsFound := []string{}

	// Convertir a minúsculas para comparación
	lowerMessage := strings.ToLower(message)
	
	for _, badWord := range bwrs.badWords {
		// Usar regex para encontrar palabras completas
		pattern := `\b` + regexp.QuoteMeta(strings.ToLower(badWord)) + `\b`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if regex.MatchString(lowerMessage) {
			wordsFound = append(wordsFound, badWord)
			// Reemplazar la palabra original (manteniendo mayúsculas/minúsculas)
			modifiedMessage = regex.ReplaceAllStringFunc(message, func(match string) string {
				return bwrs.replacement
			})
			action = "modify"
			reason = "Inappropriate words detected and replaced"
			confidence = 0.8
		}
	}

	return ModerationResult{
		OriginalMessage: originalMessage,
		ModifiedMessage: modifiedMessage,
		Action:          action,
		Reason:          reason,
		Confidence:      confidence,
		Timestamp:       time.Now(),
		StrategyUsed:    bwrs.GetName(),
	}
}

func (bwrs *BadWordReplacementStrategy) GetName() string {
	return "BadWordReplacement"
}

// StrictBlockingStrategy bloquea mensajes con contenido inapropiado
type StrictBlockingStrategy struct {
	badWords []string
}

func NewStrictBlockingStrategy() *StrictBlockingStrategy {
	badWords := []string{
		"spam", "scam", "hack", "virus", "malware",
		"phishing", "fraud", "illegal", "drugs",
		// Palabras más severas que requieren bloqueo
	}
	
	return &StrictBlockingStrategy{
		badWords: badWords,
	}
}

func (sbs *StrictBlockingStrategy) Moderate(message string) ModerationResult {
	lowerMessage := strings.ToLower(message)
	
	for _, badWord := range sbs.badWords {
		if strings.Contains(lowerMessage, strings.ToLower(badWord)) {
			return ModerationResult{
				OriginalMessage: message,
				ModifiedMessage: "",
				Action:          "block",
				Reason:          "Message contains prohibited content: " + badWord,
				Confidence:      0.9,
				Timestamp:       time.Now(),
				StrategyUsed:    sbs.GetName(),
			}
		}
	}
	
	return ModerationResult{
		OriginalMessage: message,
		ModifiedMessage: message,
		Action:          "allow",
		Reason:          "Message is clean",
		Confidence:      0.1,
		Timestamp:       time.Now(),
		StrategyUsed:    sbs.GetName(),
	}
}

func (sbs *StrictBlockingStrategy) GetName() string {
	return "StrictBlocking"
}

// WarningStrategy envía advertencias pero permite el mensaje
type WarningStrategy struct {
	warningWords []string
}

func NewWarningStrategy() *WarningStrategy {
	warningWords := []string{
		"violencia", "agresión", "amenaza", "peligro",
		"riesgo", "cuidado", "atención",
	}
	
	return &WarningStrategy{
		warningWords: warningWords,
	}
}

func (ws *WarningStrategy) Moderate(message string) ModerationResult {
	lowerMessage := strings.ToLower(message)
	warnings := []string{}
	
	for _, warningWord := range ws.warningWords {
		if strings.Contains(lowerMessage, strings.ToLower(warningWord)) {
			warnings = append(warnings, warningWord)
		}
	}
	
	if len(warnings) > 0 {
		return ModerationResult{
			OriginalMessage: message,
			ModifiedMessage: message,
			Action:          "warn",
			Reason:          "Message contains warning words: " + strings.Join(warnings, ", "),
			Confidence:      0.6,
			Timestamp:       time.Now(),
			StrategyUsed:    ws.GetName(),
		}
	}
	
	return ModerationResult{
		OriginalMessage: message,
		ModifiedMessage: message,
		Action:          "allow",
		Reason:          "No warnings detected",
		Confidence:      0.2,
		Timestamp:       time.Now(),
		StrategyUsed:    ws.GetName(),
	}
}

func (ws *WarningStrategy) GetName() string {
	return "Warning"
}

// CompositeModerationStrategy combina múltiples estrategias
type CompositeModerationStrategy struct {
	strategies []ModerationStrategy
	name        string
}

func NewCompositeModerationStrategy() *CompositeModerationStrategy {
	return &CompositeModerationStrategy{
		strategies: []ModerationStrategy{
			NewStrictBlockingStrategy(),
			NewBadWordReplacementStrategy(),
			NewWarningStrategy(),
		},
		name: "Composite",
	}
}

func (cms *CompositeModerationStrategy) Moderate(message string) ModerationResult {
	// Ejecutar estrategias en orden de prioridad
	for _, strategy := range cms.strategies {
		result := strategy.Moderate(message)
		
		// Si alguna estrategia bloquea, retornar inmediatamente
		if result.Action == "block" {
			result.StrategyUsed = cms.GetName()
			return result
		}
		
		// Si se modificó el mensaje, usar el mensaje modificado para la siguiente estrategia
		if result.Action == "modify" {
			message = result.ModifiedMessage
		}
	}
	
	// Si llegamos aquí, el mensaje fue procesado por todas las estrategias
	return ModerationResult{
		OriginalMessage: message,
		ModifiedMessage: message,
		Action:          "allow",
		Reason:          "Message passed all moderation checks",
		Confidence:      0.3,
		Timestamp:       time.Now(),
		StrategyUsed:    cms.GetName(),
	}
}

func (cms *CompositeModerationStrategy) GetName() string {
	return cms.name
}

// ModerationObserver implementa Observer para moderar mensajes
type ModerationObserver struct {
	id           string
	moderator    *ModerationContext
	blockedCount int64
	modifiedCount int64
	warningCount int64
}

func NewModerationObserver(strategy ModerationStrategy) *ModerationObserver {
	return &ModerationObserver{
		id:        "moderation_observer",
		moderator: NewModerationContext(strategy),
	}
}

func (mo *ModerationObserver) Update(event Event) {
	if event.Type == MessageEvent {
		// Obtener el mensaje del evento
		message := event.Message
		if message == "" {
			return
		}
		
		// Moderar el mensaje
		result := mo.moderator.ModerateMessage(message)
		
		// Actualizar contadores
		switch result.Action {
		case "block":
			mo.blockedCount++
		case "modify":
			mo.modifiedCount++
		case "warn":
			mo.warningCount++
		}
		
		// Log del resultado de moderación
		fmt.Printf("[MODERATION] %s: %s (Confidence: %.2f)\n", 
			result.Action, result.Reason, result.Confidence)
		
		// Si el mensaje fue modificado, actualizar el evento
		if result.Action == "modify" && result.ModifiedMessage != result.OriginalMessage {
			// Aquí podrías actualizar el evento con el mensaje modificado
			fmt.Printf("[MODERATION] Modified: '%s' -> '%s'\n", 
				result.OriginalMessage, result.ModifiedMessage)
		}
	}
}

func (mo *ModerationObserver) GetID() string {
	return mo.id
}

func (mo *ModerationObserver) SetStrategy(strategy ModerationStrategy) {
	mo.moderator.SetStrategy(strategy)
}

func (mo *ModerationObserver) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"blocked_messages":  mo.blockedCount,
		"modified_messages": mo.modifiedCount,
		"warning_messages":  mo.warningCount,
		"strategy":          mo.moderator.strategy.GetName(),
	}
}
