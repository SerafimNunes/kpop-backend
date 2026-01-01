package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"kpop-backend/db"
	"kpop-backend/handler"
	"kpop-backend/hub"
	"kpop-backend/translate"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	ctx := context.Background()

	// 2. Inicializa o Banco de Dados (Postgres + GORM)
	db.InitDB()

	// 3. Inicializa o Gemini Service (Refinador)
	// Passamos a chave uma única vez para o serviço persistente
	geminiSvc, err := translate.NewGeminiService(ctx, os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatalf("Erro ao iniciar Gemini: %v", err)
	}
	defer geminiSvc.Close()

	// 4. Inicializa o Hub do WebSocket (Gerenciador de salas)
	legendasHub := hub.NewHub()
	go legendasHub.Run()

	// 5. Configura as Rotas
	r := mux.NewRouter()

	// Passamos o hub e o serviço de tradução para o handler
	r.HandleFunc("/ws/live/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Note que agora o handler precisa receber o geminiSvc também
		handler.ServeWS(legendasHub, geminiSvc, w, r)
	})

	// Servir arquivos estáticos
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor iniciado na porta %s 🚀", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
