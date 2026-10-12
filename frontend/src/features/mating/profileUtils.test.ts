import { fractionToPercentText, newTraitWeight, nominalShares, numberOrNull, percentToFraction } from './profileUtils'

describe('profileUtils', () => {
  it('converte percentuais digitados em frações e vice-versa', () => {
    expect(percentToFraction('60')).toBe(0.6)
    expect(percentToFraction('62,5')).toBe(0.625)
    expect(percentToFraction('')).toBeNull()
    expect(percentToFraction('abc')).toBeNull()
    expect(fractionToPercentText(0.6)).toBe('60')
    expect(fractionToPercentText(0.625)).toBe('62,5')
    expect(fractionToPercentText(null)).toBe('')
    expect(numberOrNull('3,5')).toBe(3.5)
    expect(numberOrNull(' ')).toBeNull()
  })

  it('mostra a fração nominal dos pesos, ignorando pesos zerados', () => {
    const a = newTraitWeight('WW', 'DEP'); a.weight = 30
    const b = newTraitWeight('YW', 'DEP'); b.weight = 10
    const c = newTraitWeight('MILK', 'DEP'); c.weight = 0
    expect(nominalShares([a, b, c])).toEqual([0.75, 0.25, 0])
    expect(nominalShares([c])).toEqual([0])
  })

  it('uma característica nova não escolhe direção por conta própria', () => {
    expect(newTraitWeight('WW', 'DEP').direction).toBe('')
  })
})
