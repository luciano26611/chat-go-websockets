# Implementación del Patrón Strategy para Moderación de Contenido

## Descripción General

Se ha implementado el patrón Strategy para manejar diferentes estrategias de moderación de contenido en el chat. Este patrón permite cambiar dinámicamente el comportamiento de moderación sin modificar el código del servidor.

## Arquitectura del Patrón Strategy

### 1. Interfaz Strategy

```go
type ModerationStrategy interface {
    Moderate(message string) ModerationResult
    GetName() string
}
```

Define el contrato que todas las estrategias de moderación deben implementar.

### 2. Context (ModerationContext)

```go
type ModerationContext struct {
    strategy ModerationStrategy
}
```

Maneja las estrategias y proporciona una interfaz unificada para moderar mensajes.

### 3. Resultado de Moderación

```go
type ModerationResult struct {
    OriginalMessage string    `json:"original_message"`
    ModifiedMessage string    `json:"modified_message"`
    Action          string    `json:"action"` // "allow", "modify", "block", "warn"
    Reason          string    `json:"reason"`
    Confidence      float64   `json:"confidence"` // 0.0 - 1.0
    Timestamp       time.Time `json:"timestamp"`
    StrategyUsed    string    `json:"strategy_used"`
}
```

Contiene toda la información sobre el proceso de moderación.

## Estrategias Implementadas

### 1. BadWordReplacementStrategy

**Propósito**: Reemplaza malas palabras con asteriscos (***)

**Características**:
- Detecta palabras completas usando regex
- Mantiene mayúsculas/minúsculas del texto original
- Lista configurable de palabras prohibidas
- Acción: `modify`

**Palabras de ejemplo**: "malo", "feo", "tonto", "idiota", "estúpido"

### 2. StrictBlockingStrategy

**Propósito**: Bloquea completamente mensajes con contenido severo

**Características**:
- Bloquea mensajes que contengan palabras específicas
- Acción: `block`
- Alta confianza (0.9)

**Palabras de ejemplo**: "spam", "hack", "virus", "malware", "fraud"

### 3. WarningStrategy

**Propósito**: Envía advertencias pero permite el mensaje

**Características**:
- Detecta palabras que requieren atención
- Acción: `warn`
- Confianza media (0.6)

**Palabras de ejemplo**: "violencia", "amenaza", "peligro", "riesgo"

### 4. CompositeModerationStrategy

**Propósito**: Combina múltiples estrategias en secuencia

**Características**:
- Ejecuta estrategias en orden de prioridad
- Si una estrategia bloquea, detiene el procesamiento
- Si una estrategia modifica, usa el mensaje modificado para la siguiente
- Acción final: `allow` si pasa todas las verificaciones

## Integración con Observer Pattern

### ModerationObserver

```go
type ModerationObserver struct {
    id           string
    moderator    *ModerationContext
    blockedCount int64
    modifiedCount int64
    warningCount int64
}
```

- Implementa la interfaz `Observer`
- Recibe eventos de tipo `MessageEvent`
- Mantiene estadísticas de moderación
- Permite cambiar estrategias dinámicamente

## API Endpoints

### Cambiar Estrategias

```bash
# Reemplazar malas palabras con ***
POST /moderation/badword

# Bloqueo estricto
POST /moderation/strict

# Solo advertencias
POST /moderation/warning

# Estrategia compuesta
POST /moderation/composite
```

### Estadísticas

```bash
# Obtener estadísticas de moderación
GET /moderation/stats
```

**Respuesta**:
```json
{
    "blocked_messages": 5,
    "modified_messages": 12,
    "warning_messages": 3,
    "strategy": "BadWordReplacement"
}
```

## Interfaz Web

Se ha creado una interfaz web completa (`moderation.html`) que incluye:

- **Botones para cambiar estrategias** en tiempo real
- **Panel de estadísticas** que se actualiza automáticamente
- **Función de prueba** que envía mensajes de ejemplo
- **Indicador de estrategia actual**
- **Chat integrado** para probar la moderación

### Acceso a la Interfaz

```
http://localhost:8080/moderation.html
```

## Flujo de Moderación

```
1. Usuario envía mensaje
2. Servidor recibe mensaje
3. ModerationContext aplica estrategia actual
4. ModerationResult contiene:
   - Mensaje original
   - Mensaje modificado (si aplica)
   - Acción tomada
   - Razón
   - Confianza
5. Si acción es "block": mensaje no se envía
6. Si acción es "modify": se envía mensaje modificado
7. Si acción es "allow" o "warn": se envía mensaje original
8. ModerationObserver actualiza estadísticas
```

## Beneficios del Patrón Strategy

### 1. **Flexibilidad**
- Cambiar estrategias sin reiniciar el servidor
- Agregar nuevas estrategias fácilmente
- Combinar estrategias con Composite

### 2. **Mantenibilidad**
- Cada estrategia está encapsulada
- Fácil testing individual de estrategias
- Código más limpio y organizado

### 3. **Extensibilidad**
Ejemplos de estrategias adicionales que se pueden crear:

- **MLModerationStrategy**: Usar machine learning
- **DatabaseModerationStrategy**: Consultar base de datos
- **APIModerationStrategy**: Usar servicios externos
- **TimeBasedStrategy**: Moderación según horario
- **UserRoleStrategy**: Moderación según rol de usuario

### 4. **Configurabilidad**
- Listas de palabras configurables
- Umbrales de confianza ajustables
- Acciones personalizables

## Ejemplo de Uso

```go
// Crear contexto con estrategia de reemplazo
context := NewModerationContext(NewBadWordReplacementStrategy())

// Moderar mensaje
result := context.ModerateMessage("Eres muy malo")

// Resultado:
// OriginalMessage: "Eres muy malo"
// ModifiedMessage: "Eres muy ***"
// Action: "modify"
// Reason: "Inappropriate words detected and replaced"
// Confidence: 0.8

// Cambiar a estrategia estricta
context.SetStrategy(NewStrictBlockingStrategy())
result = context.ModerateMessage("Esto es spam")
// Action: "block"
```

## Testing

Para probar las diferentes estrategias:

1. **Abrir** `http://localhost:8080/moderation.html`
2. **Cambiar estrategias** usando los botones
3. **Enviar mensajes** con palabras de prueba
4. **Observar** cómo cada estrategia maneja el contenido
5. **Ver estadísticas** en tiempo real

El patrón Strategy proporciona una arquitectura robusta y flexible para la moderación de contenido, permitiendo adaptar el comportamiento del sistema según las necesidades específicas de cada aplicación.
