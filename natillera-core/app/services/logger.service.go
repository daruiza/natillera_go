package services

import (
	"encoding/json"
	"time"

	"natillera-shared/utils"
)

// LogEntry representa una entrada de log estructurada
type LogEntry struct {
	Event     string    `json:"event"`
	Status    string    `json:"status"`
	TraceID   string    `json:"traceId"`
	Message   string    `json:"message"`
	Data      string    `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// LoggerService proporciona métodos para logging estructurado
type LoggerService struct{}

// NewLoggerService crea una nueva instancia de LoggerService
func NewLoggerService() *LoggerService {
	return &LoggerService{}
}

// LogEvent registra un evento con datos estructurados
func (l *LoggerService) LogEvent(event, status, traceID, message, data string) {
	entry := LogEntry{
		Event:     event,
		Status:    status,
		TraceID:   traceID,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		utils.Error.Printf("Error al serializar entrada de log: %v", err)
		return
	}

	switch status {
	case "ERROR":
		utils.Error.Println(string(jsonEntry))
	case "WARNING":
		utils.Warning.Println(string(jsonEntry))
	default:
		utils.Info.Println(string(jsonEntry))
	}
}

// LogInfo registra un mensaje informativo
func (l *LoggerService) LogInfo(event, traceID, message, data string) {
	l.LogEvent(event, "INFO", traceID, message, data)
}

// LogError registra un mensaje de error
func (l *LoggerService) LogError(event, traceID, message, data string) {
	l.LogEvent(event, "ERROR", traceID, message, data)
}

// LogWarning registra un mensaje de advertencia
func (l *LoggerService) LogWarning(event, traceID, provider, message, data string) {
	l.LogEvent(event, "WARNING", traceID, message, data)
}

// LogSuccess registra un mensaje de éxito
func (l *LoggerService) LogSuccess(event, traceID, message, data string) {
	l.LogEvent(event, "OK", traceID, message, data)
}
