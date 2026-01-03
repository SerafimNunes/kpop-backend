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
	ClipDuration int
	AspectRatio  string
}

type Cutter struct {
	StoragePath string
	CurrentConf Config
	// Canal para avisar quando um clipe fica pronto
	NotifyChan chan string
}

func NewCutter() *Cutter {
	path := "./recordings"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	return &Cutter{
		StoragePath: path,
		CurrentConf: Config{ClipDuration: 61, AspectRatio: "9:16"},
		NotifyChan:  make(chan string, 10),
	}
}

func (c *Cutter) GetStreamURL(youtubeURL string) ([]string, error) {
	log.Printf("🔍 [yt-dlp] Resolvendo URL: %s", youtubeURL)

	// Usa exec.LookPath para encontrar yt-dlp no PATH (Windows, Linux, macOS)
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		// Se yt-dlp não está no PATH, tenta usar direto o nome (Docker/Linux)
		ytDlpPath = "yt-dlp"
	}

	// Busca URLs separadas de vídeo e áudio
	cmd := exec.Command(ytDlpPath, "--no-playlist", "-g", "-f", "bestvideo+bestaudio/best", youtubeURL)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erro yt-dlp: %v", err)
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
		urls, err := c.GetStreamURL(youtubeURL)
		if err != nil || len(urls) == 0 {
			log.Printf("❌ [Cutter] Erro ao obter URLs: %v", err)
			return
		}

		// 2. Calcular ponto de início (5s antes do gatilho)
		startPoint := (timestamp / 1000.0) - 5.0
		if startPoint < 0 {
			startPoint = 0
		}

		safeRatio := strings.ReplaceAll(c.CurrentConf.AspectRatio, ":", "x")
		clipName := fmt.Sprintf("KLENS_%s_%s_%s.mp4", label, safeRatio, time.Now().Format("150405"))
		outputPath := filepath.Join(c.StoragePath, clipName)

		watermark := "K-LENS STUDIO"
		var videoFilter string
		if c.CurrentConf.AspectRatio == "9:16" {
			videoFilter = fmt.Sprintf("crop=ih*9/16:ih,unsharp=3:3:1.5:3:3:0.5,drawtext=text='%s':fontcolor=white@0.8:fontsize=24:x=(w-tw)/2:y=60:shadowcolor=black:shadowx=2:shadowy=2", watermark)
		} else {
			videoFilter = fmt.Sprintf("scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,drawtext=text='%s':fontcolor=white@0.8:fontsize=32:x=(w-tw)/2:y=50:shadowcolor=black:shadowx=2:shadowy=2", watermark)
		}

		// Argumentos otimizados para sincronia e reconexão
		args := []string{
			"-y",
			"-reconnect", "1", "-reconnect_at_eof", "1", "-reconnect_streamed", "1", "-reconnect_delay_max", "5",
			"-ss", fmt.Sprintf("%.2f", startPoint),
		}

		// Adiciona cada URL como um input separado (-i url1 -i url2)
		for _, u := range urls {
			args = append(args, "-i", strings.TrimSpace(u))
		}

		args = append(args,
			"-t", fmt.Sprintf("%d", c.CurrentConf.ClipDuration),
			"-filter_complex", "[0:v]"+videoFilter+"[outv]", // Filtro no vídeo do primeiro input
			"-map", "[outv]", // Usa o vídeo filtrado
			"-map", "1:a?", // Tenta pegar o áudio do segundo input
			"-map", "0:a?", // Fallback: pega áudio do primeiro se o segundo falhar
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-crf", "23",
			"-c:a", "aac",
			"-b:a", "128k",
			"-ar", "44100", // Força sample rate padrão para evitar chiado/troca
			"-shortest",
			outputPath,
		)

		log.Printf("🎬 [Cutter] Iniciando FFmpeg para: %s", clipName)
		cmd := exec.Command("ffmpeg", args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			log.Printf("❌ [Cutter] FFmpeg falhou: %v\nSaída: %s", err, string(output))
		} else {
			log.Printf("✅ [Cutter] Clipe concluído com sucesso: %s", clipName)
			c.NotifyChan <- clipName
		}
	}()
}
