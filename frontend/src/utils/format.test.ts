import { formatAge, formatDate, localToday } from './format'

describe('format', () => {
  it('data de hoje é local, não UTC', () => {
    // 22:30 in UTC-3 is already the next day in UTC
    const d = new Date(2026, 9, 11, 22, 30) // local time
    expect(localToday(d)).toBe('2026-10-11')
    expect(localToday(new Date(2026, 0, 5, 0, 5))).toBe('2026-01-05')
  })

  it('formata datas e idades aproximadas', () => {
    expect(formatDate('2026-10-11')).toBe('11/10/2026')
    expect(formatDate('')).toBe('—')
    const now = new Date(2026, 9, 11)
    expect(formatAge('2023-08-11', now)).toBe('3 a 2 m')
    expect(formatAge('2026-04-11', now)).toBe('6 m')
    expect(formatAge('2025-10-11', now)).toBe('1 a')
    expect(formatAge('', now)).toBe('—')
    expect(formatAge('2027-01-01', now)).toBe('—')
  })
})
