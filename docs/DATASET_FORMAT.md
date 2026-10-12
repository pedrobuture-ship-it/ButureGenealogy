# Formato de dataset: `.genlib`

Uma **biblioteca** é um pacote ZIP chamado `*.genlib`. Ela traz avaliações genéticas e as definições das características
(`docs/TRAITS.md`). O conhecimento genético não fica no código: vem daqui.

Código: `internal/datasets` (leitura, validação, extração) e `internal/service/library.go` (instalação).

## Conteúdo

```
dataset.genlib
├── manifest.json     (obrigatório)
├── traits.json       (obrigatório)
├── data.csv ou data.json   (exatamente um)
├── README.md         (opcional)
└── LICENSE           (opcional; também aceita LICENSE.txt)
```

Nenhum outro nome, diretório, caminho com `..`, `/`, `\` ou `:` é aceito.

### manifest.json
```json
{
  "schemaVersion": 1,
  "id": "example.dataset",
  "name": "Example Dataset",
  "version": "1.0.0",
  "species": "Bos taurus",
  "breeds": ["NEL"],
  "purpose": ["beef"],
  "authors": [],
  "organization": "",
  "sourceUrl": "",
  "releaseDate": "",
  "license": { "name": "", "url": "", "redistributionAllowed": false, "commercialUseAllowed": null },
  "evaluation": { "program": "Programa X", "release": "2026.1", "geneticBase": "base 2020" },
  "data": { "format": "csv", "idColumn": "registro", "nameColumn": "nome", "sexColumn": "", "defaultSex": "male",
            "decimalComma": false, "accuracyScale": "fraction" }
}
```
- `id`: `[a-z0-9][a-z0-9._-]{1,99}`; `version`: `[0-9A-Za-z][0-9A-Za-z._+-]{0,49}`.
- `evaluation.program` é obrigatório (sem programa não há proveniência). `release` vazio assume a versão.
- `license`: nome vazio = **licença desconhecida**. `redistributionAllowed` é `false` salvo declaração explícita.
- `data.accuracyScale`: `fraction` (0–1, padrão) ou `percent` (0–100, convertida para fração).
- `data.defaultSex` vale quando não há `sexColumn` (`male` para catálogos de touros).

### traits.json
Lista de definições + colunas de dados:
```json
[{ "code": "WW", "name": "Peso à desmama", "category": "growth", "unit": "kg", "evaluationType": "DEP",
   "column": "DEP_PD", "accuracyColumn": "ACC_PD", "percentileColumn": "PCT_PD" }]
```
Bibliotecas fornecem `DEP`, `EPD`, `EBV`, `PTA` ou `INDEX` (nunca fenótipo). Código + tipo não podem repetir; uma coluna não pode
servir a duas características. Demais campos seguem `docs/TRAITS.md`.

`metadata` pode declarar o que o motor de acasalamento precisa para normalizar os valores:
```json
"metadata": { "percentileSemantics": "ascending", "population": { "mean": 10.0, "sd": 2.0 } }
```
Sem `percentileSemantics` o percentil é só exibido (o motor não o usa); sem `population`, o escore z não está disponível.
O assistente de CSV/JSON pede essas informações ao mapear uma coluna de percentil.

### Dados
Formato largo: uma linha por animal, uma coluna por característica (+ colunas opcionais de acurácia e percentil).
- CSV: delimitador `,` `;` ou tab (detectado), UTF-8 (BOM aceito) ou Latin-1 (fallback). JSON: array de objetos planos.
- Células vazias, `NA`, `NaN`, `-` = ausente. Texto não numérico é contado e ignorado.
- Identificação repetida (sem diferenciar maiúsculas/espaços): vale a primeira linha.

## Segurança da importação
Reforçada em `ReadPackage` e coberta por testes (`internal/datasets/datasets_test.go`):
- **Path traversal / nomes suspeitos**: rejeitados antes de ler qualquer conteúdo. Nada é extraído para o disco com o nome do ZIP: o
  pacote é guardado como `libraries/<uuid>.genlib`.
- **Arquivos inesperados, diretórios, links simbólicos, duplicatas**: rejeitados.
- **Tamanho**: limites aplicados aos bytes **realmente lidos** (não ao declarado): pacote 64 MB, dados 128 MB, textos 1 MB, total 160 MB,
  10 arquivos, 500 mil linhas, 500 colunas.
- **Manifesto inválido / schema incompatível / JSON malformado**: rejeitados com mensagem clara.
- **Duplicidade id + versão**: recusada; versões diferentes coexistem.
- Operações de instalação são atômicas (transação + arquivo temporário); falha não deixa resíduos.

## Licenças e distribuição
- `LicenseStatus`: `unknown` (sem licença identificada), `no_redistribute`, `redistributable`. Nunca se presume.
- **Distribuição oficial** (`/datasets`): só licença conhecida (nome + URL) **e** redistribuição explicitamente permitida.
  Verificado por teste e por `go run ./cmd/genlib verify --official`.
- **Importação manual** (CSV/JSON/.genlib): permitida com qualquer licença, mas o usuário deve confirmar:
  > A responsabilidade pela autorização de uso e distribuição de dados importados manualmente é do usuário.
  Esses dados ficam só neste computador e nunca são copiados para outros usuários.

## Instalação, versões e histórico
- Cada instalação cria uma linha em `datasets` (por versão), guarda o `.genlib` original e seu SHA-256, define as características
  da biblioteca (`traits.source_dataset = id da instalação`) e grava as avaliações em `genetic_evaluations` com `dataset_id`.
- Registros são associados a animais pelo **registro** (sem diferenciar maiúsculas/espaços). Sem correspondência: ignorar ou cadastrar
  como novos animais (escolha do usuário). Mais de uma correspondência: ignorado, nunca adivinhado.
- **Desativar** esconde os valores da biblioteca sem apagar nada. **Remover** apaga valores e características da versão, mas é
  **bloqueado** se algum cálculo salvo a usou (`calculation_datasets`), preservando a reprodutibilidade.

## Importação de CSV/JSON genérico
O assistente pede metadados, a licença (em branco = desconhecida), as colunas de identificação, nome, sexo, vírgula decimal,
escala da acurácia, e o mapeamento coluna → característica/tipo/unidade. Cada importação vira um `.genlib` guardado.
Antes de instalar mostra a validação:
```
VALIDAÇÃO
8.412 registros · 23 características reconhecidas · 482 sem acurácia · 17 duplicados · 3 traits não reconhecidos
```

## Backup
O `.genbackup` inclui o pacote original de **toda** biblioteca importada pelo usuário (mesmo que ela declare licença redistribuível: o usuário pode ser o único a ter o arquivo).
Só bibliotecas **oficiais** (distribuídas com o aplicativo, `datasets.is_official = 1`; hoje nenhuma) entram apenas como metadados (id, versão, SHA-256).
Ao restaurar, pacotes que deixam de pertencer aos dados restaurados são movidos para `backups/replaced-libraries/`, nunca apagados.
