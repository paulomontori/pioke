package synth

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
