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
