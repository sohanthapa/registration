package http

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/example/registration/internal/app"
	"github.com/example/registration/internal/domain"
)

const maxBodyBytes = 1 << 20 // 1 MB

type Handler struct {
	service *app.Service
}

func Register(mux *http.ServeMux, service *app.Service) {
	handler := &Handler{
		service: service,
	}

	mux.HandleFunc("/signup", handler.signUp)
	mux.HandleFunc("/login", handler.login)
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type loginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        userResponse `json:"user"`
}

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req authRequest
	if !readJSON(w, r, &req) {
		return
	}

	user, err := h.service.SignUp(r.Context(), app.Credentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user": toUserResponse(user),
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req authRequest
	if !readJSON(w, r, &req) {
		return
	}

	result, err := h.service.SignIn(r.Context(), app.Credentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: result.Token.AccessToken,
		TokenType:   result.Token.TokenType,
		ExpiresIn:   result.Token.ExpiresIn,
		ExpiresAt:   result.Token.ExpiresAt,
		User:        toUserResponse(result.User),
	})
}

func handleAuthError(w http.ResponseWriter, err error) {
	var validationErr *app.ValidationError

	switch {
	case errors.As(err, &validationErr):
		errorJSON(w, http.StatusBadRequest, validationErr.Message)

	case errors.Is(err, domain.ErrUserAlreadyExists):
		errorJSON(w, http.StatusConflict, "email is already registered")

	case errors.Is(err, domain.ErrInvalidCredentials):
		errorJSON(w, http.StatusUnauthorized, "invalid email or password")

	default:
		log.Printf("app error: %v", err)
		errorJSON(w, http.StatusInternalServerError, "internal server error")
	}
}

func toUserResponse(user domain.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		errorJSON(w, http.StatusBadRequest, "body must contain only one JSON object")
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func methodNotAllowed(w http.ResponseWriter, allowedMethods ...string) {
	w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	errorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
}
