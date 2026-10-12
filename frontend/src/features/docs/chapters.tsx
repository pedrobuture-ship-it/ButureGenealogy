import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { docPath, docSection } from '../mating/docLinks'
import { Example, Formula, type DocStatus } from './Formula'

export interface Chapter {
  id: string
  number: number
  title: string
  status: DocStatus
  /** One sentence shown in both modes. */
  lead: string
  /** Plain-language explanation for producers. */
  simple: ReactNode
  /** Formulas, parameters and limits for veterinarians, breeders and researchers. */
  technical: ReactNode
  /** Short badge text shown next to the title when it matters (e.g. the planned/unsupported notice). */
  notice?: string
}

const code = (s: string) => <code>{s}</code>

export const chapters: Chapter[] = [
  {
    id: docSection.score, number: 1, title: 'Score de acasalamento', status: 'implemented',
    lead: 'O score resume, em uma nota de 0 a 100, o quanto um touro atende ao objetivo definido para uma matriz.',
    simple: (
      <>
        <p>A nota é montada em três partes, sempre mostradas separadas em “Ver cálculo”:</p>
        <Formula>{`componentes ponderados
+ complementaridade
− penalidades
= score final`}</Formula>
        <p>Quais componentes existem depende dos <strong>datasets instalados</strong> e das <strong>configurações do perfil</strong> ativo: se você não escolher uma característica, ela não entra. Se algo não puder ser calculado com os dados disponíveis, o touro aparece como <strong>“não calculável”</strong>, e nunca com um número inventado.</p>
        <Example>
          <p>Soma dos componentes <strong>64,00</strong> · complementaridade <strong>+4,10</strong> · penalidades <strong>−2,30</strong> ⇒ score final <strong>65,80</strong>.</p>
          <p className="muted small">Valores ilustrativos; não são de nenhum animal real.</p>
        </Example>
      </>
    ),
    technical: (
      <>
        <p>Pipeline determinístico, implementado em {code('internal/engine')}, uma interface Go por etapa:</p>
        <Formula>{`Candidate → Validation → Hard Filters → Trait Compatibility → Normalization
          → Weighted Components → Complementarity → Penalties → Final Score`}</Formula>
        <Formula>{`base  = Σ score_i × peso_i              (peso normalizado entre os componentes calculáveis)
final = limitar(base + complementaridade − penalidades, 0, 100)`}</Formula>
        <ul>
          <li><strong>Cobertura</strong> = peso calculável ÷ peso total do perfil. Se for menor que {code('minCoverage')} (padrão 100%), o par fica <em>não calculável</em> e não recebe score.</li>
          <li>O valor sem limitação ({code('unclamped')}) é gravado; se houver corte em 0 ou 100, a tela avisa.</li>
          <li><strong>Ranking</strong>: score decrescente; empates em ordem alfabética do nome e depois do id (resultado idêntico independentemente da ordem de entrada). Não calculáveis e excluídos vêm depois, sem posição.</li>
          <li>Sem IA e sem aleatoriedade; o resultado não depende de data ou hora (a data só é registrada). Duas execuções com a mesma entrada produzem o mesmo resultado, byte a byte.</li>
        </ul>
        <p>Todos os parâmetros usados (pesos, direções, bônus máximo, limiares) vêm do perfil e são copiados para o snapshot. Ver <Link to={docPath(docSection.reproducibility)}>Reprodutibilidade</Link>.</p>
      </>
    ),
  },
  {
    id: docSection.weights, number: 2, title: 'Pesos', status: 'implemented',
    lead: 'O peso diz quanto cada característica importa para o seu objetivo.',
    simple: (
      <>
        <p>Cada característica recebe uma nota de 0 a 100 e um peso. A contribuição é:</p>
        <Formula>score_da_característica × peso</Formula>
        <p>Os pesos são <strong>relativos</strong>: 30, 20 e 10 valem o mesmo que 50%, 33% e 17%. Peso zero tira a característica do cálculo. Você pode editar todos os pesos na página Acasalamentos.</p>
        <Example>
          <p>Pesos 60 e 40; scores 80 e 70 ⇒ 0,60 × 80 + 0,40 × 70 = <strong>76,00</strong>.</p>
        </Example>
      </>
    ),
    technical: (
      <>
        <Formula>{`peso_normalizado_i = peso_bruto_i ÷ Σ peso_bruto (apenas componentes calculáveis)
contribuição_i     = score_i × peso_normalizado_i`}</Formula>
        <ul>
          <li>Se um componente não puder ser calculado, os pesos dos demais são renormalizados e a <strong>cobertura</strong> é mostrada. Com {code('minCoverage')} = 100% (padrão) isso nunca acontece: o par fica sem score.</li>
          <li>Pesos negativos ou não numéricos são rejeitados; características repetidas no perfil também.</li>
          <li>O peso bruto e o normalizado ficam gravados em cada componente, com a origem “perfil: peso X”.</li>
        </ul>
      </>
    ),
  },
  {
    id: docSection.normalization, number: 3, title: 'Normalização', status: 'implemented',
    lead: 'Valores em kg, cm², percentil ou índice não podem ser somados diretamente; a normalização os leva a uma mesma escala de 0 a 100.',
    simple: (
      <>
        <p>Somar kg com cm², ou um percentil com um índice, não faz sentido: são unidades e escalas diferentes. Por isso, antes de pesar, cada valor é convertido em uma nota de 0 a 100, de um jeito que o aplicativo consiga justificar.</p>
        <p>O MVP suporta <strong>apenas</strong> estes caminhos:</p>
        <ul>
          <li><strong>Percentil</strong>, quando o dataset diz o que o percentil significa.</li>
          <li><strong>Escore z</strong>, quando o dataset traz a média e o desvio-padrão da população.</li>
          <li><strong>Faixa-alvo</strong>, quando você define a faixa ideal e a tolerância.</li>
        </ul>
        <p>Se nenhum se aplica, a característica fica <strong>não calculável</strong>. O aplicativo não escolhe um método “por palpite”.</p>
      </>
    ),
    technical: (
      <>
        <p>Modo da normalização (perfil): {code('auto')} (percentil, senão escore z), {code('percentile')} ou {code('zscore')}. Implementação: {code('internal/engine/normalize.go')}.</p>
        <Formula label="Percentil">{`percentil (metadado percentileSemantics = ascending):
  MAXIMIZE → score = percentil            MINIMIZE → score = 100 − percentil
percentil (percentileSemantics = descending):
  MAXIMIZE → score = 100 − percentil      MINIMIZE → score = percentil`}</Formula>
        <Formula label="Escore z">{`z = (x − μ) ÷ σ
MAXIMIZE → score = 100 · Φ(z)           MINIMIZE → score = 100 · (1 − Φ(z))
(Φ = distribuição normal padrão; supõe distribuição aproximadamente normal)`}</Formula>
        <ul>
          <li>Percentil <strong>exige</strong> {code('percentileSemantics')} = {code('ascending')} ou {code('descending')} nos metadados da característica (declarado pelo dataset). Sem isso, o percentil <em>não é usado</em>: o significado nunca é presumido.</li>
          <li>Escore z <strong>exige</strong> {code('population.mean')} e {code('population.sd')} (σ &gt; 0) nos metadados.</li>
          <li>Faixa-alvo: ver <Link to={docPath(docSection.targetRange)}>capítulo 6</Link>.</li>
          <li>Qualquer método que não esteja nesta lista <strong>não existe</strong> no MVP.</li>
        </ul>
        <Example>
          <p>x = 12, μ = 10, σ = 2 ⇒ z = 1,00 ⇒ Φ(z) ≈ 0,8413 ⇒ score <strong>84,13</strong> (MAXIMIZE) ou <strong>15,87</strong> (MINIMIZE).</p>
        </Example>
      </>
    ),
  },
  {
    id: docSection.maximize, number: 4, title: 'MAXIMIZE', status: 'implemented',
    lead: 'Valores maiores são melhores: o score cresce com o valor.',
    simple: (
      <>
        <p>Usado quando “mais é melhor”, como em peso à desmama para quem busca crescimento. Um percentil alto vira uma nota alta.</p>
        <Example><p>Percentil 80 (ascendente) ⇒ score <strong>80,00</strong>.</p></Example>
      </>
    ),
    technical: (
      <>
        <p>A direção vem da definição da característica ou do perfil (o perfil prevalece). Características {code('INFORMATION_ONLY')} sem direção no perfil ficam <em>não calculáveis</em>: o aplicativo não decide por você o que é “melhor”. As características nativas são todas informativas por esse motivo.</p>
        <p>Fórmulas: capítulo <Link to={docPath(docSection.normalization)}>Normalização</Link>.</p>
      </>
    ),
  },
  {
    id: docSection.minimize, number: 5, title: 'MINIMIZE', status: 'implemented',
    lead: 'Valores menores são melhores: o score é invertido.',
    simple: (
      <>
        <p>Usado quando “menos é melhor”, por exemplo contagem de células somáticas. Um percentil baixo vira uma nota alta.</p>
        <Example><p>Percentil 30 (ascendente) ⇒ score <strong>70,00</strong>.</p></Example>
      </>
    ),
    technical: (
      <>
        <p>O sentido é invertido depois de interpretar o percentil (ou o escore z): {code('score = 100 − score_MAXIMIZE')}. A semântica do percentil continua sendo obrigatória.</p>
      </>
    ),
  },
  {
    id: docSection.targetRange, number: 6, title: 'TARGET_RANGE', status: 'implemented',
    lead: 'Nem mais nem menos: o melhor é ficar dentro de uma faixa.',
    simple: (
      <>
        <p>Usado quando existe um valor “ideal”, como peso ao nascer: muito baixo e muito alto são ruins. Dentro da faixa a nota é 100; fora, ela cai conforme a distância.</p>
        <p>A diferença para MAXIMIZE e MINIMIZE: aqui o melhor é um intervalo, não uma direção.</p>
      </>
    ),
    technical: (
      <>
        <Formula>{`distância = 0                  se mínimo ≤ x ≤ máximo
          mínimo − x           se x < mínimo
          x − máximo           se x > máximo
score = 100 · max(0, 1 − distância ÷ tolerância)`}</Formula>
        <ul>
          <li>Exige mínimo e máximo ideais (do perfil ou da definição da característica) e uma <strong>tolerância definida no perfil</strong>: a distância além da faixa em que a nota chega a 0. Sem tolerância, não é calculável: o aplicativo não inventa uma.</li>
          <li>Usa o valor, não o percentil; por isso a base de comparação (programa, release, unidade) continua obrigatória.</li>
        </ul>
        <Example><p>Faixa 30–40, tolerância 10: valor 35 ⇒ <strong>100</strong>; valor 43 (3 acima) ⇒ <strong>70</strong>; valor 55 ⇒ <strong>0</strong>.</p></Example>
      </>
    ),
  },
  {
    id: docSection.accuracy, number: 7, title: 'Acurácia', status: 'partial',
    lead: 'Mérito genético (o valor) e confiabilidade (a acurácia) são conceitos diferentes.',
    simple: (
      <>
        <p>Dois touros podem ter o mesmo valor genético estimado, mas um deles ter sido avaliado com muito mais informação (mais filhos, mais dados). A acurácia indica o quanto confiar no número.</p>
        <p><strong>No MVP a acurácia não altera a nota de cada característica.</strong> Ela aparece em todas as telas e pode ser usada de duas formas opcionais, ambas escolhidas por você no perfil:</p>
        <ul>
          <li>como <strong>filtro</strong>: exclui touros abaixo de uma acurácia mínima (acurácia desconhecida também exclui);</li>
          <li>como <strong>penalidade</strong>: tira pontos quando a acurácia fica abaixo de um limiar.</li>
        </ul>
      </>
    ),
    technical: (
      <>
        <p>Acurácia é guardada como fração (0–1). Uso exato no motor:</p>
        <Formula label="Penalidade por baixa acurácia">{`falta_t   = max(0, (limiar − acurácia_t) ÷ limiar)
pontos    = máximo × Σ peso_normalizado_t × falta_t`}</Formula>
        <ul>
          <li>Os parâmetros {code('limiar')} e {code('máximo')} vêm do perfil (desativada com 0). Acurácia <em>desconhecida</em> não é penalizada (desconhecido ≠ baixo) e gera aviso.</li>
          <li>Filtro de acurácia mínima: excluir se a melhor acurácia disponível da característica for menor que o mínimo; se nenhuma for informada, exclui (não dá para verificar).</li>
          <li>Não há ponderação do score pela acurácia (nem “encolhimento” bayesiano). Isso seria um modelo adicional que o MVP não implementa.</li>
        </ul>
        <Example><p>Limiar 60%, máximo 10 pontos; acurácia 30% numa característica com peso 0,60 ⇒ falta 0,5 ⇒ 10 × 0,60 × 0,5 = <strong>3,00</strong> pontos.</p></Example>
      </>
    ),
  },
  {
    id: docSection.complementarity, number: 8, title: 'Complementaridade', status: 'implemented',
    lead: 'O objetivo não é buscar “o maior touro”, e sim o que melhor complementa determinada matriz.',
    simple: (
      <>
        <p>Um touro excelente em uma característica em que a matriz já é excelente acrescenta pouco. Já um touro que melhora justamente o ponto em que a matriz é mais fraca acrescenta mais. A complementaridade é um <strong>bônus</strong> em pontos por essa melhora, até um máximo que você define no perfil (0 desliga).</p>
        <p>Só conta onde a matriz tem avaliação <strong>comparável</strong> (mesma base). Onde ela não tem, o bônus daquela característica é zero e o aplicativo avisa.</p>
        <Example><p>Máximo 10 pontos; pesos 0,6 e 0,4; ganhos 0,4 e 0,1 ⇒ 10 × (0,6 × 0,4 + 0,4 × 0,1) = <strong>+2,80</strong>.</p></Example>
      </>
    ),
    technical: (
      <>
        <Formula>{`ganho_t = limitar((score_touro_t − score_matriz_t) ÷ 100, 0, 1)
pontos  = bônus_máximo × Σ peso_normalizado_t × ganho_t`}</Formula>
        <ul>
          <li>Os dois scores usam o <strong>mesmo método e a mesma base</strong>. Se a matriz foi avaliada em outra base: “⚠ Comparação genética não suportada entre essas bases” e o item não é calculado.</li>
          <li>O bônus nunca é negativo: matriz melhor que o touro dá ganho 0, não penalidade.</li>
          <li>O denominador é o total de pesos calculáveis, então itens sem comparação contribuem 0: o bônus só é dado onde está demonstrado.</li>
          <li>Esta é uma <strong>regra de configuração do MVP</strong> (um método simples e auditável), não um modelo zootécnico validado. O máximo é um parâmetro do perfil.</li>
        </ul>
      </>
    ),
  },
  {
    id: docSection.pedigree, number: 9, title: 'Pedigree', status: 'implemented',
    lead: 'Algoritmos de genealogia que existem de fato no aplicativo.',
    simple: (
      <>
        <p>O aplicativo mostra a árvore de 2 a 5 gerações, destaca ancestrais que aparecem mais de uma vez, avisa de inconsistências (por exemplo, um pai fêmea) e mostra o quanto do pedigree é conhecido.</p>
        <p>No acasalamento, ele <strong>lista ancestrais em comum</strong> entre a matriz e cada touro como alerta e, se você quiser, <strong>exclui</strong> touros que são ancestrais da matriz.</p>
      </>
    ),
    technical: (
      <>
        <p>Implementados e testados em {code('internal/genetics')}:</p>
        <ul>
          <li><strong>Coleta de ancestrais</strong> em largura, com caminho (S = pai, D = mãe), limitada a N gerações (não entra em laço mesmo com dados corrompidos).</li>
          <li><strong>Ancestrais comuns</strong> de dois animais, incluindo o próprio animal como geração 0 (assim, um pai e seu filho têm o pai como ancestral comum).</li>
          <li><strong>Detecção de ciclos</strong> (busca em profundidade) e <strong>bloqueio de ciclos</strong> ao gravar.</li>
          <li><strong>Inconsistências</strong>: o próprio animal como ascendente; pai e mãe iguais; pai que não é macho; mãe que não é fêmea; ascendente nascido depois (ou no mesmo dia) do descendente; ciclo.</li>
          <li><strong>Profundidade e completude</strong> por geração (conhecidos ÷ 2<sup>g</sup>).</li>
        </ul>
        <p>No motor: ancestrais comuns viram avisos; {code('excludeDirectAncestors')} é um filtro rígido. Nenhum deles altera o score.</p>
      </>
    ),
  },
  {
    id: docSection.relationship, number: 10, title: 'Parentesco / endogamia', status: 'planned',
    notice: 'Planejado — ainda não utilizado no score do MVP',
    lead: 'Coeficientes formais de parentesco e de endogamia ainda não são calculados.',
    simple: (
      <>
        <p className="doc-planned"><strong>Planejado — ainda não utilizado no score do MVP.</strong></p>
        <p>Acasalar animais aparentados aumenta a endogamia. Hoje o aplicativo só avisa quando encontra ancestrais em comum, sem dar um número. Ele <strong>não</strong> mostra nenhum coeficiente nem desconta pontos por parentesco, porque uma fórmula simplificada passaria uma falsa impressão de precisão.</p>
      </>
    ),
    technical: (
      <>
        <p className="doc-planned"><strong>Planejado — ainda não utilizado no score do MVP.</strong></p>
        <ul>
          <li>Na tela do pedigree aparece “não calculado nesta versão”; no “Ver cálculo”, a penalidade de parentesco/endogamia aparece como <em>não aplicada</em>, com 0 ponto.</li>
          <li>O método formal planejado é a matriz de parentesco aditivo (algoritmo tabular de Henderson) com o coeficiente de endogamia de Wright. <strong>Não está implementado</strong>; só será exibido depois de validado contra casos de referência publicados.</li>
          <li>Interface de extensão pronta: {code('genetics.RelationshipCalculator')} (implementação atual: nunca devolve valor).</li>
        </ul>
      </>
    ),
  },
  {
    id: docSection.progeny, number: 11, title: 'Progênie esperada', status: 'not_implemented',
    notice: 'Não implementado no MVP',
    lead: 'O MVP não calcula o valor genético esperado do bezerro.',
    simple: (
      <>
        <p className="doc-planned"><strong>Não implementado no MVP.</strong></p>
        <p>O score <strong>não é</strong> a previsão da DEP, do EBV ou da PTA do filho. Ele é uma nota de adequação ao seu objetivo, baseada nos valores dos touros e na comparação com a matriz.</p>
        <p>DEP, EPD, EBV e PTA não devem ser tratados como se fossem a mesma escala: cada um depende do tipo de valor e da base do programa que o publica.</p>
      </>
    ),
    technical: (
      <>
        <ul>
          <li>Nenhuma média dos pais ou expectativa para a progênie é calculada em lugar algum do motor.</li>
          <li>Fazer isso corretamente exige conhecer o tipo (DEP/EPD/PTA são diferenças esperadas; EBV é o valor genético completo), a escala em que o programa o expressa e a mesma base genética entre pai e mãe.</li>
          <li>O motor já recusa misturar tipos: um EBV nunca substitui um DEP, e bases diferentes nunca são convertidas (capítulo <Link to={docPath(docSection.datasets)}>Datasets</Link>).</li>
          <li>O cálculo só será documentado aqui quando existir, com metodologia validada por especialistas (ver roteiro pós-MVP).</li>
        </ul>
      </>
    ),
  },
  {
    id: docSection.datasets, number: 12, title: 'Datasets', status: 'implemented',
    lead: 'Dataset A ≠ Dataset B: valores de bases diferentes não são equivalentes.',
    simple: (
      <>
        <Formula>{`Dataset A
≠
Dataset B`}</Formula>
        <p>Dois programas de avaliação podem usar populações de referência e bases genéticas diferentes; um 80 de um programa não é um 80 do outro. O aplicativo só compara valores que sejam comparáveis e, quando não são, mostra:</p>
        <p className="warn-line">⚠ Comparação genética não suportada entre essas bases.</p>
        <p>A comparabilidade depende de: <strong>raça, programa, release, base genética, característica, unidade e metodologia</strong>.</p>
      </>
    ),
    technical: (
      <>
        <p>Duas avaliações são comparáveis quando todo o conjunto abaixo é igual ({code('engine.Basis')}):</p>
        <Formula>{`característica · tipo (DEP/EPD/EBV/PTA/INDEX) · programa · release · base genética · raça · unidade`}</Formula>
        <ul>
          <li><strong>Base de referência</strong>: para cada característica o motor escolhe uma única base para comparar os touros: a que tem mais candidatos avaliados; empate: a em que a matriz também foi avaliada; novo empate: ordem alfabética da chave. A escolha e o motivo ficam gravados.</li>
          <li>Touros avaliados em outra base ficam <em>não calculáveis</em> naquela característica; nada é convertido.</li>
          <li>Se o mesmo animal tiver várias avaliações na mesma base (por exemplo, duas versões instaladas do mesmo dataset), usa-se a instalada mais recentemente e há aviso.</li>
          <li><strong>Metodologia</strong> não pode ser verificada automaticamente: é responsabilidade de quem publica ou importa a biblioteca declará-la (README) e do usuário escolher o que comparar.</li>
        </ul>
      </>
    ),
  },
  {
    id: docSection.reproducibility, number: 13, title: 'Reprodutibilidade', status: 'implemented',
    lead: 'Todo resultado guarda tudo o que é preciso para explicá-lo, mesmo depois que os dados mudarem.',
    simple: (
      <>
        <p>Cada cálculo salvo guarda um <strong>snapshot</strong>: a versão do motor, as versões dos datasets, os valores usados, o perfil com todos os pesos e o resultado de cada etapa. Se você atualizar ou desativar um dataset depois, o cálculo antigo <strong>não muda</strong>. Versões usadas em cálculos salvos não podem ser removidas.</p>
      </>
    ),
    technical: (
      <>
        <ul>
          <li><strong>Versão do motor</strong>: {code('engine.Version')}; qualquer mudança de comportamento exige novo número.</li>
          <li><strong>Datasets</strong>: id, nome, versão, programa, release e SHA-256 do pacote de cada biblioteca realmente usada.</li>
          <li><strong>Snapshot</strong> (JSON versionado, um por par matriz × touro): matriz e touro, todas as avaliações das características do perfil (inclusive as de bases não usadas), perfil completo, bases de referência, resultado detalhado e avisos.</li>
          <li><strong>Configurações</strong>: o perfil é copiado inteiro; editar ou apagar o perfil depois não afeta o histórico.</li>
          <li>O histórico é lido do snapshot salvo; nada é recalculado ao abrir “Ver cálculo”.</li>
        </ul>
        <p>Um cálculo novo com outra versão do motor pode dar resultado diferente do antigo; por isso a versão fica sempre registrada.</p>
      </>
    ),
  },
]
