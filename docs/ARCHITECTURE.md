# Arquitetura

A descrição completa e atual está em **`docs/ARCHITECTURE_FINAL.md`**.

Resumo: aplicativo desktop local-first (Wails v2 + Go + React/TypeScript + SQLite). O frontend só apresenta; `app.go` é uma ponte fina; as regras ficam em
`internal/service` (casos de uso), `internal/engine` (motor de acasalamento), `internal/genetics` (pedigree), `internal/datasets` (formato `.genlib`),
`internal/domain` (entidades) e `internal/backup`; todo SQL fica em `internal/repository`.

```
cd frontend && npm install
wails dev   -tags webkit2_41      # Linux com webkit2gtk 4.1
wails build -tags webkit2_41
go test ./... && (cd frontend && npm run typecheck && npm run lint && npm test)
```
