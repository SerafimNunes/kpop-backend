package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"auren-platform/core/user"
	"auren-platform/db"
	"auren-platform/internal/engine" // Novo motor central
	"auren-platform/media_analysis"
	"auren-platform/realtime"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// --- CONFIGURAÇÃO DE LOGGING PERSISTENTE (MANIFESTO 1.4) ---
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("❌ Falha crítica ao criar diretório de logs: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "auren_system.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("❌ Falha crítica ao abrir arquivo de log: %v\n", err)
		os.Exit(1)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("### 🛠️  Iniciando sequência de boot: AUREN PLATFORM V1.4 ###")

	// 0. Carregamento de Ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aviso: Arquivo .env não localizado, utilizando variáveis de ambiente.")
	}

	// 1. Inicialização de Motores e Infraestrutura
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Garantir diretórios de storage (Manifesto 1.4)
	dirs := []string{"./storage/evidences", "./storage/knowledge"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("❌ Erro ao criar diretório de armazenamento: %v", err)
		}
	}

	log.Println("🐘 Sincronizando Banco de Dados PostgreSQL...")
	db.InitDB()

	// Inicialização do Engine Central (O Cérebro)
	log.Println("🧠 Inicializando Auren Semantic Engine & Agents...")
	aurenEngine, err := engine.NewAurenEngine(ctx)
	if err != nil {
		log.Fatalf("❌ CRÍTICO: Falha ao iniciar Engine: %v", err)
	}

	log.Println("🎙️  Configurando Processadores de Mídia...")
	streamAudioProc := media_analysis.NewStreamAudioProcessor(40.0)
	videoProc := media_analysis.NewVideoProcessor()
	fileAudioProc := media_analysis.NewFileAudioProcessor()

	log.Println("🌐 Iniciando Hub de Comunicação Realtime...")
	aurenHub := realtime.NewHub()
	go aurenHub.Run()

	// 2. Roteamento (Dispatcher Pattern)
	r := mux.NewRouter()

	// --- API PROTEGIDA (O CORAÇÃO DO MANIFESTO) ---
	api := r.PathPrefix("/api").Subrouter()
	api.Use(user.SecurityMiddleware)

	// ENDPOINT ÚNICO: O SEMANTIC DISPATCHER
	// Aqui o Frontend envia {module, action, payload}
	api.HandleFunc("/execute", aurenEngine.HandleExecution).Methods("POST")

	// Manter endpoint de saúde
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "Auren Online",
			"engine": "AUREN-SEMANTIC-V1.4",
			"time":   time.Now().Format(time.RFC3339),
		})
	}).Methods("GET")

	// --- WEBSOCKETS ---
	wsSub := r.PathPrefix("/ws").Subrouter()
	wsSub.Use(user.SecurityMiddleware)
	wsSub.HandleFunc("/engine", func(w http.ResponseWriter, r *http.Request) {
		realtime.ServeWS(aurenHub, aurenEngine.Gemini, streamAudioProc, videoProc, fileAudioProc, aurenEngine.Librarian, w, r)
	})

	// --- SERVIDOR DE ARQUIVOS ESTÁTICOS (UX CLEAN INDUSTRIAL) ---
	staticDir := "./static"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(staticDir, "auren.html"))
	})

	fileServer := http.FileServer(http.Dir(staticDir))
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		fileServer.ServeHTTP(w, r)
	}))

	// 3. Inicialização do Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// Banner Profissional
	localIP := getLocalIP()
	log.Println("==================================================")
	log.Println("🛡️  AUREN PLATFORM : SISTEMA DE CONFORMIDADE AMBIENTAL")
	log.Printf("🚀 ERP LOCAL: http://localhost:%s", port)
	log.Printf("🌍 REDE: http://%s:%s", localIP, port)
	log.Println("==================================================")

	// 4. Graceful Shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("\n🛑 Iniciando encerramento seguro...")
		aurenEngine.Close()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("❌ Falha no shutdown: %v", err)
		}
		logFile.Close()
		log.Println("👋 Auren Platform encerrada.")
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("❌ ERRO CRÍTICO: %v", err)
	}
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
