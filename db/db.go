package db

import (
	"fmt"
	"k-lens/models"
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

	// Mantendo o seu Logger.Info para facilitar o debug durante o desenvolvimento
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Falha ao conectar no banco de dados:", err)
	}

	// Sincroniza as tabelas respeitando os nomes que você definiu:
	// LiveArchive (Histórico de Lives) e CaptionLog (Logs de Legendas/Traduções)
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