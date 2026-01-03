package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"k-lens/db"
	"k-lens/handler"
	"k-lens/hub"
	"k-lens/translate"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não carregado")
	}

	ctx := context.Background()
	geminiKey := os.Getenv("GEMINI_API_KEY")
	var geminiSvc *translate.GeminiService
	var err error

	if geminiKey != "" {
		geminiSvc, err = translate.NewGeminiService(ctx, geminiKey)
		if err != nil {
			log.Printf("Erro Gemini: %v", err)
		} else {
			defer geminiSvc.Close()
		}
	}

	db.InitDB()
	legendasHub := hub.NewHub()
	go legendasHub.Run()

	r := mux.NewRouter()

	r.HandleFunc("/ws/studio/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWS(legendasHub, geminiSvc, w, r)
	})

	r.HandleFunc("/api/translate-reverse", handler.ReverseTranslate).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	localIP := getLocalIP()
	log.Printf("==========================================")
	log.Printf("🚀 K-LENS ARMY STUDIO (FFmpeg 8.0.1)")
	log.Printf("🔗 PC: http://localhost:%s/studio.html", port)
	log.Printf("📱 MOBILE: http://%s:%s/studio.html", localIP, port)
	log.Printf("==========================================")

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, r))
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
					return ipStr
				}
			}
		}
	}
	return "localhost"
}
