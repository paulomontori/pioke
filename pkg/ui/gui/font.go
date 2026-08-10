package gui

import (
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// FontManager gerencia o tamanho e medição das fontes na GUI
type FontManager struct {
	face font.Face
}

func NewFontManager() *FontManager {
	return &FontManager{
		face: basicfont.Face7x13,
	}
}

// MeasureString retorna a largura e altura aproximadas em pixels de um texto
func (fm *FontManager) MeasureString(text string) (width int, height int) {
	bounds, _ := font.BoundString(fm.face, text)
	width = (bounds.Max.X - bounds.Min.X).Ceil()
	height = (bounds.Max.Y - bounds.Min.Y).Ceil()
	return width, height
}

// DrawText desenha um texto em uma imagem/buffer simples
func (fm *FontManager) DrawText(dot fixed.Point26_6, text string, col color.Color) {
	// Auxiliar de medição/posição
}
