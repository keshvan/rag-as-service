package handlers

import (
	"net/http"

	authClient "github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/auth"
)

type AuthHandler struct {
	authClient *authClient.Client
}

func NewAuthHandler(client *authClient.Client) *AuthHandler {
	return &AuthHandler{
		authClient: client,
	}
}

//TODO

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) Verification(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) GetCurrentUserGUID(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) GetTokenPairByUserGUID(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {

}
