# Implementación del Patrón Observer/Publish-Subscribe

## Descripción General

Se ha refactorizado la aplicación de chat para implementar el patrón Observer (también conocido como Publisher-Subscriber). Este patrón permite que múltiples observadores reciban notificaciones sobre eventos que ocurren en el sistema de manera desacoplada y escalable.

## Estructura Implementada

### 1. Interfaz Observer

```go
type Observer interface {
    Update(event Event)
    GetID() string
}
```

Todos los observadores deben implementar esta interfaz para recibir actualizaciones de eventos.

### 2. Interfaz Publisher

```go
type Publisher interface {
    Subscribe(observer Observer)
    Unsubscribe(observer Observer)
    Notify(event Event)
}
```

El Publisher maneja la suscripción y notificación de eventos a sus observadores.

### 3. Tipos de Eventos

Se definieron cuatro tipos principales de eventos:

- `MessageEvent`: Cuando se envía un mensaje
- `UserJoinEvent`: Cuando un usuario se conecta
- `UserLeave`: Cuando un usuario se desconecta  
- `SystemEvent`: Para mensajes del sistema

### 4. Implementaciones de Observer

#### ConnectionObserver

Maneja conexiones WebSocket individuales:

- **Responsabilidad**: Enviar eventos a un cliente específico
- **Características**: 
  - Canal interno para mensajes con buffer de 100 eventos
  - Manejo concurrente con goroutines
  - Limpieza automática al desconectarse

#### LoggerObserver

Proporciona logging de todos los eventos:

- **Responsabilidad**: Registrar eventos en consola
- **Características**: 
  - Formateo de tiempo
  - Información de usuario incluida
  - Sin estado interno

#### StatsObserver

Recopila estadísticas del chat:

- **Responsabilidad**: Mantener métricas del sistema
- **Características**:
  - Contador de mensajes totales
  - Usuarios únicos conectados
  - Mensajes por hora
  - Timer para mostrar estadísticas cada 30 segundos

## Beneficios del Patrón Implementado

### 1. **Desacoplamiento**
- Los observadores no conocen los internals del Publisher
- Fácil agregar/quitar funcionalidades sin modificar código existente

### 2. **Escalabilidad**
- Nuevos tipos de observadores se pueden agregar sin cambios al servidor principal
- Cada observador maneja sus propios eventos de manera independiente

### 3. **Concurrencia**
- Cada observador corre en su propia goroutine
- El Publisher usa canales para decoplar el proceso de distribución

### 4. **Extensibilidad**
Ejemplos de observadores adicionales que se pueden crear:

- **DatabaseObserver**: Guardar mensajes en base de datos
- **NotificationObserver**: Enviar notificaciones push
- **AnalyticsObserver**: Enviar datos a sistemas de analytics
- **ModerationObserver**: Detectar contenido inapropiado

## Diagrama de Arquitectura

```
Server (Publisher)
    ↓ Notify()
EventPublisher
    ↓ Update()
├── ConnectionObserver → WebSocket Clients
├── LoggerObserver → Console Logs
└── StatsObserver → Statistics
```

## Uso del Frontend

El frontend ahora maneja eventos JSON estructurados:

```javascript
// Recibe eventos estructurados
{
    "type": "message",
    "message": "Hola mundo",
    "username": "usuario1",
    "timestamp": "2023-10-03T14:30:00Z",
    "data": {...}
}
```

### Tipos de Eventos Frontend:

- `message`: Mensaje de chat normal
- `user_join`: Notificación de usuario conectado
- `user_leave`: Notificación de usuario desconectado  
- `system`: Mensaje del sistema

## Ventajas de la Nueva Implementación

1. **Mejor Organización**: Cada responsabilidad está separada
2. **Más Flexible**: Agregar nuevas funcionalidades es trivial
3. **Mejor Testing**: Cada observer se puede testear independientemente
4. **Mejor Mantenimiento**: Código más modular y fácil de entender
5. **Mejor Performance**: Procesamiento concurrente y distribución eficiente

## Cómo Ejecutar

```bash
# Compilar
go build .

# Ejecutar
go run .

# Acceder en navegador
http://localhost:8080
```

El servidor mostrará en la consola:
- Logs de eventos en tiempo real
- Estadísticas cada 30 segundos
- Información de conexiones/desconexiones
