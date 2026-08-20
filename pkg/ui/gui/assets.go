package gui

import (
	"bytes"
	_ "embed"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/logo.png
var logoPNG []byte

// Logo é a marca do PioKe, embutida no binário (em vez de lida de um caminho relativo em disco) —
// consistente com o objetivo do projeto de ser um binário único, leve, que roda igual não importa
// de onde é copiado (Raspberry Pi, Smart TV, etc.).
var logoImage = loadLogo()

func loadLogo() *ebiten.Image {
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(img)
}
