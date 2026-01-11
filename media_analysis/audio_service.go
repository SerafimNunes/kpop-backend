package media_analysis

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os/exec"
)

type AudioProcessor struct {
	Threshold float64
	ZCRLimit  float64
	RecentRMS float64
}

func NewAudioProcessor(threshold float64) *AudioProcessor {
	return &AudioProcessor{
		Threshold: threshold,
		ZCRLimit:  0.55,
	}
}

// StartDigitalEar captura o áudio técnico da vistoria (via Stream de vídeo ou IP Camera)
func (ap *AudioProcessor) StartDigitalEar(ctx context.Context, streamURL string, outputChan chan []byte) error {
	log.Printf("🔗 [MEDIA-ANALYSIS] Iniciando escuta técnica na URL: %s", streamURL)

	// Comando otimizado para extrair PCM linear (Padrão para análise de IA)
	// Removemos 'cmd /C' para compatibilidade com ambientes de produção Linux/Docker
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", streamURL,
		"-f", "wav",
		"-ac", "1",
		"-ar", "16000",
		"-acodec", "pcm_s16le",
		"-loglevel", "error",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("erro pipe stdout: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar ffmpeg: %v", err)
	}

	go func() {
		defer func() {
			log.Println("🛑 [MEDIA-ANALYSIS] Finalizando captura de áudio.")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			close(outputChan)
		}()

		const (
			sampleRate     = 16000
			bytesPerSample = 2
			targetSeconds  = 20 // Reduzido de 30 para 20s para respostas mais rápidas no PGRS
			overlapSeconds = 2
		)

		targetSize := targetSeconds * sampleRate * bytesPerSample
		overlapSize := overlapSeconds * sampleRate * bytesPerSample

		readBuffer := make([]byte, 16384)
		var accumulatedAudio []byte

		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := stdout.Read(readBuffer)
				if err != nil {
					if err != io.EOF {
						log.Printf("❌ [MEDIA-ERR]: %v", err)
					}
					return
				}

				if n > 0 {
					accumulatedAudio = append(accumulatedAudio, readBuffer[:n]...)

					if len(accumulatedAudio) >= targetSize {
						chunk := make([]byte, targetSize)
						copy(chunk, accumulatedAudio[:targetSize])

						wavChunk := ensureWav(chunk)

						// Verificação VAD antes de enviar para a IA (Economia de tokens)
						if ap.ShouldProcess(wavChunk) {
							outputChan <- wavChunk
						} else {
							log.Println("🔇 [VAD] Ambiente ruidoso ou silêncio. Descartando segmento.")
						}

						remaining := accumulatedAudio[targetSize-overlapSize:]
						accumulatedAudio = make([]byte, len(remaining))
						copy(accumulatedAudio, remaining)
					}
				}
			}
		}
	}()

	return nil
}

// FetchWebText extrai conteúdo de bases normativas (Ex: Diário Oficial ou Leis)
func (ap *AudioProcessor) FetchWebText(ctx context.Context, url string) (string, error) {
	log.Printf("📄 [WEB-FETCH] Consultando base normativa: %s", url)
	cmd := exec.CommandContext(ctx, "curl", "-L", "-A", "AurenPlatform/1.0", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (ap *AudioProcessor) ShouldProcess(audioData []byte) bool {
	if len(audioData) <= 44 {
		return false
	}

	analysisData := audioData[44:]
	isSilent, rms := ap.AnalyzeSilence(analysisData)

	// Atualiza média móvel do ruído de fundo da unidade
	ap.RecentRMS = ap.RecentRMS*0.9 + rms*0.1

	if isSilent {
		return false
	}

	isMusic, zcr := ap.AnalyzeMusic(analysisData)
	if isMusic && zcr > 0.8 {
		return false
	}

	// Se o som for muito baixo em relação ao threshold, ignora
	return rms >= (ap.Threshold * 0.7)
}

func (ap *AudioProcessor) AnalyzeSilence(audioData []byte) (bool, float64) {
	var sum float64
	count := 0
	for i := 0; i < len(audioData)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(audioData[i : i+2]))
		sum += float64(sample) * float64(sample)
		count++
	}
	rms := math.Sqrt(sum / float64(count))
	return rms < ap.Threshold, rms
}

func (ap *AudioProcessor) AnalyzeMusic(audioData []byte) (bool, float64) {
	crossings := 0
	samples := 0
	var lastSample int16
	for i := 0; i < len(audioData)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(audioData[i : i+2]))
		if (sample > 0 && lastSample < 0) || (sample < 0 && lastSample > 0) {
			crossings++
		}
		lastSample = sample
		samples++
	}
	zcr := float64(crossings) / float64(samples)
	return zcr > ap.ZCRLimit, zcr
}

func ensureWav(data []byte) []byte {
	if len(data) >= 4 && bytes.HasPrefix(data, []byte("RIFF")) {
		return data
	}
	// Cabeçalho WAV padrão para 16kHz Mono 16-bit
	buf := new(bytes.Buffer)
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+len(data)))
	buf.WriteString("WAVEfmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint32(16000))
	binary.Write(buf, binary.LittleEndian, uint32(32000))
	binary.Write(buf, binary.LittleEndian, uint16(2))
	binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}
