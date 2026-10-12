export function EmptyState({ title, children }: { title: string; children?: React.ReactNode }) {
  return (
    <div className="empty-state">
      <h2>{title}</h2>
      {children && <p className="muted">{children}</p>}
    </div>
  )
}
