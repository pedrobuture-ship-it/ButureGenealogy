# Sistema de características (traits)

Uma **característica** é *dado*, não código. Nenhuma regra do aplicativo pode depender de uma raça ou de um nome de
característica (`if breed == "Nelore"` é proibido). O que muda entre raças e programas vem de bibliotecas (datasets).

Código: `internal/domain/trait.go` (modelo e validação), `migrations/0003_traits.sql` (tabela `traits`),
`internal/service/trait.go` (catálogo e visão por animal).

## Modelo

| Campo | Significado |
|---|---|
| `code` | Identificador estável (`[A-Za-z][A-Za-z0-9_]*`), ex.: `WW`, `BODY_WEIGHT`. |
| `name`, `description` | Texto exibido. |
| `category` | `growth`, `reproduction`, `maternal`, `carcass`, `dairy`, `conformation`, `health`, `temperament`, `efficiency`, `other`. |
| `purpose` | `beef`, `dairy`, `dual` ou `any`. |
| `unit` | Unidade do valor (kg, cm, …). Vazia se não se aplica. |
| `data_type` | `number`, `integer` ou `score` (escore inteiro). |
| `evaluation_type` | **Tipo de valor** descrito: `PHENOTYPE`, `DEP`, `EPD`, `EBV`, `PTA` ou `INDEX`. |
| `optimization_direction` | `MAXIMIZE`, `MINIMIZE`, `TARGET_RANGE`, `INFORMATION_ONLY`. |
| `ideal_min`, `ideal_max` | Faixa-alvo (obrigatórias para `TARGET_RANGE`; `min ≤ max`). |
| `higher_is_better` | Só para `MAXIMIZE` (verdadeiro) / `MINIMIZE` (falso); proibido em `TARGET_RANGE` e `INFORMATION_ONLY`. |
| `source_dataset` | Id da **instalação** (versão) da biblioteca que definiu a característica (vazio = nativa ou criada pelo usuário). Cada versão tem suas próprias definições. |
| `metadata` | Objeto JSON livre. |

A chave única é `(code, evaluation_type, source_dataset)`. Assim `WW` como DEP e `WW` como EBV são características
diferentes, e a definição de um dataset prevalece sobre a nativa quando o valor veio dele.

### Direções de otimização
- **MAXIMIZE** — valores maiores são melhores.
- **MINIMIZE** — valores menores são melhores.
- **TARGET_RANGE** — melhor é estar dentro de `[ideal_min, ideal_max]`.
- **INFORMATION_ONLY** — exibida, mas não entra em pontuação.

As características nativas do aplicativo são **só nomes neutros** e usam `INFORMATION_ONLY`: o que é "melhor" depende
do objetivo de melhoramento (perfil de acasalamento) ou do dataset, nunca do aplicativo.

## Tipos de valor — não são a mesma coisa

| Tipo | O que é |
|---|---|
| Fenótipo | Medida observada no animal; mistura genética e ambiente. Não é mérito genético. |
| DEP / EPD | Diferença esperada na progênie, publicada por um programa, relativa à base genética daquele programa. |
| EBV | Estimativa do valor genético aditivo do animal, relativa à base do programa. |
| PTA | Habilidade de transmissão prevista, relativa à base do programa. |
| Índice | Combinação ponderada de características definida por quem o publica. |

Regras impostas pelo sistema:

1. Um valor só é aceito se existir a característica para aquele **tipo** (`WW` PTA não é aceito só porque `WW` DEP existe).
2. A interface sempre rotula o tipo e **nunca soma, converte ou compara** tipos diferentes (`domain.AreComparable`).
3. Mesmo com o mesmo tipo, comparar exige o mesmo programa, release, base genética, característica e unidade; essa
   verificação é do motor de cálculo (ver `docs/CALCULATION_ENGINE.md`).
4. Unidades divergentes da característica são **rejeitadas**, nunca convertidas silenciosamente.
5. Valores inteiros/escores recusam decimais.

## Proveniência (`DataProvenance`)

Componente `frontend/src/components/DataProvenance.tsx`, alimentado por `service.Provenance`:

```
Característica: Peso à desmama
Tipo: DEP
Dataset: X
Versão: Y
Programa: Z
Acurácia: 81%
```

Para fenótipos mostra a origem da medição e a data. Dados `DEMO / SYNTHETIC` são sempre sinalizados.

Acurácia é guardada como fração `[0,1]` e exibida em %. Percentil em `[0,100]`.

## Na interface do animal
As abas **Fenótipo** e **Genética** agrupam por categoria e mostram unidade, valor, tipo, acurácia e percentil
(quando existem) e a origem de cada valor. Valores sem definição no catálogo aparecem marcados "sem definição".

## Metadados que o motor de acasalamento entende
Em `metadata` (definição da característica), declarados pelo dataset ou pelo usuário no assistente de importação:

```json
{ "percentileSemantics": "ascending", "population": { "mean": 10.0, "sd": 2.0 } }
```
- `percentileSemantics`: `ascending` (valor maior ⇒ percentil maior) ou `descending` (valor maior ⇒ percentil menor). **Sem isso o percentil não é usado.**
- `population.mean` / `population.sd`: habilitam a normalização por escore z (também aceitos como `populationMean`/`populationSd`).
Ver `docs/CALCULATION_ENGINE.md`.
