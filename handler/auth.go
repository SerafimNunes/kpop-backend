package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	// O RedirectURL deve vir de uma variável de ambiente no Cloud Run
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"), 
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// SecurityMiddleware: O "Guarda de Elite" que protege suas rotas
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers essenciais de segurança
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Proteção simples por Token para a apresentação (evita intrusos)
		// Para o celular acessar, basta colocar ?token=SUA_CHAVE na URL
		secretToken := os.Getenv("APP_SECRET_TOKEN")
		if secretToken != "" {
			token := r.URL.Query().Get("token")
			if token != secretToken && r.URL.Path != "/auth/google/login" {
				http.Error(w, "Acesso restrito: Token inválido ou ausente.", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// HandleGoogleLogin inicia o fluxo OAuth
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Em um sistema real, você salvaria esse 'state' em um cookie para validar no callback
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}