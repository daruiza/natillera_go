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

	"github.com/google/uuid"

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

	uh.loggerService.LogInfo(event, traceId, "", "")
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
