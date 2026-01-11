package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"auren-platform/ai"
	"auren-platform/core/user"
	"auren-platform/db"
	"auren-platform/media_analysis"
	"auren-platform/realtime"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("### 🛠️  Iniciando sequência de boot da Auren Platform ###")

	// 0. Carregamento de Ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aviso: Arquivo .env não localizado, utilizando variáveis de ambiente do sistema.")
	} else {
		log.Println("✅ Variáveis de ambiente carregadas via .env")
	}

	// Configurações Google Cloud
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	credsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if location == "" {
		location = "us-central1"
		log.Printf("ℹ️  Localização Google Cloud não definida, assumindo: %s", location)
	}

	if credsFile != "" {
		if _, err := os.Stat(credsFile); err == nil {
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credsFile)
			log.Printf("✅ Credenciais Google configuradas via arquivo: %s", credsFile)
		} else {
			log.Printf("⚠️  Aviso: Arquivo de credenciais definido mas não encontrado: %s", credsFile)
		}
	}

	// 1. Inicialização de Motores e Infraestrutura
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Garantir existência de diretórios de storage para Evidências e Conhecimento
	dirs := []string{"./storage/evidences", "./storage/knowledge"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("❌ Erro ao criar diretório de armazenamento: %v", err)
		}
	}

	log.Println("🐘 Conectando ao Banco de Dados e Sincronizando Esquemas...")
	db.InitDB()

	log.Println("🧠 Inicializando Vertex AI (Gemini Service Multimodal)...")
	geminiSvc, err := ai.NewGeminiService(ctx, projectID, location)
	if err != nil {
		log.Fatalf("❌ CRÍTICO: Falha ao iniciar Gemini Service: %v", err)
	}
	log.Println("✅ Gemini Service operacional.")

	log.Println("🎙️  Configurando Processador de Mídia e Escuta Digital...")
	audioProc := media_analysis.NewAudioProcessor(40.0)

	log.Println("🌐 Iniciando Hub de Comunicação Realtime...")
	aurenHub := realtime.NewHub()
	go aurenHub.Run()

	// 2. Roteamento e Middlewares
	log.Println("🛣️  Configurando malha de roteamento HTTP/WS...")
	r := mux.NewRouter()

	// --- API & WS (Protegidos por Token/Security) ---
	api := r.PathPrefix("/api").Subrouter()
	api.Use(user.SecurityMiddleware)

	// ENDPOINT DE TESTE DE SANIDADE DA IA
	api.HandleFunc("/ai-test", func(w http.ResponseWriter, r *http.Request) {
		log.Println("🧪 [DIAGNÓSTICO] Testando comunicação direta com Vertex AI...")
		testCtx, testCancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer testCancel()

		res, err := geminiSvc.ExecuteSimplePrompt(testCtx, "Responda apenas: Auren Engine Online")
		if err != nil {
			log.Printf("❌ [DIAGNÓSTICO] Erro: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "status": "fail"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"response": res, "status": "success"})
	}).Methods("GET")

	wsSub := r.PathPrefix("/ws").Subrouter()
	wsSub.Use(user.SecurityMiddleware)

	wsSub.HandleFunc("/engine", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔌 [WS] Conexão estabelecida: %s", r.RemoteAddr)
		realtime.ServeWS(aurenHub, geminiSvc, audioProc, w, r)
	})

	// Endpoints de Autenticação e Saúde do Sistema
	r.HandleFunc("/auth/google/login", user.HandleGoogleLogin).Methods("GET")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "Auren Online",
			"engine": "AUREN-CORE-V1.0",
			"db":     "PostgreSQL Connected",
			"time":   time.Now().Format(time.RFC3339),
		})
	}).Methods("GET")

	// --- SERVIDOR DE ARQUIVOS ESTÁTICOS (CORRIGIDO) ---
	staticDir := "./static"

	// Rota para a interface principal (Entrypoint)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📦 Servindo Core UI: %s/auren.html", staticDir)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(staticDir, "auren.html"))
	})

	// FileServer para recursos (JS, CSS, Seções HTML)
	fileServer := http.FileServer(http.Dir(staticDir))
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proteção contra cache e definição de tipos MIME
		path := r.URL.Path
		if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		}

		fileServer.ServeHTTP(w, r)
	}))

	// 3. Inicialização do Servidor HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      r,
		ReadTimeout:  60 * time.Second, // Timeout expandido para uploads de mídia
		WriteTimeout: 60 * time.Second,
	}

	// Banner de Inicialização Profissional
	localIP := getLocalIP()
	tokenStr := os.Getenv("APP_SECRET_TOKEN")

	log.Println("==================================================")
	log.Println("🛡️  AUREN PLATFORM : SISTEMA DE CONFORMIDADE AMBIENTAL")
	log.Printf("🚀 ERP LOCAL: http://localhost:%s?token=%s", port, tokenStr)
	log.Printf("🌍 REDE (IPv4): http://%s:%s?token=%s", localIP, port, tokenStr)
	log.Println("==================================================")

	// 4. Mecanismo de Graceful Shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("\n🛑 Sinal de interrupção recebido. Iniciando encerramento seguro...")

		// Fecha conexão com Vertex AI
		if geminiSvc != nil {
			geminiSvc.Close()
		}

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("❌ Falha ao encerrar servidor HTTP: %v", err)
		}
		log.Println("👋 Auren Platform encerrada. Todos os processos finalizados.")
	}()

	// Início do Loop de Escuta
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("❌ ERRO CRÍTICO DE EXECUÇÃO: %v", err)
	}
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}