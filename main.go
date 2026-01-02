package main

import (
	"context"
	"log"
	"net"
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
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não carregado, usando variáveis de ambiente")
	}

	ctx := context.Background()
	geminiKey := os.Getenv("GEMINI_API_KEY")
	var geminiSvc *translate.GeminiService
	var err error

	if geminiKey != "" {
		geminiSvc, err = translate.NewGeminiService(ctx, geminiKey)
		if err != nil {
			log.Printf("Atenção: Gemini não iniciado: %v", err)
		} else {
			defer geminiSvc.Close()
		}
	}

	db.InitDB()
	legendasHub := hub.NewHub()
	go legendasHub.Run()

	r := mux.NewRouter()

	// Rota WebSocket para legendas e áudio
	r.HandleFunc("/ws/studio/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWS(legendasHub, geminiSvc, w, r)
	})

	// Rota para o Chat Reverso (Fã -> Idol)
	r.HandleFunc("/api/translate-reverse", handler.ReverseTranslate).Methods("POST")

	// Servidor de arquivos estáticos (HTML/JS/CSS)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	localIP := getLocalIP()
	log.Printf("==========================================")
	log.Printf("🚀 K-STUDIO PRO INICIADO")
	log.Printf("🔗 PC: http://localhost:%s/studio.html", port)
	log.Printf("📱 MOBILE: http://%s:%s/studio.html", localIP, port)
	log.Printf("==========================================")

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, r))
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				// Filtra IPs de redes locais comuns
				if len(ipStr) >= 7 && (ipStr[:7] == "192.168" || ipStr[:3] == "10." || ipStr[:3] == "172.") {
					return ipStr
				}
			}
		}
	}
	return "localhost"
}
