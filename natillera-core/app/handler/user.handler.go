package handler

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"natillera-core/app/domain"
	"natillera-core/app/repository"
	"natillera-core/app/services"
	"natillera-core/config"
	natsManager "natillera-shared/nats"
	"natillera-shared/utils"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/idtoken"
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

func (uh *UserHandler) ValidateGoogleToken(w http.ResponseWriter, r *http.Request) {
	const event = "UserHandler.ValidateGoogleToken"
	traceId := uuid.NewString()

	// 1. Leer body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		uh.loggerService.LogError(event+".ReadAll", traceId, err.Error(), "Failed to read body")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, err.Error(), event)
		return
	}
	defer r.Body.Close()

	// 2. Deserializar JSON
	var data domain.GoogleTokenRequest
	if err := json.Unmarshal(body, &data); err != nil {
		uh.loggerService.LogError(event+".Unmarshal", traceId, err.Error(), "Failed to Unmarshal body")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, err.Error(), event)
		return
	}

	// 3. Validar struct
	errorMessage := make(map[int]string)
	if err = utils.Validate.Struct(data); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			errorMessage[0] = err.Error()
		}
		for i, vErr := range err.(validator.ValidationErrors) {
			errorMessage[i+1] = utils.GetCustomErrorMessage(vErr)
		}
	}

	if len(errorMessage) > 0 {
		var errorMessagesSlice []string
		for _, msg := range errorMessage {
			errorMessagesSlice = append(errorMessagesSlice, msg)
		}
		fullErrorMessage := strings.Join(errorMessagesSlice, ", ")
		uh.loggerService.LogError(event+".ValidationError", traceId, fullErrorMessage, "")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, fullErrorMessage, event)
		return
	}

	// 4. Limpiar token y clientID
	idToken := strings.TrimSpace(data.Token)
	clientID := strings.TrimSpace(uh.cnf.Get("OAUTH_CLIENT_ID"))

	// DEBUG TEMPORAL - ver el token exacto que llega
	utils.Info.Printf("Token raw bytes (primeros 50 chars): %q", idToken[:50])
	utils.Info.Printf("Token raw bytes (ultimos 20 chars): %q", idToken[len(idToken)-20:])

	if clientID == "" {
		uh.loggerService.LogError(event+".MissingClientID", traceId, "missing OAUTH_CLIENT_ID", "")
		uh.errorResponse.SendErrorResponse(w, http.StatusInternalServerError, "Configuración inválida del servidor", event)
		return
	}

	// 5. Verificar formato JWT (3 partes separadas por ".")
	if len(strings.Split(idToken, ".")) != 3 {
		uh.loggerService.LogError(event+".InvalidTokenFormat", traceId, "token is not a valid JWT", "")
		uh.errorResponse.SendErrorResponse(w, http.StatusBadRequest, "Token inválido: se requiere el id_token, no el access_token", event)
		return
	}

	utils.Info.Printf("Validating token (Token length: %d)", len(idToken))

	googleCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	claims, err := uh.validateGoogleToken(googleCtx, idToken)
	if err != nil {
		// Log decoded claims for debugging
		unverifiedClaims := uh.decodeJWTClaims(idToken)
		uh.loggerService.LogError(event+".InvalidToken", traceId, err.Error(),
			fmt.Sprintf("Decoded claims (unverified): iss=%v, aud=%v, sub=%v",
				unverifiedClaims["iss"], unverifiedClaims["aud"], unverifiedClaims["sub"]))

		uh.errorResponse.SendErrorResponse(w, http.StatusUnauthorized, "Token inválido o expirado", event)
		return
	}

	// Verificar audience (debe coincidir con alguno de nuestros IDs)
	tokenAud, _ := claims["aud"].(string)
	webID := uh.cnf.Get("OAUTH_CLIENT_ID")
	androidID := uh.cnf.Get("ANROID_OAUTH_CLIENT_ID")

	if tokenAud != webID && tokenAud != androidID {
		uh.loggerService.LogError(event+".AudienceMismatch", traceId,
			fmt.Sprintf("expected aud=%s or %s, got %s", webID, androidID, tokenAud), "")
		uh.errorResponse.SendErrorResponse(w, http.StatusUnauthorized, "Token no válido para esta aplicación", event)
		return
	}

	user := map[string]interface{}{
		"id":             claims["sub"],
		"email":          claims["email"],
		"verified_email": claims["email_verified"],
		"name":           claims["name"],
		"given_name":     claims["given_name"],
		"family_name":    claims["family_name"],
		"picture":        claims["picture"],
	}

	utils.Info.Printf("Token válido para usuario: %s", claims["email"])

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Token válido",
		"user":    user,
	}); err != nil {
		uh.loggerService.LogError(event+".EncodeResponse", traceId, err.Error(), "Failed to encode JSON")
		uh.errorResponse.SendErrorResponse(w, http.StatusInternalServerError, "Failed to encode response", event)
		return
	}
}

// validateGoogleToken intenta validar el id_token contra múltiples Client IDs.
func (uh *UserHandler) validateGoogleToken(ctx context.Context, idToken string) (map[string]interface{}, error) {

	clientIDs := []string{
		uh.cnf.Get("OAUTH_CLIENT_ID"),
		uh.cnf.Get("ANROID_OAUTH_CLIENT_ID"),
	}

	for _, rawID := range clientIDs {
		// Limpiar ID por si acaso tiene comillas o espacios
		clientID := strings.Trim(strings.TrimSpace(rawID), "\"")
		if clientID == "" {
			continue
		}

		utils.Info.Printf("Verifying with ClientID: [%s] (len: %d) - Hex: %s",
			clientID, len(clientID), hex.EncodeToString([]byte(clientID)))

		// Intento 1: validación local con idtoken (sin red, más rápido)
		payload, err := idtoken.Validate(ctx, idToken, clientID)
		if err == nil {
			return map[string]interface{}{
				"sub":            payload.Subject,
				"email":          payload.Claims["email"],
				"email_verified": payload.Claims["email_verified"],
				"name":           payload.Claims["name"],
				"given_name":     payload.Claims["given_name"],
				"family_name":    payload.Claims["family_name"],
				"picture":        payload.Claims["picture"],
				"aud":            payload.Audience,
			}, nil
		}
		utils.Info.Printf("idtoken.Validate falló para %s: %v", clientID, err)
	}

	utils.Info.Printf("idtoken.Validate falló para todos los IDs, intentando con tokeninfo API de Google...")

	// Intento 2: fallback a Google API (requiere red)
	params := url.Values{}
	params.Set("id_token", idToken)
	// Usamos la URL v3 que es más moderna y robusta
	fullURL := "https://www.googleapis.com/oauth2/v3/tokeninfo?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error contactando Google API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta de Google: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google tokeninfo rechazó el token: %s", string(body))
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, fmt.Errorf("error parseando claims: %w", err)
	}

	return claims, nil
}

func (uh *UserHandler) decodeJWTClaims(token string) map[string]interface{} {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}

	return claims
}
