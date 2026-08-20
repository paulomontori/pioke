package gui

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// Fonts agrupa as faces de texto escaláveis usadas pela tela de reprodução, em tamanhos fixos
// pensados pra leitura a distância (cenário de karaokê, possivelmente numa TV) — bem maiores que a
// fonte de debug de 7x13px usada antes.
type Fonts struct {
	Header  *text.GoTextFace // título/artista
	Line    *text.GoTextFace // linha atual, cantada
	Preview *text.GoTextFace // prévia da próxima linha
	Chord   *text.GoTextFace // badge de acorde
	Small   *text.GoTextFace // rodapé, barra de progresso
}

// NewFonts carrega a fonte Go Regular (já embarcada em golang.org/x/image, sem precisar de nenhum
// arquivo externo) uma única vez e monta as faces em cada tamanho usado na tela.
func NewFonts() *Fonts {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		// goregular.TTF é um recurso embarcado e constante — só falharia por corrupção do próprio
		// binário do Go, não por nada em tempo de execução do PioKe.
		panic(err)
	}

	face := func(size float64) *text.GoTextFace {
		return &text.GoTextFace{Source: source, Size: size}
	}

	return &Fonts{
		Header:  face(20),
		Line:    face(48),
		Preview: face(26),
		Chord:   face(40),
		Small:   face(16),
	}
}
