package media_analysis

import (
	"log"
	"math"
)

// VADProcessor isola a voz humana do ruído industrial
type VADProcessor struct {
	RMSThreshold      float64
	SilenceThreshold  float64
	MinFramesForVoice int
	FrameCount        int
	VoiceFrames       int
}

func NewVADProcessor() *VADProcessor {
	return &VADProcessor{
		RMSThreshold:      1200.0, // Aumentado para ignorar ruído de fundo de galpões
		SilenceThreshold:  300.0,
		MinFramesForVoice: 4, // Mais rigoroso para evitar gatilhos falsos por batidas metálicas
	}
}

// CalculateRMS analisa a potência do sinal de áudio da vistoria
func (vad *VADProcessor) CalculateRMS(audioBuffer []byte) float64 {
	if len(audioBuffer) < 2 {
		return 0
	}

	var sumSquares float64
	sampleCount := 0

	for i := 0; i < len(audioBuffer)-1; i += 2 {
		sample := int16(audioBuffer[i]) | (int16(audioBuffer[i+1]) << 8)
		val := float64(sample)
		sumSquares += val * val
		sampleCount++
	}

	return math.Sqrt(sumSquares / float64(sampleCount))
}

// IsVoicePresent verifica se o padrão de áudio condiz com fala humana
func (vad *VADProcessor) IsVoicePresent(audioBuffer []byte) bool {
	rms := vad.CalculateRMS(audioBuffer)

	if rms > vad.RMSThreshold {
		vad.VoiceFrames++
		if vad.VoiceFrames == vad.MinFramesForVoice {
			log.Println("🗣️ [VAD] Voz humana detectada na unidade.")
		}
		return vad.VoiceFrames >= vad.MinFramesForVoice
	}

	vad.VoiceFrames = 0
	return false
}

// ProcessAudioChunk decide se o chunk deve ser enviado para a IA do PGRS
func (vad *VADProcessor) ProcessAudioChunk(audioBuffer []byte) string {
	if len(audioBuffer) == 0 {
		return "silence"
	}

	rms := vad.CalculateRMS(audioBuffer)

	if rms < vad.SilenceThreshold {
		return "silence"
	}

	// Detecção de música (ignorar rádios ligados em pátios)
	if rms > 2000 && vad.calculateStdDev(audioBuffer, rms) < 300 {
		return "music"
	}

	if vad.IsVoicePresent(audioBuffer) {
		return "voice"
	}

	return "noise"
}

func (vad *VADProcessor) calculateStdDev(audioBuffer []byte, mean float64) float64 {
	var sumSquareDiffs float64
	count := 0

	for i := 0; i < len(audioBuffer)-1; i += 2 {
		sample := int16(audioBuffer[i]) | (int16(audioBuffer[i+1]) << 8)
		diff := float64(sample) - mean
		sumSquareDiffs += diff * diff
		count++
	}
	return math.Sqrt(sumSquareDiffs / float64(count))
}

func (vad *VADProcessor) Reset() {
	vad.FrameCount = 0
	vad.VoiceFrames = 0
}
