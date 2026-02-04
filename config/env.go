package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não encontrado, carregando variáveis de ambiente do ambiente")
	}
}
