package gui

import (
	"reflect"
	"testing"

	"pioke/pkg/model"
)

func TestBuildLyricSpansExpandsSyllables(t *testing.T) {
	s := &model.Song{
		Timeline: []model.TimelineEvent{
			{
				TimeMS: 1000,
				Lyric:  "Parabéns pra você",
				Syllables: []model.Syllable{
					{Text: "Pa", OffsetMS: 0, DurationMS: 300},
					{Text: "ra", OffsetMS: 300, DurationMS: 300},
					{Text: "béns ", OffsetMS: 600, DurationMS: 600},
				},
			},
		},
	}

	got := buildLyricSpans(s)
	want := []lyricSpan{
		{Text: "Pa", StartMS: 1000, EndMS: 1300},
		{Text: "ra", StartMS: 1300, EndMS: 1600},
		{Text: "béns ", StartMS: 1600, EndMS: 2200},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// buildLyricSpans nunca inventa espaço nenhum entre eventos — confia só no que já vem embutido no
// texto (ver comentário em buildLyricSpans). Arquivos vindos de MusicXML cortam sílabas em
// compassos, que às vezes terminam no meio de uma palavra (aí não deve entrar espaço nenhum) e às
// vezes no fim de uma frase (aí o próprio parser, pkg/parser/musicxml.go, já embute o espaço no
// texto da sílaba) — só ele tem a informação (<syllabic>) pra saber qual dos dois é o caso.
func TestBuildLyricSpansNeverInsertsSpaceAtEventBoundary(t *testing.T) {
	s := &model.Song{
		Timeline: []model.TimelineEvent{
			{
				TimeMS: 0,
				Syllables: []model.Syllable{
					{Text: "vo", OffsetMS: 0, DurationMS: 600},
					{Text: "cê", OffsetMS: 600, DurationMS: 1200}, // sem espaço: continua no próximo evento
				},
			},
			{
				TimeMS: 3600,
				Syllables: []model.Syllable{
					{Text: "nes", OffsetMS: 0, DurationMS: 300}, // meio de palavra: "vocênes", sem espaço
				},
			},
		},
	}

	got := buildLyricSpans(s)
	if full := spansText(got); full != "vocênes" {
		t.Fatalf("esperava concatenação sem espaço ('vocênes'), obtido %q", full)
	}
}

// Quando o texto da sílaba já traz o espaço embutido (como o parser de MusicXML agora garante pra
// início de frase, e como os arquivos YAML/JSON hand-authored já faziam), esse espaço deve
// aparecer normalmente na concatenação.
func TestBuildLyricSpansRespectsEmbeddedSpace(t *testing.T) {
	s := &model.Song{
		Timeline: []model.TimelineEvent{
			{
				TimeMS: 0,
				Syllables: []model.Syllable{
					{Text: "vo", OffsetMS: 0, DurationMS: 600},
					{Text: "cê ", OffsetMS: 600, DurationMS: 1200}, // espaço já embutido: fim de frase
				},
			},
			{
				TimeMS: 3600,
				Syllables: []model.Syllable{
					{Text: "Nes", OffsetMS: 0, DurationMS: 300},
				},
			},
		},
	}

	got := buildLyricSpans(s)
	if full := spansText(got); full != "você Nes" {
		t.Fatalf("esperava espaço embutido preservado ('você Nes'), obtido %q", full)
	}
}

func TestBuildLyricSpansMergesRepeatedSectionMarkers(t *testing.T) {
	s := &model.Song{
		Timeline: []model.TimelineEvent{
			{TimeMS: 0, DurationMS: 297, Lyric: "[INTRO]"},
			{TimeMS: 297, DurationMS: 297, Lyric: "[INTRO]"},
			{TimeMS: 594, DurationMS: 297, Lyric: "[INTRO]"},
			{TimeMS: 891, DurationMS: 297, Lyric: "Quan"},
			{TimeMS: 1188, Lyric: "do "},
		},
	}

	got := buildLyricSpans(s)

	full := spansText(got)
	if full != "[INTRO] Quando " {
		t.Fatalf("esperava marcadores repetidos mesclados num só, separado da letra ('[INTRO] Quando '), obtido %q", full)
	}
	if len(got) != 3 {
		t.Fatalf("esperava 3 spans ([INTRO] mesclado + Quan + do ), obteve %d: %+v", len(got), got)
	}
	if got[0].StartMS != 0 || got[0].EndMS != 891 {
		t.Errorf("esperava o marcador mesclado cobrindo 0..891ms (do primeiro ao último evento repetido), obteve %+v", got[0])
	}
}

func TestBuildLyricSpansFallsBackToWholeEvent(t *testing.T) {
	s := &model.Song{
		Timeline: []model.TimelineEvent{
			{TimeMS: 297, DurationMS: 297, Lyric: "Quan"},
			{TimeMS: 594, Lyric: "sem duracao"}, // DurationMS <= 0 -> fallback
			{TimeMS: 900, Lyric: ""},             // sem letra -> ignorado
		},
	}

	got := buildLyricSpans(s)
	want := []lyricSpan{
		{Text: "Quan", StartMS: 297, EndMS: 594},
		{Text: "sem duracao", StartMS: 594, EndMS: 594 + defaultEventDurationMS},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestBuildLyricSpansNilSong(t *testing.T) {
	if got := buildLyricSpans(nil); got != nil {
		t.Fatalf("esperado nil, obtido %+v", got)
	}
}

// widthMeasure é um medidor determinístico pros testes: cada caractere "pesa" 1 unidade.
func widthMeasure(s string) float64 {
	return float64(len([]rune(s)))
}

func TestWrapSpansIntoLinesRespectsMaxWidthWithoutSplittingSpans(t *testing.T) {
	spans := []lyricSpan{
		{Text: "Pa", StartMS: 0, EndMS: 300},
		{Text: "ra", StartMS: 300, EndMS: 600},
		{Text: "béns ", StartMS: 600, EndMS: 1200},
		{Text: "pra ", StartMS: 1200, EndMS: 1800},
		{Text: "vo", StartMS: 1800, EndMS: 2400},
		{Text: "cê", StartMS: 2400, EndMS: 3600},
	}

	// "Pa"+"ra"+"béns " = 9 caracteres, já ultrapassa 8 ao somar "pra " -> deve quebrar antes.
	lines := wrapSpansIntoLines(spans, 8, widthMeasure)

	if len(lines) == 0 {
		t.Fatal("esperava ao menos uma linha")
	}

	var totalSpans int
	for _, line := range lines {
		totalSpans += len(line)
		w := widthMeasure(spansText(line))
		if w > 8 && len(line) > 1 {
			t.Errorf("linha %+v excede maxWidth (%v > 8) com mais de um span", line, w)
		}
	}
	if totalSpans != len(spans) {
		t.Fatalf("esperava %d spans no total, obteve %d", len(spans), totalSpans)
	}
}

func TestWrapSpansIntoLinesNeverSplitsASingleOversizedSpan(t *testing.T) {
	spans := []lyricSpan{{Text: "supercalifragilisticexpialidocious", StartMS: 0, EndMS: 1000}}
	lines := wrapSpansIntoLines(spans, 5, widthMeasure)

	if len(lines) != 1 || len(lines[0]) != 1 || lines[0][0].Text != spans[0].Text {
		t.Fatalf("span único não deveria ser quebrado, obteve %+v", lines)
	}
}

func TestWrapSpansIntoLinesEmpty(t *testing.T) {
	if lines := wrapSpansIntoLines(nil, 100, widthMeasure); lines != nil {
		t.Fatalf("esperado nil para spans vazio, obtido %+v", lines)
	}
}

func TestActiveSpanIndex(t *testing.T) {
	spans := []lyricSpan{
		{Text: "a", StartMS: 100, EndMS: 200},
		{Text: "b", StartMS: 200, EndMS: 300},
		{Text: "c", StartMS: 300, EndMS: 400},
	}

	cases := []struct {
		posMS int64
		want  int
	}{
		{posMS: 0, want: 0},   // antes do início
		{posMS: 100, want: 0}, // exatamente no início do primeiro
		{posMS: 250, want: 1}, // no meio
		{posMS: 999, want: 2}, // depois do fim
	}
	for _, c := range cases {
		if got := activeSpanIndex(spans, c.posMS); got != c.want {
			t.Errorf("activeSpanIndex(%d) = %d, want %d", c.posMS, got, c.want)
		}
	}
}

func TestActiveSpanIndexEmpty(t *testing.T) {
	if got := activeSpanIndex(nil, 500); got != 0 {
		t.Errorf("esperado 0 com spans vazio, obtido %d", got)
	}
}

func TestLineIndexForSpan(t *testing.T) {
	lines := [][]lyricSpan{
		{{Text: "a"}, {Text: "b"}}, // spans globais 0,1
		{{Text: "c"}},              // span global 2
		{{Text: "d"}, {Text: "e"}}, // spans globais 3,4
	}

	cases := map[int]int{0: 0, 1: 0, 2: 1, 3: 2, 4: 2}
	for spanIdx, want := range cases {
		if got := lineIndexForSpan(lines, spanIdx); got != want {
			t.Errorf("lineIndexForSpan(%d) = %d, want %d", spanIdx, got, want)
		}
	}
}

func TestLineIndexForSpanEmptyLines(t *testing.T) {
	if got := lineIndexForSpan(nil, 0); got != 0 {
		t.Errorf("esperado 0 com lines vazio, obtido %d", got)
	}
}
