package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ErrorResponse define la estructura para las respuestas de error
type ErrorResponse struct {
	StatusCode  int    `json:"status_code"`
	Error       bool   `json:"error"`
	Message     string `json:"message"`
	RequestBody string `json:"requestBody"`
}

type ClientErrorResponse struct {
	TransactionId     string `json:"transactionId,omitempty"`
	ReferenceExternal string `json:"referenceExternal,omitempty"`
	RequestId         string `json:"requestId,omitempty"`
	Message           string `json:"message"`
}

// StandardResponse define el formato estándar para todas las respuestas.
type StandardResponse struct {
	Status string      `json:"status"` // "error" o "success"
	Code   int         `json:"code"`   // Código HTTP (por ejemplo, 200, 404, etc.)
	Data   interface{} `json:"data"`   // Datos en caso de éxito o información del error
}

// ErrorData contiene la información detallada de un error.
type ErrorData struct {
	Code    string `json:"code"`    // Por ejemplo, "RECURSO_NO_ENCONTRADO"
	Message string `json:"message"` // Mensaje de error
	Time    string `json:"time"`    // Timestamp en formato RFC3339
	Route   string `json:"route"`   // Ruta del endpoint
}

func NewErrorResponse() *ErrorResponse {
	return &ErrorResponse{}
}

// sendErrorResponse envía la respuesta de error en formato JSON
func (er *ErrorResponse) SendErrorResponse(w http.ResponseWriter, statusCode int, message string, request string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	var err bool
	if statusCode >= 200 && statusCode < 300 {
		err = false
	} else {
		err = true
	}
	response := ErrorResponse{
		StatusCode:  statusCode,
		Error:       err,
		Message:     message,
		RequestBody: fmt.Sprintf("%v", request),
	}
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse envía la respuesta de error en formato JSON para el cliente
func (er *ErrorResponse) ClientErrorResponse(w http.ResponseWriter, statusCode int, message string, requestId string, transactionId string, referenceExternal string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ClientErrorResponse{
		TransactionId:     transactionId,
		ReferenceExternal: referenceExternal,
		RequestId:         requestId,
		Message:           message,
	}
	json.NewEncoder(w).Encode(response)
}

// Función auxiliar para responder errores en el formato estándar sin necesidad de pasar un error.
func (er *ErrorResponse) ErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, codigo, mensaje string) {
	w.Header().Set("Content-Type", "application/json")
	errorData := ErrorData{
		Code:    codigo,
		Message: mensaje,
		Time:    time.Now().UTC().Format(time.RFC3339),
		Route:   r.URL.Path,
	}
	standardResp := StandardResponse{
		Status: "error",
		Code:   statusCode,
		Data:   errorData,
	}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(standardResp)
}
