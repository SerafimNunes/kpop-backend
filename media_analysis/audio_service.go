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
	"time"
)

// StreamAudioProcessor handles real-time audio stream processing.
type StreamAudioProcessor struct {
	Threshold float64
	ZCRLimit  float64
	RecentRMS float64
}

// NewStreamAudioProcessor creates a new StreamAudioProcessor.
func NewStreamAudioProcessor(threshold float64) *StreamAudioProcessor {
	return &StreamAudioProcessor{
		Threshold: threshold,
		ZCRLimit:  0.55,
	}
}

// StartDigitalEar captures audio from a stream URL, now optimized for YouTube with yt-dlp.
func (sap *StreamAudioProcessor) StartDigitalEar(ctx context.Context, streamURL string, outputChan chan []byte) error {
	log.Printf("🔗 [MEDIA-ANALYSIS] Iniciando escuta técnica com yt-dlp: %s", streamURL)

	// Using a shell to create a pipe between yt-dlp and ffmpeg.
	// yt-dlp gets the best audio stream and outputs it to stdout.
	// ffmpeg reads from stdin (pipe:0) and processes the audio.
	cmdStr := fmt.Sprintf("yt-dlp -f bestaudio -o - \"%s\" | ffmpeg -i pipe:0 -f wav -ac 1 -ar 16000 -acodec pcm_s16le -loglevel error pipe:1", streamURL)
	
	// exec.Command for shell execution is different between Windows and Unix-like systems.
	// For simplicity and assuming a Linux/macOS deployment environment for this kind of tool.
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("erro ao criar pipe stdout: %v", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar pipe yt-dlp/ffmpeg: %v | stderr: %s", err, stderr.String())
	}

	go func() {
		defer func() {
			log.Println("🛑 [MEDIA-ANALYSIS] Finalizando captura de áudio e limpando processos.")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			close(outputChan)
		}()

		const (
			sampleRate     = 16000
			bytesPerSample = 2
			targetSeconds  = 30 // Aumentado para 30 segundos conforme solicitado
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
						log.Printf("❌ [MEDIA-ERR] Erro na leitura do stream: %v", err)
					}
					return
				}

				if n > 0 {
					accumulatedAudio = append(accumulatedAudio, readBuffer[:n]...)
					if len(accumulatedAudio) >= targetSize {
						chunk := make([]byte, targetSize)
						copy(chunk, accumulatedAudio[:targetSize])
						wavChunk := ensureWav(chunk)

						if sap.ShouldProcess(wavChunk) {
							select {
							case outputChan <- wavChunk:
								log.Println("🎙️ [VAD] Segmento de áudio de 30s enviado para IA.")
							case <-time.After(2 * time.Second):
								log.Println("⚠️ [VAD] Timeout ao enviar para o canal. Canal cheio?")
							}
						} else {
							log.Println("🔇 [VAD] Ambiente ruidoso ou silêncio detectado. Descartando segmento.")
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

func (sap *StreamAudioProcessor) ShouldProcess(audioData []byte) bool {
	if len(audioData) <= 44 {
		return false
	}
	analysisData := audioData[44:]
	isSilent, rms := sap.AnalyzeSilence(analysisData)
	sap.RecentRMS = sap.RecentRMS*0.9 + rms*0.1
	if isSilent {
		return false
	}
	isMusic, zcr := sap.AnalyzeMusic(analysisData)
	if isMusic && zcr > 0.8 {
		return false
	}
	return rms >= (sap.Threshold * 0.7)
}

func (sap *StreamAudioProcessor) AnalyzeSilence(audioData []byte) (bool, float64) {
	var sum float64
	count := 0
	for i := 0; i < len(audioData)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(audioData[i : i+2]))
		sum += float64(sample) * float64(sample)
		count++
	}
	if count == 0 {
		return true, 0
	}
	rms := math.Sqrt(sum / float64(count))
	return rms < sap.Threshold, rms
}

func (sap *StreamAudioProcessor) AnalyzeMusic(audioData []byte) (bool, float64) {
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
	if samples == 0 {
		return false, 0
	}
	zcr := float64(crossings) / float64(samples)
	return zcr > sap.ZCRLimit, zcr
}

// --- File Audio Processor ---

// FileAudioProcessor handles the analysis of single audio files.
type FileAudioProcessor struct {
	vad *VADProcessor
}

// NewFileAudioProcessor creates a new FileAudioProcessor.
func NewFileAudioProcessor() *FileAudioProcessor {
	return &FileAudioProcessor{
		vad: NewVADProcessor(),
	}
}

// ProcessAudioFile takes audio data and checks for voice activity.
// It returns a classification: "voice", "noise", "silence", or "music".
func (fap *FileAudioProcessor) ProcessAudioFile(audioData []byte) (string, error) {
	if len(audioData) == 0 {
		return "", fmt.Errorf("audio data is empty")
	}
	classification := fap.vad.ProcessAudioChunk(audioData)
	fap.vad.Reset()
	return classification, nil
}

// --- Common Functions ---

func ensureWav(data []byte) []byte {
	if len(data) >= 4 && bytes.HasPrefix(data, []byte("RIFF")) {
		return data
	}
	buf := new(bytes.Buffer)
	// ... (WAV header writing logic remains the same)
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+len(data)))
	buf.WriteString("WAVEfmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))     // PCM
	binary.Write(buf, binary.LittleEndian, uint16(1))     // Mono
	binary.Write(buf, binary.LittleEndian, uint32(16000)) // Sample Rate
	binary.Write(buf, binary.LittleEndian, uint32(32000)) // Byte Rate
	binary.Write(buf, binary.LittleEndian, uint16(2))     // Block Align
	binary.Write(buf, binary.LittleEndian, uint16(16))    // Bits per Sample
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

// FetchWebText is a utility function that could be used by multiple services.
// Keeping it separate for now.
func FetchWebText(ctx context.Context, url string) (string, error) {
	log.Printf("📄 [WEB-FETCH] Consultando base normativa: %s", url)
	cmd := exec.CommandContext(ctx, "curl", "-L", "-A", "AurenPlatform/1.0", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
