# Bibliotecas distribuídas com o aplicativo

Esta pasta só pode conter pacotes `.genlib` cuja licença:

1. seja **conhecida** (nome e URL no `manifest.json`), e
2. **permita redistribuição** (`license.redistributionAllowed = true`).

"Disponível gratuitamente na internet" **não** é uma licença. Se não for possível comprovar a licença, o dataset não entra aqui:
o usuário pode importá-lo manualmente pela página Bibliotecas, sob sua responsabilidade.

A regra é verificada automaticamente (`internal/datasets/bundle_test.go`) e pode ser checada à mão:

```
go run ./cmd/genlib verify --official datasets/*.genlib
```

Atualmente esta pasta está vazia: nenhum dataset real foi validado juridicamente ainda (ver `docs/ROADMAP_POST_MVP.md`).
