package db

import (
	"fmt"
	"kpop-backend/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Falha ao conectar no banco de dados:", err)
	}

	// Sincroniza as tabelas de acordo com os arquivos em /models
	err = db.AutoMigrate(
		&models.User{},
		&models.LiveArchive{},
		&models.CaptionLog{},
	)
	if err != nil {
		log.Fatal("Erro ao sincronizar tabelas (AutoMigrate):", err)
	}

	DB = db
	log.Println("🐘 Banco sincronizado com sucesso!")
}