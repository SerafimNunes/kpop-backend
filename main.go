package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type CommentResponse struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Mensagem string `json:"mensagem"`
}

type FeedPost struct {
	ID          int               `json:"id"`
	Nome        string            `json:"nome"`
	Mensagem    string            `json:"mensagem"`
	Tag         string            `json:"tag"`
	Likes       int               `json:"likes"`
	Comentarios []CommentResponse `json:"comentarios"`
}

var db *sql.DB

func initDB() {
	godotenv.Load()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Erro ao abrir banco:", err)
	}

	// 4 TABELAS DO MVP
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS feed_posts (
			id SERIAL PRIMARY KEY, 
			usuario_nome VARCHAR(50), 
			mensagem TEXT, 
			tag VARCHAR(20), 
			criado_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS feed_comments (
			id SERIAL PRIMARY KEY, 
			post_id INTEGER REFERENCES feed_posts(id) ON DELETE CASCADE, 
			usuario_nome VARCHAR(50), 
			mensagem TEXT, 
			criado_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS post_reactions (
			post_id INTEGER PRIMARY KEY REFERENCES feed_posts(id) ON DELETE CASCADE, 
			likes INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS app_members (
			id SERIAL PRIMARY KEY, 
			nome VARCHAR(50), 
			plano VARCHAR(20)
		);`,
	}

	for _, q := range schemas {
		_, err := db.Exec(q)
		if err != nil {
			log.Println("Aviso na tabela:", err)
		}
	}
	fmt.Println("🐘 Base de Dados Sincronizada com 4 Tabelas!")
}

func main() {
	initDB()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	// LISTAR FEED
	http.HandleFunc("/api/feed", func(w http.ResponseWriter, r *http.Request) {
		// Query usando nomes explícitos para evitar erro de coluna
		rows, err := db.Query(`
			SELECT 
				p.id, 
				p.usuario_nome, 
				p.mensagem, 
				COALESCE(p.tag, 'Membro'), 
				COALESCE(pr.likes, 0)
			FROM feed_posts p
			LEFT JOIN post_reactions pr ON p.id = pr.post_id
			ORDER BY p.criado_at DESC`)
		
		if err != nil {
			log.Println("Erro na query de posts:", err)
			http.Error(w, "Erro ao buscar posts", 500)
			return
		}
		defer rows.Close()

		posts := []FeedPost{}
		for rows.Next() {
			var p FeedPost
			err := rows.Scan(&p.ID, &p.Nome, &p.Mensagem, &p.Tag, &p.Likes)
			if err != nil {
				log.Println("Erro no Scan:", err)
				continue
			}
			p.Comentarios = []CommentResponse{} 
			posts = append(posts, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	})

	// CRIAR POST
	http.HandleFunc("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { return }
		var p FeedPost
		json.NewDecoder(r.Body).Decode(&p)
		_, err := db.Exec("INSERT INTO feed_posts (usuario_nome, mensagem, tag) VALUES ($1, $2, $3)", p.Nome, p.Mensagem, p.Tag)
		if err != nil {
			log.Println("Erro ao inserir post:", err)
			http.Error(w, "Erro ao salvar", 500)
			return
		}
		w.WriteHeader(201)
	})

	// EDITAR POST
	http.HandleFunc("/api/edit-post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" { return }
		var req struct { ID int `json:"id"`; Mensagem string `json:"mensagem"` }
		json.NewDecoder(r.Body).Decode(&req)
		db.Exec("UPDATE feed_posts SET mensagem = $1 WHERE id = $2", req.Mensagem, req.ID)
		w.WriteHeader(200)
	})

	// DELETAR POST
	http.HandleFunc("/api/delete-post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" { return }
		var req struct { ID int `json:"id"` }
		json.NewDecoder(r.Body).Decode(&req)
		db.Exec("DELETE FROM feed_posts WHERE id = $1", req.ID)
		w.WriteHeader(200)
	})

	// REAÇÕES
	http.HandleFunc("/api/react-post", func(w http.ResponseWriter, r *http.Request) {
		var req struct { ID int `json:"id"` }
		json.NewDecoder(r.Body).Decode(&req)
		db.Exec(`INSERT INTO post_reactions (post_id, likes) VALUES ($1, 1) 
				 ON CONFLICT (post_id) DO UPDATE SET likes = post_reactions.likes + 1`, req.ID)
		w.WriteHeader(200)
	})

	fmt.Println("🚀 K-Pop Hub Online: http://localhost:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}