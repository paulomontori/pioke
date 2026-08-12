package synth

import (
	"fmt"
	"slices"

	"pioke/pkg/model"
)

// Timbre seleciona qual motor de síntese RenderSongWithTimbre usa pra uma música.
type Timbre string

const (
	// TimbreAdditive é a síntese aditiva (fundamental + harmônicos, ver HarmonicAmplitudes) —
	// timbre padrão, som contínuo/legato entre notas adjacentes.
	TimbreAdditive Timbre = "additive"
	// TimbreKarplus é Karplus-Strong (corda dedilhada) — cada nota dedilha sua própria corda do
	// zero, sem herdar fase da nota anterior.
	TimbreKarplus Timbre = "karplus"
)

// Timbres lista os valores válidos de Timbre, na ordem em que devem aparecer em UI/ajuda de CLI.
var Timbres = []Timbre{TimbreAdditive, TimbreKarplus}

// ParseTimbre converte uma string (ex: vinda de flag de linha de comando) em Timbre, com
// TimbreAdditive como padrão quando vazia. Retorna erro se o valor não for reconhecido.
func ParseTimbre(s string) (Timbre, error) {
	if s == "" {
		return TimbreAdditive, nil
	}
	t := Timbre(s)
	if slices.Contains(Timbres, t) {
		return t, nil
	}
	return "", fmt.Errorf("timbre desconhecido: %q (opções: %v)", s, Timbres)
}

// RenderSongWithTimbre sintetiza a música inteira (melodia + acompanhamento) com o motor de
// síntese escolhido por timbre — mesma música, dois timbres diferentes, mesma API de resto.
func RenderSongWithTimbre(s *model.Song, timbre Timbre) []byte {
	if timbre == TimbreKarplus {
		return RenderSongKS(s)
	}
	return RenderSong(s)
}

// HarmonicAmplitudes define o timbre aditivo usado por RenderSequence: cada elemento é o peso
// (amplitude relativa) do harmônico correspondente à sua posição — índice 0 é a fundamental
// (1×f), índice 1 é o 2º harmônico (2×f), índice 2 o 3º (3×f), e assim por diante. O número de
// harmônicos é simplesmente o tamanho do slice.
//
// Os valores abaixo aproximam o timbre de uma corda dedilhada (violão/cavaco): harmônicos graves
// fortes e decaindo progressivamente, em vez da fundamental pura (que soa como um bipe
// eletrônico). Não precisam somar 1 nem seguir nenhuma fórmula em particular — RenderSequence
// normaliza cada nota pela soma dos pesos realmente audíveis, então a amplitude final nunca
// estoura, seja qual for a combinação escolhida aqui.
//
// Pra ajustar o timbre, edite este slice diretamente:
//   - Menos harmônicos (ex: só 2-3) ou decaimento mais rápido => som mais "redondo"/apagado,
//     mais perto de uma fundamental pura.
//   - Mais harmônicos ou decaimento mais lento => som mais brilhante/metálico.
//   - Zerar harmônicos pares/ímpares específicos muda o caráter do timbre (ex: só ímpares lembra
//     mais um clarinete; a série completa lembra mais uma corda).
var HarmonicAmplitudes = []float64{1.0, 0.55, 0.3, 0.18, 0.1, 0.05}
