package scoring

import "math"

// yinResult é a saída de uma detecção de pitch YIN para um único frame de áudio.
type yinResult struct {
	FreqHz     float64
	Confidence float64 // 0-1: 1 = periodicidade perfeita, 0 = ruído puro
	OK         bool
}

// detectPitchYIN estima a frequência fundamental de frame usando o algoritmo YIN (de Cheveigné &
// Kawahara, 2002): função de diferença normalizada + limiar absoluto + interpolação parabólica.
// Robusto a oitavas erradas em comparação com autocorrelação simples, que é o motivo de YIN ser a
// escolha padrão para detecção de pitch de voz cantada.
//
// frame deve ter pelo menos 2*maxTau amostras: a função de diferença compara frame[j] com
// frame[j+tau] para tau até maxTau, então precisa de amostras além da janela de análise
// (maxTau = sampleRate/minFreqHz) para "olhar à frente".
func detectPitchYIN(frame []float64, sampleRate int, minFreqHz, maxFreqHz, threshold float64) yinResult {
	maxTau := int(float64(sampleRate) / minFreqHz)
	minTau := max(int(float64(sampleRate)/maxFreqHz), 1)
	if maxTau < minTau+1 || len(frame) < 2*maxTau {
		return yinResult{}
	}

	// Passo 1: função de diferença d(tau) = soma_j (frame[j] - frame[j+tau])^2, para j no início
	// da janela (tamanho maxTau).
	diff := make([]float64, maxTau)
	for tau := 1; tau < maxTau; tau++ {
		var sum float64
		for j := range maxTau {
			d := frame[j] - frame[j+tau]
			sum += d * d
		}
		diff[tau] = sum
	}

	// Passo 2: função de diferença média normalizada cumulativa (CMNDF). cmndf[0] é fixado em 1
	// por convenção (nunca escolhido como tau).
	cmndf := make([]float64, maxTau)
	cmndf[0] = 1
	var runningSum float64
	for tau := 1; tau < maxTau; tau++ {
		runningSum += diff[tau]
		if runningSum == 0 {
			cmndf[tau] = 1
		} else {
			cmndf[tau] = diff[tau] * float64(tau) / runningSum
		}
	}

	// Passo 3: limiar absoluto — primeiro mínimo local abaixo do limiar, dentro da faixa de tau
	// válida [minTau, maxTau). Sem isso, o algoritmo tenderia a escolher tau muito pequenos
	// (frequências altas espúrias) que também produzem cmndf baixo.
	bestTau := -1
	for tau := minTau; tau < maxTau-1; tau++ {
		if cmndf[tau] < threshold {
			for tau+1 < maxTau-1 && cmndf[tau+1] < cmndf[tau] {
				tau++
			}
			bestTau = tau
			break
		}
	}

	// Nenhum tau cruzou o limiar: usa o mínimo global como estimativa de fallback, com
	// confiança mais baixa (o chamador decide, via gate de vozeamento, se aceita essa amostra).
	if bestTau < 0 {
		minVal := math.Inf(1)
		for tau := minTau; tau < maxTau; tau++ {
			if cmndf[tau] < minVal {
				minVal = cmndf[tau]
				bestTau = tau
			}
		}
	}
	if bestTau < minTau || bestTau >= maxTau-1 || bestTau <= 0 {
		return yinResult{}
	}

	// Passo 4: interpolação parabólica em torno de bestTau para refinar a estimativa de período
	// além da resolução de uma amostra.
	tauRefined := parabolicInterpolate(cmndf, bestTau)
	if tauRefined <= 0 {
		return yinResult{}
	}

	freq := float64(sampleRate) / tauRefined
	confidence := 1 - cmndf[bestTau]
	if confidence < 0 {
		confidence = 0
	}

	if freq < minFreqHz || freq > maxFreqHz {
		return yinResult{}
	}

	return yinResult{FreqHz: freq, Confidence: confidence, OK: true}
}

// parabolicInterpolate ajusta uma parábola por (tau-1, tau, tau+1) em cmndf e retorna o tau
// (fracionário) do vértice — refina a estimativa de período para além da resolução de uma amostra.
func parabolicInterpolate(cmndf []float64, tau int) float64 {
	if tau <= 0 || tau+1 >= len(cmndf) {
		return float64(tau)
	}
	s0, s1, s2 := cmndf[tau-1], cmndf[tau], cmndf[tau+1]
	denom := s0 + s2 - 2*s1
	if denom == 0 {
		return float64(tau)
	}
	offset := 0.5 * (s0 - s2) / denom
	return float64(tau) + offset
}
