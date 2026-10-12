# Motor de cálculo (`internal/engine`, versão 0.1.0)

Determinístico, auditável e **sem IA**. Mesma entrada ⇒ mesma saída, byte a byte (sem aleatoriedade; a data só é registrada). O motor é puro: não
acessa SQL nem UI; quem monta a entrada e grava o resultado é `service.MatingService`.

## Pipeline e interfaces (`interfaces.go`)

```
Candidate → Validation → Hard Filters → Trait Compatibility → Normalization
          → Weighted Components → Complementarity → Penalties → Final Score
```

| Etapa | Interface | Implementação padrão |
|---|---|---|
| Validação da execução | `RunValidator` | `DefaultRunValidator` (matriz fêmea, ≥1 touro, perfil válido) |
| Validação do candidato | `CandidateValidator` | `DefaultCandidateValidator` (macho, ≠ matriz) |
| Filtros rígidos | `HardFilter` | status, estoque, custo, características obrigatórias, acurácia mínima, ancestral direto |
| Compatibilidade | `BasisResolver`, `EvaluationSelector` | base de referência por característica |
| Normalização | `Normalizer` | `DefaultNormalizer` |
| Componentes | `ScoringComponent` (`Calculate(ctx CalculationContext) ComponentResult`) | `TraitComponent` |
| Complementaridade | `ComplementarityScorer` | `DefaultComplementarity` |
| Penalidades | `PenaltyRule` | `LowAccuracyRule`, `InbreedingRule` (não aplicada) |
| Pontuação final | `ScoreAggregator` | `DefaultAggregator` |

`engine.New()` monta os padrões; cada campo de `Engine` é uma interface substituível.

## Entrada

`RunInput`: `Female` (+ avaliações), `Candidates` (+ avaliações, estoque, custo), `Profile`, `Pedigree` (grafo), `Datasets` (identificação), `Now`.
Uma `Evaluation` carrega valor, acurácia (fração 0–1), percentil (0–100), **base** (`Basis`), dataset/versão e metadados da característica.

### Base de comparação (`Basis`)
```
característica · tipo (DEP/EPD/EBV/PTA/INDEX) · programa · release · base genética · raça · unidade
```
Duas avaliações só são comparadas se a chave for idêntica. Para cada característica do perfil o motor escolhe **uma** base de referência:
mais candidatos elegíveis avaliados nela → empate: a base em que a matriz também foi avaliada → empate: ordem alfabética da chave. A escolha, as alternativas
e o motivo são gravados. Candidatos só avaliados em outra base ficam **não calculáveis** nessa característica, com
`⚠ Comparação genética não suportada entre essas bases.` Nada é convertido.

## Normalização (`normalize.go`)

Sempre 0–100. Direção: a do perfil, senão a da característica; `INFORMATION_ONLY` sem direção ⇒ não calculável.

| Método | Exige | Fórmula |
|---|---|---|
| Percentil | `percentileSemantics` = `ascending`/`descending` declarado pelo dataset | asc.: MAX → `p`, MIN → `100−p`; desc.: MAX → `100−p`, MIN → `p` |
| Escore z | `population.mean`, `population.sd` (σ>0) nos metadados | `z=(x−μ)/σ`; MAX → `100·Φ(z)`; MIN → `100·(1−Φ(z))` |
| Faixa-alvo | min e max ideais + `targetTolerance` do perfil | `100·max(0, 1 − distância/tolerância)` |

Modo do perfil: `auto` (percentil, senão z), `percentile`, `zscore`. Sem método aplicável ⇒ `Reason` explícito, nunca um valor “aproximado”.

## Pontuação (`aggregate.go`)

```
peso_norm_i   = peso_i / Σ peso (componentes calculáveis)
contribuição  = score_i × peso_norm_i
base          = Σ contribuição
final         = limitar(base + complementaridade − penalidades, 0, 100)
cobertura     = peso calculável / peso total
```
- `cobertura < minCoverage` (padrão 100%) ⇒ **sem score** (não calculável). Com `minCoverage` menor, os pesos são renormalizados e a cobertura é exibida.
- O valor sem limitação (`unclamped`) e o aviso de corte são gravados.
- Ranking: score ↓; empate: nome (sem diferenciar maiúsculas) e depois id. Não calculáveis e excluídos vêm depois, sem posição. `Top` = 5 melhores.

### Complementaridade (`complement.go`)
`ganho_t = limitar((score_touro_t − score_matriz_t)/100, 0, 1)`; `pontos = bônus_máximo × Σ peso_norm_t × ganho_t`. Mesma base e mesmo método para touro e matriz.
Sem avaliação comparável da matriz ⇒ o item contribui 0 (com aviso). Nunca negativa. É uma **regra de configuração do MVP**, não um modelo validado.

### Penalidades (`penalties.go`)
- **Baixa acurácia** (opcional): `pontos = máximo × Σ peso_norm_t × max(0,(limiar−acurácia_t)/limiar)`. Acurácia desconhecida não penaliza (e gera aviso).
- **Parentesco/endogamia**: *não aplicada* (0 ponto): “Planejado — ainda não utilizado no score do MVP”.

### Uso da acurácia
Não altera o score de cada característica. Aparece em todas as telas e só atua como filtro (mínimo; desconhecida exclui) ou penalidade, ambos escolhidos no perfil.

### Pedigree no motor
Ancestrais comuns entre matriz e touro viram **avisos** (`pedigreeNotes`); `excludeDirectAncestors` é um filtro rígido. Nenhum altera o score.

## Perfis (`profiles.go`)
`Personalizado` (vazio), `Demonstração Corte`, `Demonstração Leite` (marcados `DEMO — NÃO REPRESENTA RECOMENDAÇÃO ZOOTÉCNICA VALIDADA`; pesos e direções arbitrários).
Todo parâmetro usado vem do perfil e é copiado ao snapshot; o código não tem constantes de pontuação (somente a tolerância numérica `1e-9`, `TopN=5`).
Perfis do usuário ficam em `mating_profiles`; copiar um perfil DEMO gera um perfil do usuário (não DEMO).

## Resultado e snapshot (`types.go`, `snapshot.go`)
`ComponentResult` guarda nome, valor de entrada (touro e matriz), unidade, acurácia, percentil, score, método, transformação, fórmula, peso bruto e normalizado,
contribuição, base, fonte (dataset, versão, programa, release, base, parâmetro do perfil), explicação e avisos.
`PairSnapshot` (versão de formato 1, um por par): matriz, touro, avaliações das características do perfil (inclusive as de bases não usadas), perfil, bases de
referência, datasets usados (id, versão, SHA-256, DEMO), resultado completo, tamanho da execução, avisos e marca DEMO.
Tabelas: `calculation_runs` (execução), `matings` (par + snapshot JSON), `calculation_datasets` (versões usadas: protege contra remoção).

## Versionamento e regressão
`engine.Version` muda sempre que o comportamento muda. `internal/engine/testdata/golden_run_v<versão>.json` fixa a saída completa de um cenário
sintético que cobre todas as etapas; se o motor mudar sem novo número de versão, `TestGoldenRun` falha. Para uma mudança intencional: aumentar a versão e
rodar `go test ./internal/engine -run Golden -update`.

## Testes
Cálculos à mão (`engine_test.go`), propriedades sobre 300 cenários aleatórios (pesos somam 1, contribuições somam a base, fórmula do score, ordem do ranking,
cobertura), determinismo e independência da ordem de entrada, hostilidade de entrada, filtros, bases incompatíveis, snapshots (ida e volta JSON).
