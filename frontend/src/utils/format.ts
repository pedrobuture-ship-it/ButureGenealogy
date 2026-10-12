export function formatDate(iso: string): string {
  if (!iso) return '—'
  const [y, m, d] = iso.slice(0, 10).split('-')
  return d && m && y ? `${d}/${m}/${y}` : iso
}

/** Approximate age for display, e.g. "3 a 2 m" or "8 m". */
export function formatAge(birthDate: string, now: Date = new Date()): string {
  if (!birthDate) return '—'
  const b = new Date(birthDate + 'T00:00:00')
  if (Number.isNaN(b.getTime())) return '—'
  let months = (now.getFullYear() - b.getFullYear()) * 12 + (now.getMonth() - b.getMonth())
  if (now.getDate() < b.getDate()) months -= 1
  if (months < 0) return '—'
  const years = Math.floor(months / 12)
  const rest = months % 12
  if (years === 0) return `${rest} m`
  return rest === 0 ? `${years} a` : `${years} a ${rest} m`
}

export function formatNumber(n: number, digits = 2): string {
  return n.toLocaleString('pt-BR', { maximumFractionDigits: digits })
}

export function animalTitle(a: { name: string; registration: string }): string {
  return a.name || a.registration || 'Sem identificação'
}

export function formatMoney(n: number | null): string {
  return n == null ? '—' : n.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
}

const minus = '−'

/** Two decimals with a decimal comma, e.g. 87,43. Display only: never used to compute. */
export function fmt2(n: number | null | undefined): string {
  return n == null ? '—' : n.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

/** Signed two-decimal number: +4,10 / −2,30. */
export function signed2(n: number): string {
  const s = fmt2(Math.abs(n))
  return n < 0 ? `${minus}${s}` : `+${s}`
}

export function fmtPercent(fraction: number | null | undefined, digits = 0): string {
  return fraction == null ? '—' : `${(fraction * 100).toLocaleString('pt-BR', { maximumFractionDigits: digits })}%`
}

/** Today's date in the user's own time zone, as YYYY-MM-DD (never UTC: after 21h in Brazil UTC is already "tomorrow"). */
export function localToday(now: Date = new Date()): string {
  const m = String(now.getMonth() + 1).padStart(2, '0')
  const d = String(now.getDate()).padStart(2, '0')
  return `${now.getFullYear()}-${m}-${d}`
}
