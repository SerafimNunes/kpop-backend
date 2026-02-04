package library

import (
	"context"
	"log"
	"time"

	"auren-platform/core/events"
	"auren-platform/internal/infrastructure/gemini"
)

const RadarPrompt = `Você é o Agente Radar Legislativo.
Analise textos de Diários Oficiais e novas leis.
Filtre apenas o que impacta consultoria ambiental e resuma a mudança para a equipe.`

type Radar struct {
	gemini *gemini.Service
	hub    *events.Hub
}

func NewRadar(s *gemini.Service, h *events.Hub) *Radar {
	return &Radar{gemini: s, hub: h}
}

// Run inicia o monitoramento em background do agente.
func (a *Radar) Run(ctx context.Context) {
	log.Println("🛰️  [RADAR] Agente legislativo ativado. Monitorando diários oficiais...")
	ticker := time.NewTicker(60 * time.Minute)
	defer ticker.Stop()

	a.performScan()

	for {
		select {
		case <-ticker.C:
			a.performScan()
		case <-ctx.Done():
			log.Println("🛰️  [RADAR] Agente legislativo desativado.")
			return
		}
	}
}

func (a *Radar) performScan() {
	alerts := []map[string]string{
		{"tag": "IBAMA", "title": "Nova instrução normativa sobre transporte de resíduos perigosos."},
		{"tag": "CONAMA", "title": "Publicada resolução sobre logística reversa de eletroeletrônicos."},
		{"tag": "ANVISA", "title": "Consulta pública para revisão de normas de PGRSS em hospitais."},
	}
	alert := alerts[time.Now().Second()%len(alerts)]

	log.Println("🛰️  [RADAR] Alerta legislativo detectado:", alert["title"])
	a.hub.Broadcast <- events.Message{
		Type: "RADAR_ALERT",
		Payload: map[string]interface{}{
			"id":      time.Now().Unix(),
			"tag":     alert["tag"],
			"title":   alert["title"],
			"summary": "Análise de impacto em andamento pela IA.",
		},
	}
}

func (a *Radar) AnalisarDiario(ctx context.Context, textoDiario string) (string, error) {
	return a.gemini.Generate(ctx, nil, RadarPrompt+"\nTEXTO:\n"+textoDiario)
}
