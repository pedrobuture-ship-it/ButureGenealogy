# Limitações científicas

O aplicativo é uma **fundação técnica transparente**, não uma ferramenta zootécnica validada. Esta página lista o que ele *não* garante.

## Não usar como recomendação genética real
- **Nenhum perfil, peso, direção, tolerância, bônus ou limiar foi validado por zootecnista ou geneticista.** Os perfis “Demonstração” são arbitrários.
- **Nenhum dataset real é distribuído.** Todo valor de demonstração é fictício (`DEMO / SYNTHETIC`) e não tem significado científico.
- O **score** é uma nota de adequação a um objetivo configurado; **não é** a previsão do valor genético do bezerro e não substitui a análise de um profissional.

## Hipóteses embutidas nos métodos
- **Percentil**: só é usado se o dataset declarar seu sentido (`ascending`/`descending`). O aplicativo nunca presume.
- **Escore z**: supõe distribuição aproximadamente normal e que a média/desvio fornecidos representem a população de referência da avaliação.
- **Faixa-alvo**: a tolerância é uma escolha de quem monta o perfil.
- **Complementaridade**: bônus linear pela melhora em pontos normalizados; é uma regra simples e parametrizada, sem fundamento genético formal.
- **Penalidade por acurácia**: linear abaixo de um limiar escolhido pelo usuário; a acurácia não ajusta o score de cada característica.

## Comparabilidade
- Dados de **programas, releases, bases genéticas, raças ou unidades diferentes nunca são convertidos**; tipos diferentes (DEP/EPD/EBV/PTA/índice/fenótipo) nunca se misturam.
- A **metodologia** de cada avaliação não é verificável pelo software. Quem publica ou importa a biblioteca deve declará-la; o usuário deve escolher o que comparar.
- A escolha da base de referência entre candidatos (a mais frequente) é uma regra de conveniência: ela é gravada, mas pode não ser a base que o especialista preferiria.

## O que não existe (ainda)
- Coeficientes de **parentesco e endogamia** (só alertas de ancestrais comuns); **progênie esperada**; **genômica** (haplótipos, recessivos, condições genéticas);
  índices econômicos; evolução genética do rebanho.
- Por isso o motor **não considera** risco de endogamia nem de doenças genéticas ao ordenar touros.

## Dados do usuário
- A qualidade do resultado depende da qualidade do pedigree, dos registros e dos datasets importados. O aplicativo valida o que consegue (sexos, datas, ciclos, formatos),
  mas não sabe se um valor importado está correto.
- Licenças: o aplicativo registra o que o usuário informa; **“disponível gratuitamente” não é licença**. A responsabilidade pelo uso de dados importados manualmente é do usuário.

## Técnicas
- Testado em Linux; builds de Windows e macOS não foram executados.
- Uma conexão SQLite (um escritor); backups sem criptografia.
