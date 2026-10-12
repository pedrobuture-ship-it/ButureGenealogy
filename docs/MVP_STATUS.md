# Estado do MVP

Atualizado após a auditoria final (Prompt 12). O MVP **não possui validação zootécnica**: nada aqui deve ser usado como recomendação genética real.

## IMPLEMENTADO

**Fundação** — Wails v2 + Go + React/TypeScript + SQLite (`modernc.org/sqlite`, sem CGO). Funciona offline. `farm.db` no diretório de dados do usuário
(nunca junto do executável), migrations versionadas `0001`–`0007` (com cópia automática do banco antes de aplicar migrations a dados existentes) (cada uma em transação), logger, configuração, health check, fechamento
limpo do banco, erros estruturados (`{"code","message","fields"}`) entre Go e UI.

**Rebanho** — propriedades; animais (cadastro, edição, arquivamento com confirmação, sem exclusão física); busca, filtros (propriedade, raça, sexo,
categoria, status, idade), ordenação e paginação; ficha com abas Resumo, Fenótipo, Genética, Genealogia e Reprodução.

**Genealogia** — árvore de 2 a 5 gerações, ancestrais repetidos, completude/profundidade, ancestrais comuns, detecção e bloqueio de ciclos,
inconsistências (pai fêmea, nascimento posterior, etc.), registro mínimo de ancestral desconhecido.

**Características** — catálogo genérico (categorias, direções MAXIMIZE/MINIMIZE/TARGET_RANGE/INFORMATION_ONLY, unidade, tipo de dado), tipos de valor
separados (Fenótipo, DEP, EPD, EBV, PTA, Índice), validação de valor/unidade, componente `DataProvenance`.

**Bibliotecas** — pacote `.genlib` (leitura segura contra path traversal, arquivos inesperados, symlinks, zip bomb, schema/manifesto inválidos, duplicidade
id+versão), importação de CSV/JSON com assistente de mapeamento, relatório de validação, política de licença (desconhecida/sem redistribuição/redistribuível),
declaração de responsabilidade do usuário, versões coexistentes, ativar/desativar, remoção bloqueada se algum cálculo salvo usou a versão, portão de
distribuição oficial (`go run ./cmd/genlib verify --official`, também verificado por teste).

**Sêmen** — banco de touros (touro = animal macho), lotes (fornecedor, lote, tipo configurável, preço, observações), movimentações com histórico, estoque
**nunca negativo** (UPDATE condicional + CHECK no banco, testado sob concorrência), ficha do touro com estoque, genética (dataset, versão, acurácia, percentil),
pedigree e movimentações. Interface `BullCatalogProvider` preparada (sem implementação nem rede).

**Motor de acasalamento** (`internal/engine`, v0.1.0) — determinístico e auditável; pipeline com uma interface Go por etapa; normalização por percentil
(só com semântica declarada), escore z (só com média/desvio do dataset) e faixa-alvo (com tolerância no perfil); base de comparação única por característica
(mesmo programa/release/base/raça/unidade, nada é convertido); filtros rígidos; complementaridade; penalidade por baixa acurácia (opcional); cobertura mínima;
ranking determinístico; “não calculável” em vez de número sem base; snapshot completo por par; perfis Personalizado, Demonstração Corte e Demonstração Leite
(marcados DEMO); perfis do usuário salvos.

**Ver cálculo** — painel com score final, componentes (score normalizado, peso, contribuição), “Como chegamos nisso?”, complementaridade, penalidades,
decomposição do resultado e “Dados usados neste cálculo”, sempre a partir do snapshot salvo (nada é recalculado no frontend).

**Cálculos** — documentação interna em 13 capítulos, modos simples e técnico, com o status real de cada tema.

**Backup, restauração e integridade** — `.genbackup` (ZIP com `farm.db` consistente via `VACUUM INTO`, bibliotecas particulares, configuração, anexos,
manifesto com SHA-256); restauração com validação (formato, hashes, ZIP traversal, SQLite, versão), atualização de backups antigos, backup de segurança
automático e troca atômica (tudo ou nada); verificação de integridade (SQLite corrompido, migration incompleta/lacuna, tabela ausente, versão incompatível,
referências quebradas, biblioteca ausente/alterada, cálculo histórico com dataset removido, estoque × movimentações, pedigree).

**Auditoria final** — corrigidos: risco de perda de dados no backup (pacotes de bibliotecas “redistribuíveis” declaradas pelo usuário não eram copiados e podiam ser apagados na
restauração; agora são sempre incluídos e, se saírem do conjunto, movidos para `backups/replaced-libraries/`); data “de hoje” em UTC no frontend (agora local) e tolerância de fuso
no backend; mapeamento de coluna a característica existente sem validar tipo/dono; assistente de CSV que não permitia declarar a **semântica do percentil** (sem ela o motor não usa o
percentil); contraste do selo “muted”; marco `aside` repetido na documentação (axe); falha de inicialização sem registro em log. Adicionados: *golden file* do motor
(falha se o comportamento mudar sem nova versão), contrato Go↔TypeScript em tempo de compilação, backup automático pré-migration.

**Qualidade** — testes Go (`go test ./...`), testes de UI (Vitest + Testing Library + axe-core), TypeScript estrito, ESLint, atalhos de teclado,
tratamento global de erros, layout para notebooks, listas paginadas no servidor.

## PARCIAL
- **Reprodução** (aba do animal): só resumo (descendentes, doses). Não há coberturas, diagnósticos de gestação nem partos.
- **Início**: mostra a saúde do banco e do aplicativo; não há painel do rebanho.
- **Pedigree**: ancestrais comuns e completude funcionam; **coeficientes de parentesco/endogamia não são calculados** (aparecem como “não calculado nesta versão”).
- **Motor**: apenas uma regra de penalidade (baixa acurácia); a regra de complementaridade é simples e parametrizada pelo perfil, **não é um modelo validado**.
- **Anexos/fotos**: o formato de backup os suporta, mas o aplicativo ainda não tem funcionalidade de anexos.
- **Bibliotecas públicas**: o mecanismo existe, mas **nenhum dataset real foi incluído** (nenhuma licença foi validada juridicamente).

## MOCKADO / DEMONSTRAÇÃO
- Dados `DEMO / SYNTHETIC` (carregáveis em Configurações): animais, avaliações, bibliotecas e perfis fictícios, sempre marcados. Nunca são ciência real.

## NÃO IMPLEMENTADO
- Página **Relatórios** (placeholder).
- Coeficientes de parentesco e endogamia; progênie esperada; genômica (haplótipos, recessivos).
- Integrações externas (catálogos, APIs); importação em massa de XLSX; sincronização/nuvem; índice econômico; evolução genética do rebanho; otimização do
  rebanho inteiro; qualquer uso de IA.
- Criptografia de backups; senhas/usuários.

## LIMITAÇÕES
- Sem validação zootécnica de traits, pesos, normalizações, penalidades ou complementaridade. Os perfis DEMO são arbitrários.
- Percentis só são usados se o dataset declarar sua semântica; escore z supõe distribuição aproximadamente normal.
- A comparação exige a mesma base (programa, release, base genética, raça, unidade); **a metodologia não pode ser verificada automaticamente**.
- Testado em Linux (webkit2gtk 4.1). Builds de Windows/macOS não foram executados neste ambiente.
- SQLite com uma única conexão (um escritor); adequado a uso de um usuário local.
- Backups não são criptografados.

## PRÓXIMAS ETAPAS
Ver `docs/ROADMAP_POST_MVP.md` (validação científica com especialista, primeiro pacote real, motor formal de pedigree, relatórios, ...).

## Como verificar
```
go vet -tags webkit2_41 ./... && go test ./...
cd frontend && npm run typecheck && npm run lint && npm test
wails build -tags webkit2_41
```
