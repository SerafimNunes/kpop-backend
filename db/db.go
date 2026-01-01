package db

import (
	"fmt"
	"kpop-backend/models" // Ajuste para o nome do seu módulo no go.mod
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Falha ao conectar no banco:", err)
	}

	// Migra as tabelas automaticamente de acordo com a nossa Fonte da Verdade
	DB.AutoMigrate(
		&models.User{},
		&models.LiveArchive{},
		&models.CaptionLog{},
		&models.TranslationPoll{},
	)
	fmt.Println("Banco de dados sincronizado!")
}
