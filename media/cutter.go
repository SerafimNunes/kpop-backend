package media

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	ClipDuration int    // Segundos (ex: 61 para 1:01)
	AspectRatio  string // "9:16" ou "16:9"
}

type Cutter struct {
	StoragePath string
	CurrentConf Config
}

func NewCutter() *Cutter {
	path := "./recordings"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	return &Cutter{
		StoragePath: path,
		CurrentConf: Config{ClipDuration: 61, AspectRatio: "9:16"},
	}
}

// GetStreamURL revisado para lidar com URLs duplas (Vídeo + Áudio)
func (c *Cutter) GetStreamURL(youtubeURL string) ([]string, error) {
	log.Printf("🔍 [yt-dlp] Resolvendo URL: %s", youtubeURL)

	ytDlpPath := `C:\Users\seraf\AppData\Local\Packages\PythonSoftwareFoundation.Python.3.13_qbz5n2kfra8p0\LocalCache\local-packages\Python313\Scripts\yt-dlp.exe`

	// Usamos -g para pegar as URLs. O yt-dlp retorna vídeo e áudio em linhas separadas.
	cmd := exec.Command(ytDlpPath, "-g", "-f", "bestvideo+bestaudio/best", youtubeURL)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erro ao executar yt-dlp: %v", err)
	}

	urls := strings.Split(strings.TrimSpace(string(out)), "\n")
	return urls, nil
}

func (c *Cutter) UpdateConfig(duration int, ratio string) {
	c.CurrentConf.ClipDuration = duration
	c.CurrentConf.AspectRatio = ratio
	log.Printf("⚙️ [Config] Nova meta de produção: %ds em %s", duration, ratio)
}

func (c *Cutter) CreateClip(liveID string, youtubeURL string, timestamp float64, label string) {
	go func() {
		// 1. Obter as URLs (pode retornar 1 ou 2 URLs)
		urls, err := c.GetStreamURL(youtubeURL)
		if err != nil || len(urls) == 0 {
			log.Printf("❌ [Cutter] Falha ao obter stream: %v", err)
			return
		}

		// 2. Calcular ponto de início
		startPoint := (timestamp / 1000.0) - 5.0
		if startPoint < 0 {
			startPoint = 0
		}

		// 3. Preparar nomes e caminhos
		safeRatio := strings.ReplaceAll(c.CurrentConf.AspectRatio, ":", "x")
		clipName := fmt.Sprintf("KLENS_%s_%s_%s.mp4", label, safeRatio, time.Now().Format("150405"))
		outputPath := filepath.Join(c.StoragePath, clipName)

		// 4. Filtros de Vídeo
		var videoFilter string
		if c.CurrentConf.AspectRatio == "9:16" {
			videoFilter = "crop=ih*9/16:ih"
		} else {
			videoFilter = "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2"
		}

		// 5. Montar argumentos do FFmpeg com suporte a Reconnect e Input Duplo
		args := []string{
			"-y",
			"-reconnect", "1",
			"-reconnect_at_eof", "1",
			"-reconnect_streamed", "1",
			"-reconnect_delay_max", "5",
			"-ss", fmt.Sprintf("%.2f", startPoint),
		}

		// Adiciona cada URL encontrada como um input separado
		for _, u := range urls {
			args = append(args, "-i", strings.TrimSpace(u))
		}

		// Complementa com filtros e codecs
		args = append(args,
			"-t", fmt.Sprintf("%d", c.CurrentConf.ClipDuration),
			"-vf", videoFilter,
			"-c:v", "libx264",
			"-preset", "ultrafast", // Mudado para ultrafast para aliviar o CPU
			"-crf", "23",
			"-c:a", "aac",
			"-shortest", // Garante que o vídeo pare quando a menor trilha acabar
			outputPath,
		)

		cmd := exec.Command("ffmpeg", args...)

		log.Printf("🎬 [Cutter] Recortando: %s", clipName)

		if err := cmd.Run(); err != nil {
			log.Printf("❌ [Cutter] Erro FFmpeg: %v", err)
		} else {
			log.Printf("✅ [Cutter] Sucesso! Salvo em: %s", outputPath)
		}
	}()
}
