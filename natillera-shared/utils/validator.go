package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
)

var Validate *validator.Validate

var CustomValidationMessages = map[string]string{
	"required":        "El campo %s es obligatorio.",
	"email":           "El campo %s debe ser una dirección de correo electrónico válida.",
	"gte":             "El campo %s debe ser mayor o igual a %s.",
	"lte":             "El campo %s debe ser menor o igual a %s.",
	"min":             "El campo %s debe tener al menos %s caracteres.",
	"max":             "El campo %s debe tener como máximo %s caracteres.",
	"containsany":     "El campo %s debe contener al menos uno de los siguientes caracteres especiales: %s.",
	"len":             "El campo %s debe tener exactamente %s caracteres.",
	"e164":            "El campo %s debe tener un formato de número de teléfono E.164 válido (ej. +573001234567).",
	"url":             "El campo %s debe ser una URL válida.",
	"oneof":           "El campo %s debe ser uno de los siguientes valores: %s.",
	"eqfield":         "El campo %s debe ser igual al campo %s.",
	"lte=now":         "El campo %s no puede ser una fecha en el futuro.",
	"numeric":         "El campo %s debe ser un número.",
	"containsletter":  "El campo %s debe contener al menos una letra.",
	"containsspecial": "El campo %s debe contener al menos un carácter especial.",
	"onlynumbers":     "El campo %s solo debe contener números.",
	"onlyletters":     "El campo %s solo debe contener letras.",
}

func NewValidator() {
	Validate = validator.New()

	// Registro de validaciones personalizadas
	Validate.RegisterValidation("containsletter", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`[a-zA-Z]`).MatchString(fl.Field().String())
	})

	Validate.RegisterValidation("containsspecial", func(fl validator.FieldLevel) bool {
		// Define aquí los símbolos que consideras especiales
		return regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(fl.Field().String())
	})

	Validate.RegisterValidation("onlynumbers", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[0-9]+$`).MatchString(fl.Field().String())
	})

	Validate.RegisterValidation("onlyletters", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(fl.Field().String())
	})
}

// getCustomErrorMessage busca un mensaje de error personalizado para la etiqueta de validación dada.
func GetCustomErrorMessage(err validator.FieldError) string {
	if msg, ok := CustomValidationMessages[err.Tag()]; ok {
		// Reemplazar placeholders si existen
		switch err.Tag() {
		case "required":
			return fmt.Sprintf(msg, err.Field())
		case "email":
			return fmt.Sprintf(msg, err.Field())
		case "gte":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "lte":
			if err.Param() == "now" {
				return fmt.Sprintf(CustomValidationMessages["lte=now"], err.Field())
			}
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "min":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "max":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "containsany":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "len":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "e164":
			return fmt.Sprintf(msg, err.Field())
		case "url":
			return fmt.Sprintf(msg, err.Field())
		case "oneof":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "eqfield":
			return fmt.Sprintf(msg, err.Field(), err.Param())
		case "numeric":
			return fmt.Sprintf(msg, err.Field())
		default:
			return "default error"
		}
	}
	return "default error"
}

func ValidateStruct(s interface{}, errorMessage *map[int]string) error {
	if err := Validate.Struct(s); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			(*errorMessage)[0] = err.Error()
		}

		// Iterar sobre los errores de validación y mostrar mensajes personalizados
		for i, err := range err.(validator.ValidationErrors) {
			(*errorMessage)[i+1] = GetCustomErrorMessage(err)
		}
	}

	if len(*errorMessage) > 0 {
		var errorMessagesSlice []string
		//recorremos errorMessage para armar el mensaje de error
		for _, msg := range *errorMessage {
			errorMessagesSlice = append(errorMessagesSlice, msg)
		}
		// Unimos todos los mensajes de error con ", " como separador
		fullErrorMessage := strings.Join(errorMessagesSlice, ", ")

		// retormamos el mensaje de error completo
		return fmt.Errorf("error de validación: %s", fullErrorMessage)
	}

	return nil
}
