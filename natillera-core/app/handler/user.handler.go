package handler

import (
	"context"
	"encoding/json"
	"io"
	"natillera-core/app/domain"
	"natillera-core/app/repository"
	"natillera-core/app/services"
	"natillera-core/config"
	natsManager "natillera-shared/nats"
	"natillera-shared/utils"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserHandler struct {
	repo              *repository.UserRepository
	natsService       *natsManager.NatsStarter
	cnf               *config.MicroConfig
	loggerService     *services.LoggerService
	errorResponse     *services.ErrorResponse
	googleOauthConfig *oauth2.Config
}

func NewUserHandler(repo *repository.UserRepository, natsService *natsManager.NatsStarter, cnf *config.MicroConfig) *UserHandler {

	// Inicializar servicios
	loggerService := services.NewLoggerService()
	errorResponse := services.NewErrorResponse()

	var googleOauthConfig = &oauth2.Config{
		RedirectURL:  cnf.Get("OAUTH_REDIRECT_URL"),
		ClientID:     cnf.Get("OAUTH_CLIENT_ID"),
		ClientSecret: cnf.Get("OAUTH_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return &UserHandler{
		repo:              repo,
		loggerService:     loggerService,
		errorResponse:     errorResponse,
		natsService:       natsService,
		cnf:               cnf,
		googleOauthConfig: googleOauthConfig,
	}
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	const event = "UserHandler.Login"
	traceId := uuid.NewString()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		uh.loggerService.LogError(event+".ReadAll", traceId, err.Error(), "Failed to read body")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, err.Error(), event)
		return
	}
	defer r.Body.Close()

	var data domain.LoginRequest
	if err := json.Unmarshal(body, &data); err != nil {
		uh.loggerService.LogError(event+".Unmarshal", traceId, err.Error(), "Failed to Unmarshal body")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, err.Error(), event)
		return
	}

	errorMessage := make(map[int]string)
	err = utils.Validate.Struct(data)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			errorMessage[0] = err.Error()
		}

		// Iterar sobre los errores de validación y mostrar mensajes personalizados
		for i, err := range err.(validator.ValidationErrors) {
			errorMessage[i+1] = utils.GetCustomErrorMessage(err)
		}
	}

	if len(errorMessage) > 0 {
		var errorMessagesSlice []string
		//recorremos errorMessage para armar el mensaje de error
		for _, msg := range errorMessage {
			utils.Error.Println("Error: ", msg) // Esto sigue imprimiendo cada error individualmente
			errorMessagesSlice = append(errorMessagesSlice, msg)
		}
		// Unimos todos los mensajes de error con ", " como separador
		fullErrorMessage := strings.Join(errorMessagesSlice, ", ")
		uh.loggerService.LogError(event+".ValidationError", traceId, fullErrorMessage, "")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, fullErrorMessage, event)
		return
	}

	filter := bson.M{"email": data.Email}

	users, err := uh.repo.GetUsers(&filter)
	if err != nil {
		uh.loggerService.LogError(event+".GetUsers", traceId, err.Error(), "")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, err.Error(), event)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte((*users)[0].Password), []byte(data.Password)); err != nil {
		uh.loggerService.LogError(event+".InvalidCredentials", traceId, err.Error(), "Invalid Credentials")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, err.Error(), event+"Invalid Credentials")
		return
	}

	// Generar token
	token, err := utils.GenerateToken((*users)[0].ID, (*users)[0].Name, (*users)[0].Email, (*users)[0].RolId)
	if err != nil {
		http.Error(w, "Error al generar token", http.StatusInternalServerError)
		return
	}

	utils.Info.Printf("Enviando mensaje a NATS: %s", "natillera.login.user")
	uh.natsService.EventSender.SendMsgPB("natillera.login.user", (*users)[0].ToProtocolBuffer())

	// Respuesta
	response := map[string]interface{}{
		"message": "Login exitoso",
		"token":   token,
		"user":    (*users)[0],
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		uh.loggerService.LogError(event+".EncodeResponse", traceId, err.Error(), "Failed to encode JSON")
		uh.errorResponse.SendErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", event)
		return
	}
}

func (uh *UserHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {

	url := uh.googleOauthConfig.AuthCodeURL("random_state_string")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (uh *UserHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {

	//obtención de los parametros del request en el type GoogleQuery
	var googleQuery domain.GoogleQuery
	googleQuery.Code = r.FormValue("code")
	googleQuery.State = r.FormValue("state")
	// googleQuery.Iss = r.FormValue("iss")
	// googleQuery.Scope = r.FormValue("scope")
	// googleQuery.AuthUser = r.FormValue("authuser")
	// googleQuery.Prompt = r.FormValue("prompt")

	if googleQuery.State != "random_state_string" {
		http.Error(w, "State no válido", http.StatusBadRequest)
		return
	}

	token, err := uh.googleOauthConfig.Exchange(context.Background(), googleQuery.Code)
	if err != nil {
		http.Error(w, "Fallo el intercambio de token", http.StatusInternalServerError)
		return
	}

	googleUserInfo, err := uh.GoogleGetUserInfo(token)
	if err != nil {
		http.Error(w, "Fallo el intercambio de token", http.StatusInternalServerError)
		return
	}

	utils.Info.Println("googleUserInfo:: ", googleUserInfo)

}

func (uh *UserHandler) GoogleGetUserInfo(token *oauth2.Token) (*domain.GoogleUserInfo, error) {

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		uh.loggerService.LogError("UserHandler.GoogleGetUserInfo", "", err.Error(), "Failed to get user info")
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uh.loggerService.LogError("UserHandler.GoogleGetUserInfo", "", err.Error(), "Failed to read body")
		return nil, err
	}

	var googleUserInfo domain.GoogleUserInfo
	if err := json.Unmarshal(body, &googleUserInfo); err != nil {
		uh.loggerService.LogError("UserHandler.GoogleGetUserInfo", "", err.Error(), "Failed to Unmarshal body")
		return nil, err
	}
	return &googleUserInfo, nil

}
