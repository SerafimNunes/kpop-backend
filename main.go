package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k-lens/db"
	"k-lens/handler"
	"k-lens/hub"
	"k-lens/translate"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não carregado, usando variáveis de ambiente do sistema")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Inicialização do Gemini 2.0 Flash
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatal("❌ ERRO CRÍTICO: GEMINI_API_KEY não definida. Defina a variável de ambiente antes de iniciar.")
	}

	geminiSvc, err := translate.NewGeminiService(ctx, geminiKey)
	if err != nil {
		log.Fatalf("❌ Erro crítico ao iniciar Gemini: %v", err)
	}
	defer geminiSvc.Close()

	// 2. Banco de Dados e Hub de WebSockets
	db.InitDB()
	legendasHub := hub.NewHub()
	go legendasHub.Run()

	r := mux.NewRouter()

	// --- CAMADA DE SEGURANÇA (MIDDLEWARE) ---
	// Isso protege contra ataques e exige o token ?token=... definido no seu .env
	r.Use(handler.SecurityMiddleware)

	// --- ROTAS DE AUTENTICAÇÃO ---
	r.HandleFunc("/auth/google/login", handler.HandleGoogleLogin)

	// --- ROTA PARA DOWNLOAD DOS CORTES ---
	// Protegido pelo middleware: o celular precisa do token para baixar o vídeo
	r.PathPrefix("/recordings/").Handler(http.StripPrefix("/recordings/", http.FileServer(http.Dir("./recordings"))))

	// --- HEALTH CHECK (Google Cloud Load Balancer) ---
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// --- ROTA WEBSOCKET (O coração do Studio) ---
	r.HandleFunc("/ws/studio/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWS(legendasHub, geminiSvc, w, r)
	})

	// --- API TRADUÇÃO REVERSA ---
	r.HandleFunc("/api/translate-reverse", handler.ReverseTranslate).Methods("POST", "OPTIONS")

	// --- ARQUIVOS ESTÁTICOS (Frontend) ---
	// Deve ficar por último para não interceptar as rotas acima
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// 3. Configuração do Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: r,
	}

	// Graceful Shutdown: Fecha tudo certinho ao desligar
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("Encerrando servidor...")
		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelTimeout()
		server.Shutdown(ctxTimeout)
	}()

	localIP := getLocalIP()
	log.Printf("==========================================")
	log.Printf("🚀 K-LENS ARMY STUDIO (Modo Seguro Ativo)")
	log.Printf("🔗 Link Studio: http://localhost:%s/studio.html?token=%s", port, os.Getenv("APP_SECRET_TOKEN"))
	log.Printf("📱 Mobile: http://%s:%s/studio.html?token=%s", localIP, port, os.Getenv("APP_SECRET_TOKEN"))
	log.Printf("==========================================")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
