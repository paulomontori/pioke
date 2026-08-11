package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strings"

	"pioke/pkg/model"
)

// ParseMusicXML lê um arquivo MusicXML (formato .musicxml/.xml partwise) e o converte para
// model.Song, sincronizando letra, altura (pitch) e cifras a partir da partitura.
//
// Escopo suportado (subconjunto pragmático, suficiente para uma "lead sheet" de karaokê):
//   - Apenas <score-partwise> (não <score-timewise>).
//   - Usa apenas uma <part> — a que tiver nome sugestivo de voz/melodia (ex: "Voice", "Vocal",
//     "Melody"), ou a primeira, se nenhuma corresponder. Outras partes (ex: acompanhamento em
//     pauta própria) são ignoradas nesta versão.
//   - Não trata múltiplas vozes simultâneas na mesma parte (<backup>/<forward>): assume notas
//     sequenciais em ordem de leitura, como em uma linha melódica cantada.
//   - Notas marcadas com <chord/> (soando junto da nota anterior) são ignoradas — sem
//     harmonização polifônica na própria melodia.
//   - Cada <measure> vira um model.TimelineEvent; cada <note> dentro dele vira um
//     model.Syllable (offset/duração relativos ao início do compasso, mais o pitch).
//   - <tie type="stop"> estende a duração da sílaba anterior em vez de criar uma nova
//     (nota ligada = mesmo ataque, duração maior).
//   - <harmony> (cifra) é convertida para o formato curto usado por synth.GetChordFrequencies
//     (ex: "Am7", "G7", "Cmaj7") e vale a partir daquele ponto até a próxima mudança.
//   - O andamento (<sound tempo="…">) pode mudar ao longo da partitura; cada nota usa o
//     andamento vigente no momento em que aparece para calcular sua duração em milissegundos.
func ParseMusicXML(filePath string) (*model.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo MusicXML: %w", err)
	}

	var doc mxScorePartwise
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("erro ao decodificar MusicXML: %w", err)
	}
	if len(doc.Parts) == 0 {
		return nil, fmt.Errorf("MusicXML sem nenhuma <part>")
	}

	title := firstNonEmpty(doc.Work.WorkTitle, doc.MovementTitle)
	if title == "" {
		return nil, fmt.Errorf("MusicXML sem título (<work-title> ou <movement-title>)")
	}

	s := &model.Song{
		Title:  title,
		Artist: doc.Identification.creatorName(),
	}

	part := selectMelodyPart(doc)

	const defaultBPM = 90.0
	bpm := 0.0
	divisions := 1
	currentChord := ""
	var cursorMS int64

	for _, measure := range part.Measures {
		measureStartMS := cursorMS
		var syllables []model.Syllable
		measureChord := ""

		for _, item := range measure.Items {
			switch {
			case item.Attributes != nil:
				if item.Attributes.Divisions > 0 {
					divisions = item.Attributes.Divisions
				}
				if item.Attributes.Key != nil && s.Key == "" {
					s.Key = fifthsToKeyName(item.Attributes.Key.Fifths, item.Attributes.Key.Mode)
				}
				if item.Attributes.Time != nil && s.TimeSig == "" {
					s.TimeSig = fmt.Sprintf("%s/%s", item.Attributes.Time.Beats, item.Attributes.Time.BeatType)
				}

			case item.Sound != nil:
				if item.Sound.Tempo > 0 {
					bpm = item.Sound.Tempo
				}

			case item.Harmony != nil:
				currentChord = harmonyToChordString(*item.Harmony)
				if measureChord == "" {
					measureChord = currentChord
				}

			case item.Note != nil:
				n := item.Note
				effectiveBPM := bpm
				if effectiveBPM <= 0 {
					effectiveBPM = defaultBPM
				}
				quarterMS := 60000.0 / effectiveBPM
				durMS := int64(math.Round(float64(n.Duration) / float64(divisions) * quarterMS))

				switch {
				case n.Chord != nil:
					// Nota simultânea à anterior — fora do escopo (melodia monofônica).
					continue
				case n.Rest != nil:
					cursorMS += durMS
					continue
				case hasTieStop(n.Tie) && len(syllables) > 0:
					// Nota ligada à anterior: estende a sílaba, sem novo ataque.
					syllables[len(syllables)-1].DurationMS += durMS
					cursorMS += durMS
					continue
				}

				text := ""
				if n.Lyric != nil {
					text = n.Lyric.Text
				}
				syllables = append(syllables, model.Syllable{
					Text:       text,
					OffsetMS:   cursorMS - measureStartMS,
					DurationMS: durMS,
					Pitch:      pitchName(n.Pitch),
				})
				cursorMS += durMS
			}
		}

		if len(syllables) == 0 {
			continue // compasso só com atributos/silêncio, sem notas — não gera evento
		}
		if measureChord == "" {
			measureChord = currentChord
		}

		var lyricLine strings.Builder
		for _, syl := range syllables {
			lyricLine.WriteString(syl.Text)
		}

		s.Timeline = append(s.Timeline, model.TimelineEvent{
			TimeMS:     measureStartMS,
			DurationMS: cursorMS - measureStartMS,
			ChordStr:   measureChord,
			Lyric:      lyricLine.String(),
			Syllables:  syllables,
		})
	}

	if bpm > 0 {
		s.BPM = int(math.Round(bpm))
	} else {
		s.BPM = int(defaultBPM)
	}

	return s, nil
}

// selectMelodyPart escolhe a <part> mais provável de conter a melodia cantada: procura por um
// nome sugestivo ("voice", "vocal", "melody"/"melodia", "lead") em <part-name>, e cai para a
// primeira parte do arquivo se nenhuma corresponder.
var melodyPartNameRe = regexp.MustCompile(`(?i)voc|voice|melod|lead`)

func selectMelodyPart(doc mxScorePartwise) mxPart {
	names := make(map[string]string, len(doc.PartList.ScoreParts))
	for _, sp := range doc.PartList.ScoreParts {
		names[sp.ID] = sp.PartName
	}
	for _, part := range doc.Parts {
		if melodyPartNameRe.MatchString(names[part.ID]) {
			return part
		}
	}
	return doc.Parts[0]
}

func hasTieStop(ties []mxTie) bool {
	for _, t := range ties {
		if t.Type == "stop" {
			return true
		}
	}
	return false
}

// pitchName converte <pitch> (step + alter + octave) para a notação científica usada em
// synth.NoteNameToFrequency (ex: "G4", "C#5", "Bb3").
func pitchName(p *mxPitch) string {
	if p == nil {
		return ""
	}
	name := p.Step
	switch {
	case p.Alter > 0:
		name += strings.Repeat("#", p.Alter)
	case p.Alter < 0:
		name += strings.Repeat("b", -p.Alter)
	}
	return fmt.Sprintf("%s%d", name, p.Octave)
}

// mxKindSuffix mapeia o <kind> padronizado do MusicXML para o sufixo curto de cifra
// reconhecido por synth.GetChordFrequencies. "kind" desconhecido cai para maior (sem sufixo).
var mxKindSuffix = map[string]string{
	"major":               "",
	"minor":               "m",
	"dominant":            "7",
	"minor-seventh":       "m7",
	"major-seventh":       "maj7",
	"diminished":          "dim",
	"suspended-fourth":    "sus4",
	"suspended-second":    "sus2",
	"augmented":           "aug",
	"half-diminished":     "m7b5",
	"minor-major-seventh": "m(maj7)",
}

// harmonyToChordString converte um <harmony> (root + kind) para uma cifra curta, ex: "Am7".
func harmonyToChordString(h mxHarmony) string {
	root := h.Root.RootStep
	switch {
	case h.Root.RootAlter > 0:
		root += "#"
	case h.Root.RootAlter < 0:
		root += "b"
	}
	return root + mxKindSuffix[strings.TrimSpace(h.Kind.Value)]
}

// fifthsToKeyName aproxima o nome do tom a partir do círculo de quintas (<fifths>); usa a
// armadura maior salvo quando <mode> indica "minor" explicitamente.
var fifthsMajorKeys = map[int]string{
	-7: "Cb", -6: "Gb", -5: "Db", -4: "Ab", -3: "Eb", -2: "Bb", -1: "F",
	0: "C", 1: "G", 2: "D", 3: "A", 4: "E", 5: "B", 6: "F#", 7: "C#",
}
var fifthsMinorKeys = map[int]string{
	-7: "Ab", -6: "Eb", -5: "Bb", -4: "F", -3: "C", -2: "G", -1: "D",
	0: "A", 1: "E", 2: "B", 3: "F#", 4: "C#", 5: "G#", 6: "D#", 7: "A#",
}

func fifthsToKeyName(fifths int, mode string) string {
	table := fifthsMajorKeys
	suffix := ""
	if strings.EqualFold(mode, "minor") {
		table = fifthsMinorKeys
		suffix = "m"
	}
	if name, ok := table[fifths]; ok {
		return name + suffix
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// --- Estruturas de decodificação do subconjunto de MusicXML usado acima ---

type mxScorePartwise struct {
	XMLName        xml.Name         `xml:"score-partwise"`
	Work           mxWork           `xml:"work"`
	MovementTitle  string           `xml:"movement-title"`
	Identification mxIdentification `xml:"identification"`
	PartList       mxPartList       `xml:"part-list"`
	Parts          []mxPart         `xml:"part"`
}

type mxWork struct {
	WorkTitle string `xml:"work-title"`
}

type mxIdentification struct {
	Creators []mxCreator `xml:"creator"`
}

type mxCreator struct {
	Type string `xml:"type,attr"`
	Name string `xml:",chardata"`
}

// creatorName escolhe o(a) compositor(a)/artista mais relevante para exibição; cai para o
// primeiro <creator> encontrado se nenhum tipo reconhecido existir.
func (id mxIdentification) creatorName() string {
	for _, c := range id.Creators {
		if c.Type == "composer" || c.Type == "lyricist" || c.Type == "arranger" {
			if name := strings.TrimSpace(c.Name); name != "" {
				return name
			}
		}
	}
	if len(id.Creators) > 0 {
		return strings.TrimSpace(id.Creators[0].Name)
	}
	return ""
}

type mxPartList struct {
	ScoreParts []mxScorePart `xml:"score-part"`
}

type mxScorePart struct {
	ID       string `xml:"id,attr"`
	PartName string `xml:"part-name"`
}

type mxPart struct {
	ID       string      `xml:"id,attr"`
	Measures []mxMeasure `xml:"measure"`
}

// mxMeasureItem representa um filho de <measure> (nota, cifra, atributos ou andamento),
// preservando a ordem original do documento — essencial para saber qual cifra/andamento
// valia no instante de cada nota. Exatamente um dos campos é não-nulo.
type mxMeasureItem struct {
	Note       *mxNote
	Harmony    *mxHarmony
	Attributes *mxAttributes
	Sound      *mxSound
}

type mxMeasure struct {
	Items []mxMeasureItem
}

// UnmarshalXML decodifica <measure> token a token (em vez de usar campos struct com tag por
// nome) porque o encoding/xml padrão perderia a ordem relativa entre elementos de nomes
// diferentes (<attributes>, <harmony>, <note>, <direction>) — e essa ordem é o que determina,
// por exemplo, se uma cifra vale a partir da nota seguinte ou da anterior.
func (m *mxMeasure) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "note":
				var n mxNote
				if err := d.DecodeElement(&n, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, mxMeasureItem{Note: &n})
			case "harmony":
				var h mxHarmony
				if err := d.DecodeElement(&h, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, mxMeasureItem{Harmony: &h})
			case "attributes":
				var a mxAttributes
				if err := d.DecodeElement(&a, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, mxMeasureItem{Attributes: &a})
			case "direction":
				var dir mxDirection
				if err := d.DecodeElement(&dir, &t); err != nil {
					return err
				}
				if dir.Sound != nil {
					m.Items = append(m.Items, mxMeasureItem{Sound: dir.Sound})
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}

type mxAttributes struct {
	Divisions int        `xml:"divisions"`
	Key       *mxKeySig  `xml:"key"`
	Time      *mxTimeSig `xml:"time"`
}

type mxKeySig struct {
	Fifths int    `xml:"fifths"`
	Mode   string `xml:"mode"`
}

type mxTimeSig struct {
	Beats    string `xml:"beats"`
	BeatType string `xml:"beat-type"`
}

type mxDirection struct {
	Sound *mxSound `xml:"sound"`
}

type mxSound struct {
	Tempo float64 `xml:"tempo,attr"`
}

type mxHarmony struct {
	Root mxRoot `xml:"root"`
	Kind mxKind `xml:"kind"`
}

type mxRoot struct {
	RootStep  string `xml:"root-step"`
	RootAlter int    `xml:"root-alter"`
}

type mxKind struct {
	Value string `xml:",chardata"`
}

type mxNote struct {
	Rest     *struct{} `xml:"rest"`
	Chord    *struct{} `xml:"chord"`
	Pitch    *mxPitch  `xml:"pitch"`
	Duration int64     `xml:"duration"`
	Tie      []mxTie   `xml:"tie"`
	Lyric    *mxLyric  `xml:"lyric"`
}

type mxPitch struct {
	Step   string `xml:"step"`
	Alter  int    `xml:"alter"`
	Octave int    `xml:"octave"`
}

type mxTie struct {
	Type string `xml:"type,attr"`
}

type mxLyric struct {
	Syllabic string `xml:"syllabic"`
	Text     string `xml:"text"`
}
