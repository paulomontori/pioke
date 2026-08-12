package parser

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"pioke/pkg/model"
)

// ParseMusicXML lê um arquivo MusicXML em texto puro (.musicxml/.xml partwise) e o converte
// para model.Song. Para o formato comprimido (.mxl), use ParseMXL.
//
// Escopo suportado (subconjunto pragmático, suficiente para uma "lead sheet" de karaokê ou
// uma partitura instrumental simples):
//   - Apenas <score-partwise> (não <score-timewise>).
//   - Usa apenas uma <part> — a que tiver nome sugestivo de voz/melodia (ex: "Voice", "Vocal",
//     "Melody"), ou a primeira, se nenhuma corresponder. Outras partes (ex: acompanhamento em
//     pauta própria) são ignoradas nesta versão.
//   - Dentro da parte escolhida, a primeira <voice> encontrada no documento vira a melodia
//     principal (letra/syllables); as demais vozes (ex: baixo de violão, mão esquerda de piano)
//     tocam junto como acompanhamento (TimelineEvent.Accompaniment) em vez de serem descartadas —
//     <backup>/<forward> mantêm o cursor de tempo correto entre elas.
//   - Notas marcadas com <chord/> (soando junto da nota anterior, acordes na própria melodia)
//     são ignoradas — sem harmonização polifônica na melodia.
//   - Cada <measure> vira um model.TimelineEvent; cada <note> da voz escolhida vira um
//     model.Syllable (offset/duração relativos ao início do compasso, mais o pitch).
//   - <tie type="stop"> estende a duração da sílaba anterior em vez de criar uma nova
//     (nota ligada = mesmo ataque, duração maior).
//   - <harmony> (cifra) é convertida para o formato curto usado por synth.GetChordFrequencies
//     (ex: "Am7", "G7", "Cmaj7") e vale a partir daquele ponto até a próxima mudança. Partituras
//     puramente instrumentais (sem cifra) tocam apenas a melodia, sem acompanhamento.
//   - O andamento (<sound tempo="…">) pode mudar ao longo da partitura; cada nota usa o
//     andamento vigente no momento em que aparece para calcular sua duração em milissegundos.
func ParseMusicXML(filePath string) (*model.Song, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo MusicXML: %w", err)
	}
	return parseMusicXMLBytes(data)
}

// ParseMXL lê um arquivo MusicXML comprimido (.mxl — um ZIP contendo META-INF/container.xml
// e o documento score-partwise) e o converte para model.Song, reaproveitando a mesma lógica
// de ParseMusicXML sobre o XML descompactado.
func ParseMXL(filePath string) (*model.Song, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo .mxl (zip): %w", err)
	}
	defer zr.Close()

	rootName := mxlRootFile(zr.File)
	if rootName == "" {
		return nil, fmt.Errorf(".mxl sem nenhum documento MusicXML reconhecível")
	}

	var root *zip.File
	for _, f := range zr.File {
		if f.Name == rootName {
			root = f
			break
		}
	}
	if root == nil {
		return nil, fmt.Errorf(".mxl: arquivo raiz declarado (%s) não encontrado dentro do zip", rootName)
	}

	rc, err := root.Open()
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir %s dentro do .mxl: %w", rootName, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler %s dentro do .mxl: %w", rootName, err)
	}

	return parseMusicXMLBytes(data)
}

// mxlRootFile determina qual entrada do .mxl é o documento MusicXML principal: lê o manifesto
// META-INF/container.xml quando presente (como definido pelo formato .mxl), e cai para a
// primeira entrada .xml/.musicxml fora de META-INF/ caso o manifesto esteja ausente ou inválido.
func mxlRootFile(files []*zip.File) string {
	for _, f := range files {
		if f.Name != "META-INF/container.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			break
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			break
		}
		var container mxlContainer
		if err := xml.Unmarshal(data, &container); err == nil {
			for _, rf := range container.RootFiles {
				if rf.FullPath != "" {
					return rf.FullPath
				}
			}
		}
		break
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name, "META-INF/") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".xml" || ext == ".musicxml" {
			return f.Name
		}
	}
	return ""
}

func parseMusicXMLBytes(data []byte) (*model.Song, error) {
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
	selectedVoice := "" // primeira <voice> encontrada no documento; demais vozes são ignoradas
	var cursorMS int64

	for _, measure := range part.Measures {
		measureStartMS := cursorMS
		var posMS int64    // posição dentro do compasso, afetada por <backup>/<forward>
		var maxPosMS int64 // maior posição alcançada por qualquer voz = duração real do compasso
		var syllables []model.Syllable
		var accompaniment []model.Syllable // notas de outras vozes, soando junto com a melodia
		var lyricLine strings.Builder
		measureChord := ""

		divisionsToMS := func(divisionUnits int64) int64 {
			effectiveBPM := bpm
			if effectiveBPM <= 0 {
				effectiveBPM = defaultBPM
			}
			quarterMS := 60000.0 / effectiveBPM
			return int64(math.Round(float64(divisionUnits) / float64(divisions) * quarterMS))
		}

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

			case item.Backup != nil:
				posMS -= divisionsToMS(item.Backup.Duration)
				if posMS < 0 {
					posMS = 0
				}

			case item.Forward != nil:
				posMS += divisionsToMS(item.Forward.Duration)
				if posMS > maxPosMS {
					maxPosMS = posMS
				}

			case item.Note != nil:
				n := item.Note
				if n.Chord != nil {
					// Nota simultânea à anterior — fora do escopo (melodia monofônica).
					continue
				}

				durMS := divisionsToMS(n.Duration)

				voice := n.Voice
				if voice == "" {
					voice = "1"
				}
				if selectedVoice == "" {
					selectedVoice = voice
				}
				if voice != selectedVoice {
					// Nota de outra voz (ex: baixo de um violão, ou mão esquerda de piano): não
					// entra na letra/melodia principal, mas soa ao mesmo tempo — vira acompanhamento
					// em vez de ser descartada, senão pausas na melodia (ex: dedilhado com respiro)
					// viram silêncio total mesmo com o baixo sustentando a nota.
					if n.Rest == nil {
						accompaniment = append(accompaniment, model.Syllable{
							OffsetMS:   posMS,
							DurationMS: durMS,
							Pitch:      pitchName(n.Pitch),
						})
					}
					posMS += durMS
					if posMS > maxPosMS {
						maxPosMS = posMS
					}
					continue
				}

				switch {
				case n.Rest != nil:
					posMS += durMS
				case hasTieStop(n.Tie) && len(syllables) > 0:
					// Nota ligada à anterior: estende a sílaba, sem novo ataque.
					syllables[len(syllables)-1].DurationMS += durMS
					posMS += durMS
				default:
					text := ""
					syllabic := ""
					if n.Lyric != nil {
						text = n.Lyric.Text
						syllabic = n.Lyric.Syllabic
					}
					syllables = append(syllables, model.Syllable{
						Text:       text,
						OffsetMS:   posMS,
						DurationMS: durMS,
						Pitch:      pitchName(n.Pitch),
					})
					if text != "" {
						// <syllabic> diz se esta sílaba continua a palavra anterior ("middle"/"end")
						// ou começa uma nova ("single"/"begin"/ausente) — sem isso, sílabas de
						// palavras diferentes ficariam coladas ("AndI'd" em vez de "And I'd").
						if lyricLine.Len() > 0 && syllabic != "middle" && syllabic != "end" {
							lyricLine.WriteByte(' ')
						}
						lyricLine.WriteString(text)
					}
					posMS += durMS
				}
				if posMS > maxPosMS {
					maxPosMS = posMS
				}
			}
		}

		cursorMS = measureStartMS + maxPosMS

		if len(syllables) == 0 && len(accompaniment) == 0 {
			continue // compasso sem nenhuma nota — não gera evento
		}
		if measureChord == "" {
			measureChord = currentChord
		}

		s.Timeline = append(s.Timeline, model.TimelineEvent{
			TimeMS:        measureStartMS,
			DurationMS:    maxPosMS,
			ChordStr:      measureChord,
			Lyric:         lyricLine.String(),
			Syllables:     syllables,
			Accompaniment: accompaniment,
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
	Backup     *mxBackup
	Forward    *mxForward
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
			case "backup":
				var b mxBackup
				if err := d.DecodeElement(&b, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, mxMeasureItem{Backup: &b})
			case "forward":
				var f mxForward
				if err := d.DecodeElement(&f, &t); err != nil {
					return err
				}
				m.Items = append(m.Items, mxMeasureItem{Forward: &f})
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
	Voice    string    `xml:"voice"`
	Tie      []mxTie   `xml:"tie"`
	Lyric    *mxLyric  `xml:"lyric"`
}

// mxBackup/mxForward implementam o "cursor de tempo" do MusicXML: <backup> rebobina a posição
// (usado para começar a próxima voz/pauta no mesmo instante), <forward> avança sem soar nota.
type mxBackup struct {
	Duration int64 `xml:"duration"`
}

type mxForward struct {
	Duration int64 `xml:"duration"`
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

// mxlContainer decodifica META-INF/container.xml, o manifesto que aponta para o documento
// MusicXML principal dentro de um arquivo .mxl.
type mxlContainer struct {
	RootFiles []mxlRootFileEntry `xml:"rootfiles>rootfile"`
}

type mxlRootFileEntry struct {
	FullPath string `xml:"full-path,attr"`
}
