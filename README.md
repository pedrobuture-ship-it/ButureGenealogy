# Genética

Plataforma desktop, **local-first**, aberta e auditável para planejamento de melhoramento genético bovino (Go · React/TypeScript · SQLite · Wails).
Funciona offline; os dados ficam no computador do usuário.

> O MVP **não possui validação zootécnica**. Perfis e dados de demonstração são fictícios. Veja `docs/SCIENTIFIC_LIMITATIONS.md`.

## Documentos
- `docs/MVP_STATUS.md` — o que está implementado, parcial e não implementado
- `docs/ARCHITECTURE_FINAL.md` — arquitetura e princípios
- `docs/CALCULATION_ENGINE.md` — motor de acasalamento
- `docs/DATASET_FORMAT.md` · `docs/TRAITS.md` — bibliotecas e características
- `docs/SCIENTIFIC_LIMITATIONS.md` · `docs/ROADMAP_POST_MVP.md`
- `prompts.txt` — as etapas de desenvolvimento

## Desenvolvimento
```
cd frontend && npm install
wails dev   -tags webkit2_41        # Linux com webkit2gtk 4.1
wails build -tags webkit2_41
go vet -tags webkit2_41 ./... && go test ./...
cd frontend && npm run typecheck && npm run lint && npm test
GENETICA_DATA_DIR=/tmp/demo go run ./cmd/seeddemo   # dados DEMO / SYNTHETIC num diretório de teste
```
