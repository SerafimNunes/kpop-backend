package main

import (
	"context"
	"encoding/json"
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

	"auren-platform/core/events" // Importação vital para o Hub
	"auren-platform/db"
	adminsvc "auren-platform/internal/domain/admin/service"
	"auren-platform/internal/engine"
	"auren-platform/media_analysis"
	"auren-platform/realtime"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// --- CONFIGURAÇÃO DE LOGGING ---
	logDir := "./logs"
	_ = os.MkdirAll(logDir, 0755)
	logFile, _ := os.OpenFile(filepath.Join(logDir, "auren_system.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("### 🛠️  Iniciando sequência de boot: AUREN PLATFORM V1.4 ###")

	// 0. Carregamento de Ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aviso: Arquivo .env não localizado.")
	}

	appToken := os.Getenv("APP_SECRET_TOKEN")
	if appToken == "" {
		appToken = "auren-platform-DEMO"
	}

	// 1. Inicialização de Infraestrutura
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dirs := []string{"./storage/evidences", "./storage/knowledge", "./static"}
	for _, dir := range dirs {
		_ = os.MkdirAll(dir, 0755)
	}

	db.InitDB()

	// 1.1 Hub de comunicação (Core)
	// CORREÇÃO AQUI: Agora chamamos events.NewHub() em vez de realtime.NewHub()
	aurenHub := events.NewHub()
	go aurenHub.Run()

	// 1.2 Engine (Negócio) - Injetando o Hub do tipo *events.Hub
	aurenEngine, err := engine.NewAurenEngine(ctx, aurenHub)
	if err != nil {
		log.Fatalf("❌ CRÍTICO: Falha ao iniciar Engine: %v", err)
	}

	// 1.3 Lançamento de Agentes Autônomos
	go aurenEngine.HunterAgent.Run(ctx)
	go aurenEngine.RadarAgent.Run(ctx)

	// 1.4 Processadores de Mídia
	streamAudioProc := media_analysis.NewStreamAudioProcessor(40.0)
	videoProc := media_analysis.NewVideoProcessor()
	fileAudioProc := media_analysis.NewFileAudioProcessor()

	// 2. Roteamento
	r := mux.NewRouter()

	// API Protegida
	api := r.PathPrefix("/api").Subrouter()
	api.Use(adminsvc.SecurityMiddleware)
	api.HandleFunc("/execute", aurenEngine.HandleExecution).Methods("POST")
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "online"})
	}).Methods("GET")

	// WebSockets
	wsSub := r.PathPrefix("/ws").Subrouter()
	wsSub.Use(adminsvc.SecurityMiddleware)
	wsSub.HandleFunc("/engine", func(w http.ResponseWriter, r *http.Request) {
		agentsMap := map[string]any{
			"auditor":   aurenEngine.AuditorCampoAgent,
			"redator":   aurenEngine.RedatorPGRSAgent,
			"propostas": aurenEngine.PropostasAgent,
		}

		realtime.ServeWS(
			aurenHub,
			aurenEngine.Gemini,
			agentsMap,
			streamAudioProc,
			videoProc,
			fileAudioProc,
			w,
			r,
		)
	})

	// Servidor de Arquivos Estáticos (SPA Fallback)
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/ws") {
			return
		}

		path := r.URL.Path
		if path == "/" {
			path = "auren.html"
		}

		fullPath := filepath.Join("static", path)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, fullPath)
			return
		}

		http.ServeFile(w, r, filepath.Join("static", "auren.html"))
	}))

	// 3. Configuração do Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      r,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	// Banner com Token Visível
	localIP := getLocalIP()
	log.Println("==================================================")
	log.Println("🛡️  AUREN PLATFORM : SISTEMA DE CONFORMIDADE")
	log.Printf("🚀 ACESSO LOCAL: http://localhost:%s/auren.html?token=%s", port, appToken)
	log.Printf("🌍 ACESSO REDE:  http:// %s:%s/auren.html?token=%s", localIP, port, appToken)
	log.Printf("🔑 AUTH TOKEN:   %s", appToken)
	log.Println("==================================================")

	// 4. Graceful Shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("\n🛑 Encerrando Auren...")
		aurenEngine.Close()
		_ = server.Shutdown(ctx)
		os.Exit(0)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("❌ ERRO: %v", err)
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
