# Roteiro pós-MVP

Princípio: **primeiro significado técnico, depois mais telas.** A ordem abaixo é a recomendada; cada fase só avança quando a anterior estiver validada.

## Fase A — Validação científica (próximo passo)
Trabalhar com zootecnista/geneticista para: validar traits e direções; pesos; índices; normalizações (quando percentil/escore z são apropriados); a regra de
complementaridade e as penalidades; a interpretação de DEP/EPD/EBV/PTA de cada programa; documentar **referências científicas**; criar testes com casos conhecidos
(entradas e saídas esperadas fornecidas por especialistas). Resultado: um perfil real validado e a retirada do rótulo DEMO das partes aprovadas.
A meta não é ter mais telas, é garantir que as recomendações tenham significado técnico.

## Fase B — Primeiro pacote real
Escolher **uma** raça/finalidade (ex.: Nelore/corte ou Gir/leite). Encontrar apenas fontes gratuitas, legalmente utilizáveis, com licença conhecida e compatível com
a distribuição pretendida (confirmar com a fonte). Criar o primeiro `.genlib` real com `traits.json` (direções, unidades, `percentileSemantics`, parâmetros populacionais)
e documentação; validar com `go run ./cmd/genlib verify --official` e com testes.

## Fase C — Motor de pedigree profissional
Implementar formalmente: matriz de parentesco aditivo (algoritmo tabular de Henderson), coeficiente de relacionamento, **coeficiente de endogamia de Wright**,
ancestrais comuns, completude, profundidade e alertas, atrás de `genetics.RelationshipCalculator`; testar contra casos de referência publicados; documentar fórmulas e
referências na página Cálculos; só então ativar a penalidade de parentesco no motor (nova versão do motor).

## Fase D — Genômica
Informações genômicas, haplótipos, recessivos, condições genéticas e incompatibilidades, vindas de fontes apropriadas. Exemplo: “Touro X × Matriz Y: risco recessivo detectado → combinação alertada/bloqueada”.

## Fase E — Mais raças
Adicionar bibliotecas (Gir, Guzerá, Sindi, Brahman, Tabapuã, Holandês, Jersey, Angus, Hereford, …); o motor permanece genérico.

## Fase F — Integrações
Implementar `domain.BullCatalogProvider` e uma abstração `GeneticDataProvider` (`SearchAnimal`, `GetPedigree`, `GetEvaluations`) para fontes externas, sem dependência no domínio.
Nunca basear o funcionamento offline nessas APIs.

## Fase G — Otimização do rebanho inteiro
Hoje: 1 matriz × N touros. Futuro: N matrizes × M touros × estoque limitado × restrições (parentesco, endogamia máxima, preço, partos, objetivos). Problema formal de
otimização: programação linear inteira, constraint programming, otimização multiobjetivo.

## Fase H — Índice econômico
Retorno econômico esperado (leite, sólidos, fertilidade, longevidade, carcaça, eficiência, custo do sêmen). Nunca inventar valores econômicos universais: vêm de biblioteca ou do usuário.

## Fase I — Evolução genética do rebanho
Dashboards por ano e por característica: tendência genética e fenotípica, ganho esperado × observado, diversidade e endogamia. Sempre separar fenótipo de mérito genético.

## Fase J — Importação em massa
CSV, XLSX, JSON e API com mapeamento, pré-visualização, validação, deduplicação, desfazer (rollback) e relatório de erros, para propriedades maiores.

## Fase K — Relatórios profissionais
Planejamento de acasalamento, matriz individual, touro, pedigree, risco genético, evolução do rebanho, sêmen necessário, diversidade, metodologia e datasets utilizados;
cada relatório com sua proveniência. (A página Relatórios hoje é um placeholder.)

## Fase L — Ecossistema de bibliotecas
Catálogo de datasets que aponta para fontes autorizadas (instalar com um clique); instituições e pesquisadores publicando seus próprios `.genlib`.

## Fase M — Sincronização opcional
Backup criptografado em nuvem, sincronização entre computadores, equipes, várias fazendas, compartilhamento com veterinário/zootecnista — sem que o programa passe a depender da nuvem.

## Fase N — IA, somente depois
IA **explica**, não calcula: dados + cálculo determinístico → IA → explicação em linguagem natural (“O Touro A foi favorecido principalmente por melhorar X e Y da matriz…”).
A IA não inventa o score.

## Dívidas técnicas a pagar no caminho
- Aba Reprodução (coberturas, diagnósticos, partos) e painel do Início.
- Anexos/fotos (o formato de backup já os prevê).
- Criptografia opcional de backups.
- Builds e testes em Windows e macOS; empacotamento/assinatura.
- Fonte de `wailsjs/go/models.ts` no CI para garantir o contrato Go↔TypeScript (já verificado em `npm run typecheck`).
