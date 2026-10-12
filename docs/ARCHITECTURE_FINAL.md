# Arquitetura final do MVP

Plataforma desktop, **local-first**, aberta e auditável para planejamento de melhoramento genético bovino
(Wails v2 · Go · React/TypeScript · SQLite sem CGO). Nenhuma parte do aplicativo depende de rede.

## Conceito

```
          DATASETS / BIBLIOTECAS
                   ↓
             MOTOR GENÉTICO
                   ↓
       ┌───────────┼───────────┐
       ↓           ↓           ↓
   PEDIGREE    FENÓTIPO    GENÉTICA
       └───────────┼───────────┘
                   ↓
              ACASALAMENTO
                   ↓
                 RANKING
                   ↓
       ┌───────────┴───────────┐
       ↓                       ↓
 RECOMENDAÇÃO             VER CÁLCULO
                               ↓
                    DADOS + MÉTODO + FONTE
```

Fluxo do usuário: cadastrar rebanho → manter pedigree → registrar fenótipos → instalar bibliotecas → importar avaliações → manter touros e
estoque → escolher uma matriz → definir objetivos → comparar touros compatíveis → receber ranking → “Ver cálculo”.

## Camadas

```
frontend (React/TS) ─ Wails bindings ─▶ app.go ─▶ internal/service ─▶ internal/repository ─▶ SQLite (farm.db)
  só apresenta           ponte fina        casos de uso          SQL isolado
                                              │
                                              ├─▶ internal/engine     motor de acasalamento (puro)
                                              ├─▶ internal/genetics   algoritmos de pedigree (puro)
                                              ├─▶ internal/datasets   formato .genlib (puro)
                                              ├─▶ internal/domain     entidades + validação (sem SQL/UI)
                                              └─▶ internal/backup     backup, restauração, integridade
```

| Pacote | Responsabilidade | Depende de |
|---|---|---|
| `domain` | Entidades, enums, validação, regras de tipos de valor | nada |
| `genetics` | Ancestrais, ancestrais comuns, ciclos, completude, árvore; interface de parentesco/endogamia (não implementada) | nada |
| `datasets` | Ler/validar `.genlib`, tabelas CSV/JSON, extração de registros, política de licença | `domain` |
| `engine` | Pipeline de acasalamento, perfis, snapshots | `genetics` |
| `repository` | Todo o SQL de leitura/escrita, mapeamento de erros do SQLite | `domain` |
| `service` | Casos de uso, transações, DTOs, erros para a UI | todos acima |
| `database` | Abrir SQLite com pragmas seguros, migrations, transações, health | — |
| `backup` | `.genbackup`, restauração atômica, verificações de integridade | `database`, `genetics` |
| `config`, `logger`, `apperr` | Caminhos/configuração, `slog`, erros com código estável | — |

O **frontend** não contém SQL nem regras de negócio: `utils/api.ts` é o único acesso ao Go; `types/*.ts` espelham os DTOs e um arquivo de
contrato (`types/contract.check.ts`) compara, em tempo de compilação, os campos com os modelos que o Wails gera de Go.

## Banco de dados (`farm.db`)

Migrations embutidas `0001`–`0007`, cada uma em transação com seu registro em `schema_migrations`. Pragmas: `foreign_keys=ON`, `WAL`, `busy_timeout`;
**uma conexão** (um escritor). Antes de aplicar migrations a um banco com dados, uma cópia consistente (`VACUUM INTO`) é guardada em `backups/`.

| Grupo | Tabelas |
|---|---|
| Rebanho | `farms`, `breeds`, `animals`, `animal_measurements` |
| Genética | `traits`, `datasets`, `genetic_evaluations` |
| Sêmen | `semen_inventory`, `semen_movements`, `semen_types` |
| Acasalamento | `calculation_runs`, `matings` (snapshot JSON), `calculation_datasets`, `mating_profiles` |
| Controle | `schema_migrations`, `app_meta` |

Identificadores UUID; datas `YYYY-MM-DD`; instantes RFC 3339 UTC. Linhas fictícias levam `is_demo = 1`.
Fora do `farm.db`: `libraries/<id>.genlib` (pacote original de cada biblioteca), `backups/`, `logs/`, `config.json`.

## Como cada princípio é garantido

**1. Local-first.** Sem código de rede (verificado por varredura na auditoria); fontes do sistema; dados em `~/.local/share/Genetica` (Linux),
`~/Library/Application Support/Genetica` (macOS) ou `%AppData%\Genetica` (Windows). Backup/restauração e integridade funcionam offline.
Integrações externas são apenas uma interface (`domain.BullCatalogProvider`) sem implementação.

**2. Datasets como bibliotecas.** Características e avaliações vêm de pacotes `.genlib` (`docs/DATASET_FORMAT.md`); o código só conhece o conceito de
Trait (`docs/TRAITS.md`). Nenhuma regra por raça. Distribuição oficial só com licença conhecida e redistribuição permitida (teste + `cmd/genlib`).

**3. Não ser caixa-preta.** Todo par matriz × touro guarda um snapshot autocontido (versão do motor, versões e SHA-256 dos datasets, valores com
proveniência, perfil completo, bases de comparação, resultado de cada etapa). “Ver cálculo” lê o snapshot; o frontend nunca recalcula.

**4. Não inventar ciência.** O motor só usa métodos documentados e suportados; sem dado/método suficiente responde “não calculável”. Tipos de valor
(fenótipo, DEP, EPD, EBV, PTA, índice) nunca se misturam; bases diferentes nunca se convertem; parentesco/endogamia é “planejado” e nunca aparece como número.
Detalhes em `docs/CALCULATION_ENGINE.md` e `docs/SCIENTIFIC_LIMITATIONS.md`.

## Qualidade e verificação

- Go: `go vet` e `go test ./...` (domínio, repositórios, datasets incl. ZIPs hostis, engine incl. propriedades aleatórias e *golden file*, serviços, backup/restauração/integridade, concorrência de estoque, desempenho com 20 000 animais).
- Frontend: `tsc --noEmit` (inclui o contrato Go↔TS), ESLint, Vitest + Testing Library (fluxos) + axe-core (acessibilidade).
- Build: `wails build -tags webkit2_41` (Linux com webkit2gtk 4.1).

## Concorrência
Uma conexão SQLite serializa as operações. Regra de código: dentro de `WithTx` só se usa o `tx` (nunca `s.db`), senão haveria *deadlock*; baixa de estoque é um
`UPDATE … WHERE quantity + delta >= 0` condicional; a restauração usa uma conexão fixada e uma única transação.

## Estrutura de pastas
```
main.go · app.go          Wails: ponte fina e ciclo de vida
cmd/seeddemo · cmd/genlib ferramentas de desenvolvimento
internal/…                pacotes acima
migrations/               SQL versionado (embutido)
datasets/                 só bibliotecas com licença verificada (hoje vazia)
frontend/src/             components · features · hooks · layouts · pages · types · utils
docs/                     esta documentação
```
