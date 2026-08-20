package gui

import (
	"strings"

	"pioke/pkg/model"
)

// defaultEventDurationMS é usado quando um TimelineEvent sem Syllables não informa DurationMS —
// mesmo fallback usado pelo cálculo de duração total em pkg/engine/engine.go, pra manter os dois
// consistentes.
const defaultEventDurationMS = 2000

// lyricSpan é um trecho de letra (uma sílaba, ou uma palavra/evento inteiro quando a música não
// tem sílabas marcadas) com o intervalo de tempo absoluto (em ms, na linha do tempo da música) em
// que ele é cantado.
type lyricSpan struct {
	Text    string
	StartMS int64
	EndMS   int64
}

// buildLyricSpans achata a timeline inteira de uma música num fluxo cronológico único de spans.
// Duas músicas de exemplo do projeto usam estilos de autoria diferentes: eventos de linha inteira
// com Syllables aninhadas (songs/parabens.yaml), ou eventos de fragmento de palavra sem Syllables
// (songs/evidencias.json) — por isso a tela de karaokê não confia nas fronteiras de TimelineEvent
// como fronteiras de linha visual: aqui só produzimos o fluxo de spans; a quebra em linhas de tela
// é feita depois, por largura de texto (ver wrapSpansIntoLines).
func buildLyricSpans(s *model.Song) []lyricSpan {
	if s == nil {
		return nil
	}

	// A separação entre palavras (e entre frases) é sempre a que já vem embutida no próprio texto
	// (espaço no fim de uma sílaba/evento) — nunca inferida daqui a partir de onde um TimelineEvent
	// termina. Um evento não é necessariamente uma "linha" ou uma "palavra completa": arquivos
	// vindos de MusicXML (pkg/parser/musicxml.go) cortam em compassos, que tanto podem terminar no
	// meio de uma palavra quanto no fim de uma frase — só o próprio parser (via <syllabic>) sabe
	// dizer qual dos dois é o caso, e por isso é ele quem embute o espaço quando necessário.
	var spans []lyricSpan
	for _, ev := range s.Timeline {
		if len(ev.Syllables) > 0 {
			for _, syl := range ev.Syllables {
				if syl.Text == "" {
					continue
				}
				start := ev.TimeMS + syl.OffsetMS
				dur := syl.DurationMS
				if dur <= 0 {
					dur = defaultEventDurationMS
				}
				spans = append(spans, lyricSpan{Text: syl.Text, StartMS: start, EndMS: start + dur})
			}
			continue
		}

		if ev.Lyric == "" {
			continue
		}
		dur := ev.DurationMS
		if dur <= 0 {
			dur = defaultEventDurationMS
		}
		start, end := ev.TimeMS, ev.TimeMS+dur

		// Marcadores de seção (ex: "[INTRO]") não são sílabas cantadas — em arquivos de acorde
		// como songs/evidencias.json, o mesmo marcador se repete em vários eventos seguidos (um
		// por troca de acorde) descrevendo o trecho instrumental inteiro. Sem tratamento especial
		// eles apareciam colados um no outro ("[INTRO][INTRO][INTRO]..."); aqui mescla repetições
		// consecutivas do mesmo marcador num span só, e garante espaço depois dele (marcador nunca
		// deve colar no texto seguinte, seja outro marcador ou letra de verdade).
		if isSectionMarker(ev.Lyric) {
			if n := len(spans); n > 0 && spans[n-1].Text == ev.Lyric+" " {
				spans[n-1].EndMS = end
				continue
			}
			spans = append(spans, lyricSpan{Text: ev.Lyric + " ", StartMS: start, EndMS: end})
			continue
		}

		spans = append(spans, lyricSpan{Text: ev.Lyric, StartMS: start, EndMS: end})
	}
	return spans
}

// isSectionMarker reconhece a convenção comum de cifras/letras de marcar trechos instrumentais ou
// seções da música entre colchetes (ex: "[INTRO]", "[SOLO]"), em vez de letra cantada de verdade.
func isSectionMarker(text string) bool {
	return strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]")
}

// wrapSpansIntoLines agrupa spans em linhas de exibição, nunca quebrando um span no meio, de modo
// que a largura de cada linha (segundo measure, a largura em pixels de um texto na fonte usada em
// tela) não ultrapasse maxWidth. measure é injetado em vez de medido diretamente com Ebiten pra
// essa função poder ser testada sem um contexto gráfico.
func wrapSpansIntoLines(spans []lyricSpan, maxWidth float64, measure func(string) float64) [][]lyricSpan {
	if len(spans) == 0 {
		return nil
	}

	var lines [][]lyricSpan
	var current []lyricSpan
	var currentText string

	flush := func() {
		if len(current) > 0 {
			lines = append(lines, current)
			current = nil
			currentText = ""
		}
	}

	for _, sp := range spans {
		candidate := currentText + sp.Text
		if len(current) > 0 && measure(candidate) > maxWidth {
			flush()
			candidate = sp.Text
		}
		current = append(current, sp)
		currentText = candidate
	}
	flush()

	return lines
}

// activeSpanIndex retorna o índice do último span cujo StartMS <= posMS, ou 0 se a música ainda
// não chegou no primeiro span (ou não houver spans). spans deve estar em ordem cronológica, o que
// buildLyricSpans já garante (segue a ordem da timeline).
func activeSpanIndex(spans []lyricSpan, posMS int64) int {
	idx := 0
	for i, sp := range spans {
		if sp.StartMS <= posMS {
			idx = i
		} else {
			break
		}
	}
	return idx
}

// lineIndexForSpan retorna em qual linha (índice em lines) o span de índice spanIdx (no fluxo
// original passado a wrapSpansIntoLines) caiu.
func lineIndexForSpan(lines [][]lyricSpan, spanIdx int) int {
	count := 0
	for i, line := range lines {
		count += len(line)
		if spanIdx < count {
			return i
		}
	}
	if len(lines) == 0 {
		return 0
	}
	return len(lines) - 1
}
