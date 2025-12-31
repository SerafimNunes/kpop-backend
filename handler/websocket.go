package handler

import (
	"context"
	"kpop-backend/translate" // Certifique-se que o caminho do módulo está correto
	"log"
)

// ProcessarAudio cuida das mensagens de áudio (vinda do site/websocket)
func ProcessarAudio(apiKey string, audioData []byte) (string, error) {
	ctx := context.Background()

	traducao, err := translate.TraduzirLive(ctx, apiKey, audioData)
	if err != nil {
		log.Printf("Erro ao processar áudio no tradutor: %v", err)
		return "", err
	}

	return traducao, nil
}

// TraduzirTextoSimples cuida das mensagens de texto (vinda do seu novo App)
func TraduzirTextoSimples(apiKey string, texto string) (string, error) {
	ctx := context.Background()

	// Aqui chamamos o translate enviando o texto.
	// Vou assumir que vamos criar essa função 'TraduzirTexto' no seu pacote translate.
	traducao, err := translate.TraduzirTexto(ctx, apiKey, texto)
	if err != nil {
		log.Printf("Erro ao traduzir texto do app: %v", err)
		return "", err
	}

	return traducao, nil
}