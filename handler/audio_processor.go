package handler

import (
	"math"
)

// AudioProcessor lida com a análise primária do sinal para economizar API
type AudioProcessor struct {
	Threshold float64 // Limite de volume (RMS)
	ZCRLimit  float64 // Limite de Zero-Crossing Rate para diferenciar voz de música/ruído
}

func NewAudioProcessor(threshold float64) *AudioProcessor {
	return &AudioProcessor{
		Threshold: threshold,
		ZCRLimit:  0.15, // Valor base para voz humana em PCM16
	}
}

// IsSilent analisa se o buffer está abaixo do volume aceitável (VAD)
func (ap *AudioProcessor) IsSilent(audioData []byte) bool {
	if len(audioData) == 0 {
		return true
	}

	var sum float64
	samples := 0
	for i := 0; i < len(audioData); i += 2 {
		if i+1 < len(audioData) {
			sample := int16(audioData[i]) | int16(audioData[i+1])<<8
			sum += float64(sample) * float64(sample)
			samples++
		}
	}

	rms := math.Sqrt(sum / float64(samples))
	return rms < ap.Threshold
}

// IsMusic detecta se o áudio é música/bateria ou apenas ruído rítmico.
// Usa Zero-Crossing Rate (ZCR) para identificar a complexidade do sinal.
func (ap *AudioProcessor) IsMusic(audioData []byte) bool {
	if len(audioData) < 2 {
		return false
	}

	crossings := 0
	samples := 0
	var lastSample int16

	for i := 0; i < len(audioData); i += 2 {
		if i+1 < len(audioData) {
			sample := int16(audioData[i]) | int16(audioData[i+1])<<8

			// Detecta quando a onda cruza o eixo zero (ZCR)
			if (sample > 0 && lastSample < 0) || (sample < 0 && lastSample > 0) {
				crossings++
			}
			lastSample = sample
			samples++
		}
	}

	zcr := float64(crossings) / float64(samples)

	// Música e bateria tendem a ter ZCR muito alto ou muito constante.
	// Voz humana é irregular e fica geralmente entre 0.05 e 0.20.
	// Se for muito alto (> 0.4), geralmente é prato de bateria ou ruído branco (estática).
	if zcr > 0.4 {
		return true
	}

	return false
}

// ShouldProcess decide se o pacote de áudio merece ser enviado para a IA
func (ap *AudioProcessor) ShouldProcess(audioData []byte) bool {
	// Se estiver em silêncio OU for detectado como ruído/música excessiva, ignora.
	if ap.IsSilent(audioData) {
		return false
	}

	// Se o ZCR indicar que é apenas batida (bateria) sem voz clara, ignora.
	if ap.IsMusic(audioData) {
		return false
	}

	return true
}
