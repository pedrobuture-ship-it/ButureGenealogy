export function DemoBadge() {
  return <span className="badge badge-demo" title="DEMO / SYNTHETIC: dado fictício, não representa um animal real">DEMO</span>
}

const statusClass: Record<string, string> = {
  active: 'badge-ok', archived: 'badge-muted', sold: 'badge-info', deceased: 'badge-muted',
}

export function StatusBadge({ status, label }: { status: string; label: string }) {
  return <span className={`badge ${statusClass[status] ?? ''}`}>{label}</span>
}
