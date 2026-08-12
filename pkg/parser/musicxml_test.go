package parser

import (
	"archive/zip"
	"os"
	"testing"

	"pioke/pkg/model"
)

const testParabensMusicXML = `<?xml version="1.0" encoding="UTF-8"?>
<score-partwise version="4.0">
  <work>
    <work-title>Parabéns a Você</work-title>
  </work>
  <identification>
    <creator type="lyricist">Domínio Público</creator>
  </identification>
  <part-list>
    <score-part id="P1">
      <part-name>Voice</part-name>
    </score-part>
  </part-list>
  <part id="P1">
    <measure number="1">
      <attributes>
        <divisions>2</divisions>
        <key><fifths>0</fifths></key>
        <time><beats>4</beats><beat-type>4</beat-type></time>
      </attributes>
      <direction>
        <sound tempo="100"/>
      </direction>
      <harmony>
        <root><root-step>C</root-step></root>
        <kind>major</kind>
      </harmony>
      <note>
        <pitch><step>G</step><octave>4</octave></pitch>
        <duration>1</duration>
        <lyric><syllabic>begin</syllabic><text>Pa</text></lyric>
      </note>
      <note>
        <pitch><step>G</step><octave>4</octave></pitch>
        <duration>1</duration>
        <lyric><syllabic>middle</syllabic><text>ra</text></lyric>
      </note>
      <note>
        <pitch><step>A</step><octave>4</octave></pitch>
        <duration>2</duration>
        <lyric><syllabic>end</syllabic><text>béns</text></lyric>
      </note>
      <note>
        <pitch><step>G</step><octave>4</octave></pitch>
        <duration>2</duration>
        <lyric><syllabic>single</syllabic><text>pra</text></lyric>
      </note>
      <note>
        <pitch><step>C</step><octave>5</octave></pitch>
        <duration>2</duration>
        <lyric><syllabic>begin</syllabic><text>vo</text></lyric>
      </note>
    </measure>
    <measure number="2">
      <harmony>
        <root><root-step>G</root-step></root>
        <kind>dominant</kind>
      </harmony>
      <note>
        <pitch><step>B</step><octave>4</octave></pitch>
        <duration>4</duration>
        <tie type="start"/>
        <lyric><syllabic>end</syllabic><text>cê</text></lyric>
      </note>
      <note>
        <pitch><step>B</step><octave>4</octave></pitch>
        <duration>4</duration>
        <tie type="stop"/>
      </note>
    </measure>
  </part>
</score-partwise>
`

func writeTempMusicXML(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "song_test_*.musicxml")
	if err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("erro ao escrever arquivo temporário: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	return tmpFile.Name()
}

func TestParseMusicXML(t *testing.T) {
	path := writeTempMusicXML(t, testParabensMusicXML)

	s, err := ParseMusicXML(path)
	if err != nil {
		t.Fatalf("ParseMusicXML falhou: %v", err)
	}

	if s.Title != "Parabéns a Você" {
		t.Errorf("Title esperado 'Parabéns a Você', obtido %q", s.Title)
	}
	if s.Artist != "Domínio Público" {
		t.Errorf("Artist esperado 'Domínio Público', obtido %q", s.Artist)
	}
	if s.BPM != 100 {
		t.Errorf("BPM esperado 100, obtido %d", s.BPM)
	}
	if s.Key != "C" {
		t.Errorf("Key esperado 'C', obtido %q", s.Key)
	}
	if s.TimeSig != "4/4" {
		t.Errorf("TimeSig esperado '4/4', obtido %q", s.TimeSig)
	}

	if len(s.Timeline) != 2 {
		t.Fatalf("Esperado 2 eventos (1 por compasso), obtido %d", len(s.Timeline))
	}

	ev0 := s.Timeline[0]
	if ev0.TimeMS != 0 || ev0.DurationMS != 2400 {
		t.Errorf("Evento 0: TimeMS/DurationMS esperados 0/2400, obtidos %d/%d", ev0.TimeMS, ev0.DurationMS)
	}
	if ev0.ChordStr != "C" {
		t.Errorf("Evento 0: ChordStr esperado 'C', obtido %q", ev0.ChordStr)
	}
	if ev0.Lyric != "Parabéns pra vo" {
		t.Errorf("Evento 0: Lyric esperado 'Parabéns pra vo', obtido %q", ev0.Lyric)
	}

	wantSyllables := []model.Syllable{
		{Text: "Pa", OffsetMS: 0, DurationMS: 300, Pitch: "G4"},
		{Text: "ra", OffsetMS: 300, DurationMS: 300, Pitch: "G4"},
		{Text: "béns", OffsetMS: 600, DurationMS: 600, Pitch: "A4"},
		{Text: "pra", OffsetMS: 1200, DurationMS: 600, Pitch: "G4"},
		{Text: "vo", OffsetMS: 1800, DurationMS: 600, Pitch: "C5"},
	}
	assertSyllables(t, "Evento 0", ev0.Syllables, wantSyllables)

	ev1 := s.Timeline[1]
	if ev1.TimeMS != 2400 || ev1.DurationMS != 2400 {
		t.Errorf("Evento 1: TimeMS/DurationMS esperados 2400/2400, obtidos %d/%d", ev1.TimeMS, ev1.DurationMS)
	}
	if ev1.ChordStr != "G7" {
		t.Errorf("Evento 1: ChordStr esperado 'G7' (dominant -> 7), obtido %q", ev1.ChordStr)
	}
	// A nota ligada (tie stop) deve estender a sílaba "cê" em vez de criar uma segunda sílaba.
	assertSyllables(t, "Evento 1", ev1.Syllables, []model.Syllable{
		{Text: "cê", OffsetMS: 0, DurationMS: 2400, Pitch: "B4"},
	})
}

func assertSyllables(t *testing.T, label string, got, want []model.Syllable) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: esperado %d sílabas, obtido %d (%+v)", label, len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: sílaba %d esperada %+v, obtida %+v", label, i, want[i], got[i])
		}
	}
}

func TestPitchName(t *testing.T) {
	tests := []struct {
		pitch *mxPitch
		want  string
	}{
		{&mxPitch{Step: "G", Octave: 4}, "G4"},
		{&mxPitch{Step: "C", Alter: 1, Octave: 5}, "C#5"},
		{&mxPitch{Step: "B", Alter: -1, Octave: 3}, "Bb3"},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := pitchName(tt.pitch); got != tt.want {
			t.Errorf("pitchName(%+v) = %q, esperado %q", tt.pitch, got, tt.want)
		}
	}
}

func TestHarmonyToChordString(t *testing.T) {
	tests := []struct {
		h    mxHarmony
		want string
	}{
		{mxHarmony{Root: mxRoot{RootStep: "C"}, Kind: mxKind{Value: "major"}}, "C"},
		{mxHarmony{Root: mxRoot{RootStep: "A"}, Kind: mxKind{Value: "minor"}}, "Am"},
		{mxHarmony{Root: mxRoot{RootStep: "G"}, Kind: mxKind{Value: "dominant"}}, "G7"},
		{mxHarmony{Root: mxRoot{RootStep: "F", RootAlter: 1}, Kind: mxKind{Value: "major-seventh"}}, "F#maj7"},
		{mxHarmony{Root: mxRoot{RootStep: "B", RootAlter: -1}, Kind: mxKind{Value: "minor-seventh"}}, "Bbm7"},
	}
	for _, tt := range tests {
		if got := harmonyToChordString(tt.h); got != tt.want {
			t.Errorf("harmonyToChordString(%+v) = %q, esperado %q", tt.h, got, tt.want)
		}
	}
}

func TestFifthsToKeyName(t *testing.T) {
	if got := fifthsToKeyName(0, ""); got != "C" {
		t.Errorf("fifthsToKeyName(0, \"\") = %q, esperado \"C\"", got)
	}
	if got := fifthsToKeyName(2, ""); got != "D" {
		t.Errorf("fifthsToKeyName(2, \"\") = %q, esperado \"D\"", got)
	}
	if got := fifthsToKeyName(0, "minor"); got != "Am" {
		t.Errorf("fifthsToKeyName(0, \"minor\") = %q, esperado \"Am\"", got)
	}
}

// testTwoVoicePianoMusicXML representa um compasso de piano com duas vozes simultâneas
// (mão direita = voice 1, mão esquerda = voice 2, como em Für Elise): a voz 1 tem duas
// mínimas (C5, D5); depois de um <backup> que rebobina o compasso inteiro, a voz 2 toca uma
// semibreve (C3) sozinha. Só a voz 1 (a primeira encontrada no documento) deve virar melodia.
const testTwoVoicePianoMusicXML = `<?xml version="1.0" encoding="UTF-8"?>
<score-partwise version="4.0">
  <work><work-title>Duas Vozes</work-title></work>
  <part-list><score-part id="P1"><part-name>Piano</part-name></score-part></part-list>
  <part id="P1">
    <measure number="1">
      <attributes>
        <divisions>1</divisions>
        <time><beats>4</beats><beat-type>4</beat-type></time>
      </attributes>
      <direction><sound tempo="60"/></direction>
      <note>
        <pitch><step>C</step><octave>5</octave></pitch>
        <duration>2</duration>
        <voice>1</voice>
      </note>
      <note>
        <pitch><step>D</step><octave>5</octave></pitch>
        <duration>2</duration>
        <voice>1</voice>
      </note>
      <backup><duration>4</duration></backup>
      <note>
        <pitch><step>C</step><octave>3</octave></pitch>
        <duration>4</duration>
        <voice>2</voice>
      </note>
    </measure>
  </part>
</score-partwise>
`

func TestParseMusicXML_MultiVoiceOnlyFirstVoiceKept(t *testing.T) {
	path := writeTempMusicXML(t, testTwoVoicePianoMusicXML)

	s, err := ParseMusicXML(path)
	if err != nil {
		t.Fatalf("ParseMusicXML falhou: %v", err)
	}
	if len(s.Timeline) != 1 {
		t.Fatalf("Esperado 1 evento (1 compasso), obtido %d", len(s.Timeline))
	}

	ev := s.Timeline[0]
	if ev.TimeMS != 0 || ev.DurationMS != 4000 {
		t.Errorf("Evento: TimeMS/DurationMS esperados 0/4000 (largura total do compasso), obtidos %d/%d", ev.TimeMS, ev.DurationMS)
	}
	// Apenas as 2 notas da voice 1 (mão direita) devem aparecer — a nota da voice 2 (mão
	// esquerda, depois do <backup>) precisa ser ignorada, não concatenada como se fosse melodia.
	assertSyllables(t, "Evento", ev.Syllables, []model.Syllable{
		{Text: "", OffsetMS: 0, DurationMS: 2000, Pitch: "C5"},
		{Text: "", OffsetMS: 2000, DurationMS: 2000, Pitch: "D5"},
	})
}

func writeTempMXL(t *testing.T, entries map[string]string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "song_test_*.mxl")
	if err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	zw := zip.NewWriter(tmpFile)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("erro ao criar entrada %s no zip: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("erro ao escrever entrada %s no zip: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("erro ao fechar zip: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

const testContainerXML = `<?xml version="1.0" encoding="UTF-8"?>
<container>
  <rootfiles>
    <rootfile full-path="score.xml"/>
  </rootfiles>
</container>
`

func TestParseMXL_WithContainerManifest(t *testing.T) {
	path := writeTempMXL(t, map[string]string{
		"META-INF/container.xml": testContainerXML,
		"score.xml":              testParabensMusicXML,
	})

	s, err := ParseMXL(path)
	if err != nil {
		t.Fatalf("ParseMXL falhou: %v", err)
	}
	if s.Title != "Parabéns a Você" {
		t.Errorf("Title esperado 'Parabéns a Você', obtido %q", s.Title)
	}
	if len(s.Timeline) != 2 {
		t.Errorf("Esperado 2 eventos, obtido %d", len(s.Timeline))
	}
}

func TestParseMXL_WithoutManifestFallsBackToFirstXML(t *testing.T) {
	path := writeTempMXL(t, map[string]string{
		"minha_partitura.xml": testParabensMusicXML,
	})

	s, err := ParseMXL(path)
	if err != nil {
		t.Fatalf("ParseMXL falhou: %v", err)
	}
	if s.Title != "Parabéns a Você" {
		t.Errorf("Title esperado 'Parabéns a Você', obtido %q", s.Title)
	}
}

func TestParseMXL_NotAZip(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "not_a_zip_*.mxl")
	if err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	tmpFile.WriteString("isto não é um zip")
	tmpFile.Close()

	if _, err := ParseMXL(tmpFile.Name()); err == nil {
		t.Error("esperava erro ao tentar abrir um .mxl que não é um zip válido")
	}
}

func TestSelectMelodyPart(t *testing.T) {
	doc := mxScorePartwise{
		PartList: mxPartList{ScoreParts: []mxScorePart{
			{ID: "P1", PartName: "Piano"},
			{ID: "P2", PartName: "Voice"},
		}},
		Parts: []mxPart{{ID: "P1"}, {ID: "P2"}},
	}
	got := selectMelodyPart(doc)
	if got.ID != "P2" {
		t.Errorf("selectMelodyPart deveria escolher a parte 'Voice' (P2), obteve %q", got.ID)
	}
}
