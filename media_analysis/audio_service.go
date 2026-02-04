package media_analysis

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os/exec"
)

type StreamAudioProcessor struct {
	Threshold float64
	ZCRLimit  float64
	RecentRMS float64
}

func NewStreamAudioProcessor(threshold float64) *StreamAudioProcessor {
	return &StreamAudioProcessor{
		Threshold: threshold,
		ZCRLimit:  0.55,
	}
}

// StartDigitalEar escuta streams (Youtube/RTMP) e segmenta para a IA
func (sap *StreamAudioProcessor) StartDigitalEar(ctx context.Context, streamURL string, outputChan chan []byte) error {
	log.Printf("🎙️ [MEDIA] Iniciando captura otimizada: %s", streamURL)

	// Pipeline de alto desempenho: Extração -> Conversão PCM -> Mono 16kHz
	cmdStr := fmt.Sprintf("yt-dlp -f bestaudio -o - \"%s\" | ffmpeg -i pipe:0 -f wav -ac 1 -ar 16000 -acodec pcm_s16le -loglevel error pipe:1", streamURL)
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha no pipeline media_analysis: %v", err)
	}

	go func() {
		defer func() {
			log.Println("🛑 [MEDIA] Captura encerrada.")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			close(outputChan)
		}()

		const (
			sampleRate     = 16000
			targetSeconds  = 30
			overlapSeconds = 2
		)
		targetSize := targetSeconds * sampleRate * 2
		accumulatedAudio := make([]byte, 0, targetSize)
		readBuffer := make([]byte, 16384)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := stdout.Read(readBuffer)
				if n > 0 {
					accumulatedAudio = append(accumulatedAudio, readBuffer[:n]...)
					if len(accumulatedAudio) >= targetSize {
						// Processamento VAD antes de enviar para a IA (Economia de Tokens)
						if sap.ShouldProcess(accumulatedAudio[:targetSize]) {
							wavChunk := ensureWav(accumulatedAudio[:targetSize])
							outputChan <- wavChunk
						}
						// Manter overlap para não perder palavras cortadas
						accumulatedAudio = accumulatedAudio[targetSize-(overlapSeconds*sampleRate*2):]
					}
				}
				if err != nil {
					return
				}
			}
		}
	}()
	return nil
}

func (sap *StreamAudioProcessor) ShouldProcess(audioData []byte) bool {
	// Pula cabeçalho WAV se houver
	data := audioData
	if len(data) > 44 {
		data = data[44:]
	}

	isSilent, rms := sap.AnalyzeSilence(data)
	if isSilent {
		return false
	}

	isMusic, _ := sap.AnalyzeMusic(data)
	if isMusic {
		return false
	} // Filtro de rádio/música ambiente em vistorias

	return rms >= (sap.Threshold * 0.5)
}

func (sap *StreamAudioProcessor) AnalyzeSilence(data []byte) (bool, float64) {
	var sum float64
	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		sum += float64(sample) * float64(sample)
	}
	rms := math.Sqrt(sum / (float64(len(data)) / 2))
	return rms < sap.Threshold, rms
}

func (sap *StreamAudioProcessor) AnalyzeMusic(data []byte) (bool, float64) {
	crossings := 0
	var lastSample int16
	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		if (sample > 0 && lastSample < 0) || (sample < 0 && lastSample > 0) {
			crossings++
		}
		lastSample = sample
	}
	zcr := float64(crossings) / (float64(len(data)) / 2)
	return zcr > sap.ZCRLimit, zcr
}

func ensureWav(data []byte) []byte {
	if bytes.HasPrefix(data, []byte("RIFF")) {
		return data
	}
	buf := new(bytes.Buffer)
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+len(data)))
	buf.WriteString("WAVEfmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))     // PCM
	binary.Write(buf, binary.LittleEndian, uint16(1))     // Mono
	binary.Write(buf, binary.LittleEndian, uint32(16000)) // 16kHz
	binary.Write(buf, binary.LittleEndian, uint32(32000)) // Byte Rate
	binary.Write(buf, binary.LittleEndian, uint16(2))     // Block Align
	binary.Write(buf, binary.LittleEndian, uint16(16))    // 16-bit
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

// --- Processador para Arquivos de Áudio (Upload Direto) ---

// FileAudioProcessor encapsula a lógica para análise de arquivos de áudio estáticos.
type FileAudioProcessor struct {
	vad *VADProcessor
}

// NewFileAudioProcessor cria um novo processador de arquivos de áudio.
func NewFileAudioProcessor() *FileAudioProcessor {
	return &FileAudioProcessor{
		vad: NewVADProcessor(),
	}
}

// ProcessAudioFile classifica um buffer de áudio de arquivo (e.g., .wav, .mp3)
// em 'voice', 'music', ou 'noise'.
func (fap *FileAudioProcessor) ProcessAudioFile(audioData []byte) (string, error) {
	// Aqui, poderíamos usar FFmpeg para converter o áudio para o formato PCM 16kHz
	// que o VAD espera. Por simplicidade, vamos assumir que o áudio já está
	// em um formato compatível ou que o VAD é robusto o suficiente.
	log.Println("🎙️ [MEDIA] Classificando arquivo de áudio...")

	// A lógica de classificação é delegada ao VAD (Voice Activity Detection)
	classification := fap.vad.ProcessAudioChunk(audioData)

	log.Printf("🎙️ [MEDIA] Classificação do arquivo: %s", classification)
	return classification, nil
}
