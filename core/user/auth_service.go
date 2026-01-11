package user

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// SecurityMiddleware valida tokens e aplica Headers de segurança industrial
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers de Segurança Auren
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Bypass para rotas de autenticação
		if strings.HasPrefix(r.URL.Path, "/auth") || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		secretToken := os.Getenv("APP_SECRET_TOKEN")
		token := r.URL.Query().Get("token")

		if secretToken != "" && token != secretToken {
			log.Printf("⚠️ [SECURITY] Acesso negado de %s para a rota %s", r.RemoteAddr, r.URL.Path)
			http.Error(w, "🛡️ Acesso Auren restrito: Token inválido.", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("🔑 [AUTH] Iniciando handshake OAuth2 com Google...")
	state := generateState(16)
	url := GoogleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateState(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// TODO: HandleGoogleCallback (Sincronizar com tabela User do DB)
