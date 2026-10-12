// Mirrors the DTOs returned by the Go backend (internal/service, internal/database).
export interface AppInfo {
  name: string
  version: string
  dataDir: string
  platform: string
}

export interface DbHealth {
  ok: boolean
  path: string
  schemaVersion: number
  journalMode: string
  foreignKeys: boolean
  quickCheck: string
  error?: string
}
