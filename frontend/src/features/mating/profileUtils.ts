import type { Profile, TraitWeight } from '../../types/engine'

export const newTraitWeight = (code: string, evaluationType: string): TraitWeight => ({
  traitCode: code, evaluationType, weight: 10, direction: '', idealMin: null, idealMax: null, targetTolerance: null, required: false,
})

/** Percentage input ("60" or "60,5") to a fraction, or null when blank/invalid. */
export function percentToFraction(text: string): number | null {
  const t = text.trim().replace(',', '.')
  if (t === '') return null
  const n = Number(t)
  return Number.isFinite(n) ? n / 100 : null
}

export function fractionToPercentText(f: number | null): string {
  if (f == null) return ''
  return String(Math.round(f * 10000) / 100).replace('.', ',')
}

export function numberOrNull(text: string): number | null {
  const t = text.trim().replace(',', '.')
  if (t === '') return null
  const n = Number(t)
  return Number.isFinite(n) ? n : null
}

/** Nominal share of each weight (display only; the engine normalizes over what is calculable). */
export function nominalShares(traits: TraitWeight[]): number[] {
  const total = traits.reduce((t, x) => t + (x.weight > 0 ? x.weight : 0), 0)
  return traits.map((x) => (total > 0 && x.weight > 0 ? x.weight / total : 0))
}

export const traitKey = (t: { traitCode: string; evaluationType: string }) => `${t.traitCode}|${t.evaluationType}`

/** A copy to edit; the original (built-in) profile stays untouched. */
export const cloneProfile = (p: Profile): Profile => JSON.parse(JSON.stringify(p)) as Profile

/** Key of a catalogue trait (the catalogue uses `code`, the profile uses `traitCode`). */
export const catalogKey = (t: { code: string; evaluationType: string }) => `${t.code}|${t.evaluationType}`
