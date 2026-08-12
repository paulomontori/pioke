package audio

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/gen2brain/malgo"
)

// MicRecorder captura áudio do microfone padrão em PCM mono 16-bit, acumulando em memória a
// partir do instante em que NewMicRecorder retorna, até Stop ser chamado.
type MicRecorder struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device

	mu  sync.Mutex
	buf bytes.Buffer
}

// NewMicRecorder inicializa o dispositivo de captura padrão do sistema e já começa a gravar.
func NewMicRecorder(sampleRate int) (*MicRecorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(string) {})
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar contexto de áudio: %w", err)
	}

	r := &MicRecorder{ctx: ctx}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = uint32(sampleRate)

	callbacks := malgo.DeviceCallbacks{
		Data: func(_, captured []byte, _ uint32) {
			r.mu.Lock()
			r.buf.Write(captured)
			r.mu.Unlock()
		},
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, callbacks)
	if err != nil {
		_ = ctx.Uninit()
		ctx.Free()
		return nil, fmt.Errorf("erro ao inicializar dispositivo de captura (microfone): %w", err)
	}

	if err := device.Start(); err != nil {
		device.Uninit()
		_ = ctx.Uninit()
		ctx.Free()
		return nil, fmt.Errorf("erro ao iniciar captura de microfone: %w", err)
	}

	r.device = device
	return r, nil
}

// Stop encerra a captura e libera o dispositivo, retornando o PCM mono 16-bit gravado até aqui.
func (r *MicRecorder) Stop() []byte {
	if r.device != nil {
		r.device.Uninit()
	}
	if r.ctx != nil {
		_ = r.ctx.Uninit()
		r.ctx.Free()
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.Bytes()
}
