import type { LicenseStatus } from '../../types/domain'
import { licenseStatusLabel } from '../../utils/labels'

const cls: Record<LicenseStatus, string> = { unknown: 'badge-warn', no_redistribute: 'badge-muted', redistributable: 'badge-ok' }

export function LicenseBadge({ status, name }: { status: LicenseStatus; name?: string }) {
  return (
    <span className={`badge ${cls[status]}`} title={name ? `Licença: ${name}` : 'Nenhuma licença foi identificada para estes dados'}>
      {licenseStatusLabel[status]}
    </span>
  )
}
