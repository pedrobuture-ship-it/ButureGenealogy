import { useRef, type KeyboardEvent } from 'react'

export interface TabDef { id: string; label: string }

interface Props {
  tabs: TabDef[]
  active: string
  onChange: (id: string) => void
  label: string
}

export function Tabs({ tabs, active, onChange, label }: Props) {
  const refs = useRef<Record<string, HTMLButtonElement | null>>({})

  function onKey(e: KeyboardEvent, i: number) {
    let next = i
    if (e.key === 'ArrowRight') next = (i + 1) % tabs.length
    else if (e.key === 'ArrowLeft') next = (i - 1 + tabs.length) % tabs.length
    else return
    e.preventDefault()
    onChange(tabs[next].id)
    refs.current[tabs[next].id]?.focus()
  }

  return (
    <div role="tablist" aria-label={label} className="tabs">
      {tabs.map((t, i) => (
        <button
          key={t.id}
          role="tab"
          id={`tab-${t.id}`}
          aria-selected={t.id === active}
          aria-controls={`panel-${t.id}`}
          tabIndex={t.id === active ? 0 : -1}
          className={t.id === active ? 'tab active' : 'tab'}
          ref={(el) => { refs.current[t.id] = el }}
          onClick={() => onChange(t.id)}
          onKeyDown={(e) => onKey(e, i)}
        >
          {t.label}
        </button>
      ))}
    </div>
  )
}
