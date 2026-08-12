package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// WriteWAV salva dados PCM estéreo de 16-bit 44100Hz em um arquivo .wav
func WriteWAV(filename string, pcmData []byte, sampleRate int, numChannels int) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo WAV: %w", err)
	}
	defer file.Close()

	dataSize := uint32(len(pcmData))
	bytesPerSample := 2
	blockAlign := uint16(numChannels * bytesPerSample)
	byteRate := uint32(sampleRate * int(blockAlign))

	var header bytes.Buffer

	// RIFF Header
	header.WriteString("RIFF")
	_ = binary.Write(&header, binary.LittleEndian, uint32(36+dataSize))
	header.WriteString("WAVE")

	// fmt chunk
	header.WriteString("fmt ")
	_ = binary.Write(&header, binary.LittleEndian, uint32(16))          // Subchunk1Size (16 para PCM)
	_ = binary.Write(&header, binary.LittleEndian, uint16(1))           // AudioFormat (1 para PCM)
	_ = binary.Write(&header, binary.LittleEndian, uint16(numChannels)) // NumChannels
	_ = binary.Write(&header, binary.LittleEndian, uint32(sampleRate))  // SampleRate
	_ = binary.Write(&header, binary.LittleEndian, byteRate)            // ByteRate
	_ = binary.Write(&header, binary.LittleEndian, blockAlign)          // BlockAlign
	_ = binary.Write(&header, binary.LittleEndian, uint16(16))          // BitsPerSample

	// data chunk
	header.WriteString("data")
	_ = binary.Write(&header, binary.LittleEndian, dataSize)

	if _, err := file.Write(header.Bytes()); err != nil {
		return fmt.Errorf("erro ao escrever cabeçalho WAV: %w", err)
	}

	if _, err := file.Write(pcmData); err != nil {
		return fmt.Errorf("erro ao escrever dados PCM: %w", err)
	}

	return nil
}

// ReadWAV lê um arquivo .wav PCM 16-bit (o formato escrito por WriteWAV) e retorna os dados PCM
// crus junto da taxa de amostragem e número de canais declarados no cabeçalho.
func ReadWAV(filename string) (pcmData []byte, sampleRate int, numChannels int, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("falha ao ler arquivo WAV: %w", err)
	}

	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, 0, 0, fmt.Errorf("arquivo não é um WAV RIFF válido: %s", filename)
	}

	var channels uint16
	var rate uint32
	var bitsPerSample uint16
	var pcm []byte

	pos := 12
	for pos+8 <= len(data) {
		chunkID := string(data[pos : pos+4])
		chunkSize := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		chunkStart := pos + 8
		chunkEnd := min(chunkStart+int(chunkSize), len(data))

		switch chunkID {
		case "fmt ":
			if chunkEnd-chunkStart >= 16 {
				channels = binary.LittleEndian.Uint16(data[chunkStart+2 : chunkStart+4])
				rate = binary.LittleEndian.Uint32(data[chunkStart+4 : chunkStart+8])
				bitsPerSample = binary.LittleEndian.Uint16(data[chunkStart+14 : chunkStart+16])
			}
		case "data":
			pcm = data[chunkStart:chunkEnd]
		}

		pos = chunkEnd
		if chunkSize%2 == 1 {
			pos++ // chunks têm padding para tamanho par
		}
	}

	if pcm == nil {
		return nil, 0, 0, fmt.Errorf("chunk 'data' não encontrado em %s", filename)
	}
	if bitsPerSample != 0 && bitsPerSample != 16 {
		return nil, 0, 0, fmt.Errorf("formato WAV não suportado em %s: %d bits por amostra (esperado 16)", filename, bitsPerSample)
	}

	return pcm, int(rate), int(channels), nil
}
