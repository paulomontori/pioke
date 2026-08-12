package scoring

import "fmt"

// OctaveErrorMode controla como um erro de oitava (cantar a mesma nota uma ou mais oitavas acima
// ou abaixo do alvo) é tratado na pontuação de afinação de uma nota.
type OctaveErrorMode int

const (
	// OctaveIgnore corrige o erro de oitava antes de pontuar e não aplica nenhuma penalidade —
	// cantar na oitava errada conta como acerto.
	OctaveIgnore OctaveErrorMode = iota
	// OctavePenalize corrige o erro de oitava antes de pontuar, mas aplica OctavePenaltyFactor à
	// pontuação da nota quando a correção foi necessária.
	OctavePenalize
	// OctaveStrict não corrige oitava: a nota é pontuada com o desvio em cents bruto, então
	// cantar na oitava errada normalmente zera a pontuação daquela nota.
	OctaveStrict
)

// ScoringPreset agrupa os parâmetros de um nível de dificuldade — nenhuma regra de pontuação é
// fixa no código; tudo que varia por nível vive aqui, e a função de score final recebe o preset
// escolhido como parâmetro.
type ScoringPreset struct {
	Name string

	// CentsFullScoreAbs: desvio de afinação (valor absoluto, em cents) até o qual a nota recebe
	// pontuação de afinação máxima (100).
	CentsFullScoreAbs float64
	// CentsZeroScoreAbs: desvio de afinação (valor absoluto, em cents) a partir do qual a nota
	// recebe pontuação de afinação zero. Entre CentsFullScoreAbs e CentsZeroScoreAbs a pontuação
	// cai linearmente.
	CentsZeroScoreAbs float64

	// TimeToleranceMS: desvio do onset detectado em relação ao start_ms da nota de referência a
	// partir do qual a pontuação de ritmo daquela nota chega a zero (cai linearmente até lá).
	TimeToleranceMS int64

	OctaveErrorMode OctaveErrorMode
	// OctavePenaltyFactor multiplica a pontuação de afinação da nota quando uma correção de
	// oitava foi aplicada (só usado com OctavePenalize).
	OctavePenaltyFactor float64

	// Pesos (0-100, somam 100) de cada componente na pontuação final da música.
	WeightPitch    float64
	WeightRhythm   float64
	WeightCoverage float64
}

// PresetEasy, PresetMedium e PresetHard implementam a tabela de dificuldade combinada:
//
//	                          easy         medium        hard
//	Cents p/ score máximo     ±50          ±25           ±10
//	Cents p/ score zero       ±150         ±100          ±50
//	Tolerância de tempo       ±350ms       ±180ms        ±80ms
//	Erro de oitava            ignorado     ignorado,     não ignorado
//	                                       penalidade
//	                                       parcial (50%)
//	Peso afinação/ritmo/cob.  60/20/20     70/20/10      80/15/5
var (
	PresetEasy = ScoringPreset{
		Name:                "easy",
		CentsFullScoreAbs:   50,
		CentsZeroScoreAbs:   150,
		TimeToleranceMS:     350,
		OctaveErrorMode:     OctaveIgnore,
		OctavePenaltyFactor: 1.0,
		WeightPitch:         60,
		WeightRhythm:        20,
		WeightCoverage:      20,
	}

	PresetMedium = ScoringPreset{
		Name:                "medium",
		CentsFullScoreAbs:   25,
		CentsZeroScoreAbs:   100,
		TimeToleranceMS:     180,
		OctaveErrorMode:     OctavePenalize,
		OctavePenaltyFactor: 0.5,
		WeightPitch:         70,
		WeightRhythm:        20,
		WeightCoverage:      10,
	}

	PresetHard = ScoringPreset{
		Name:                "hard",
		CentsFullScoreAbs:   10,
		CentsZeroScoreAbs:   50,
		TimeToleranceMS:     80,
		OctaveErrorMode:     OctaveStrict,
		OctavePenaltyFactor: 1.0,
		WeightPitch:         80,
		WeightRhythm:        15,
		WeightCoverage:      5,
	}
)

// PresetByName resolve "easy", "medium" ou "hard" (case-insensitive) para o ScoringPreset
// correspondente.
func PresetByName(name string) (ScoringPreset, error) {
	switch name {
	case "easy":
		return PresetEasy, nil
	case "medium":
		return PresetMedium, nil
	case "hard":
		return PresetHard, nil
	default:
		return ScoringPreset{}, fmt.Errorf("nível de dificuldade desconhecido: %q (use easy, medium ou hard)", name)
	}
}
