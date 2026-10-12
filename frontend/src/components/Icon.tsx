import type { SVGProps } from 'react'

export type IconName =
  | 'home' | 'herd' | 'tree' | 'vial' | 'pairing' | 'library' | 'calc' | 'report' | 'settings'

const paths: Record<IconName, string> = {
  home: 'M3 11.5 12 4l9 7.5M5.5 10v9.5h13V10',
  herd: 'M4 9c0-2 1.5-3 3-3h10c1.5 0 3 1 3 3v6a2 2 0 0 1-2 2h-1v3m-12 0v-3H6a2 2 0 0 1-2-2V9Zm5 3h6',
  tree: 'M12 4v6m0 0H6v4m6-4h6v4M6 14v-1m0 1a2 2 0 1 0 0 4 2 2 0 0 0 0-4Zm12 0a2 2 0 1 0 0 4 2 2 0 0 0 0-4ZM12 4a2 2 0 1 0 0-.01',
  vial: 'M9 3h6M10 3v6.5L5.5 18a2 2 0 0 0 1.8 3h9.4a2 2 0 0 0 1.8-3L14 9.5V3M8 15h8',
  pairing: 'M8 7a3 3 0 1 0 0-.01M16 17a3 3 0 1 0 0-.01M10.5 9.5l3 5M5 20c0-2 1.5-3 3-3m11-13c0 2-1.5 3-3 3',
  library: 'M5 4v16M9 4v16M13.5 5.5l3.6 14M4 20h16',
  calc: 'M6 3h12a1 1 0 0 1 1 1v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1Zm1 4h10M8.5 11.5h.01m3.49 0h.01m3.49 0h.01M8.5 15h.01m3.49 0h.01m3.49 0h.01',
  report: 'M7 3h7l4 4v14H7V3Zm7 0v4h4M9.5 12h5m-5 3.5h5',
  settings: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Zm7-3-.1-1.2 1.7-1.3-1.6-2.8-2 .7a6 6 0 0 0-1-.6L15.7 5h-3.2l-.3 1.8a6 6 0 0 0-1 .6l-2-.7-1.6 2.8 1.7 1.3L9 12l.1 1.2-1.7 1.3 1.6 2.8 2-.7c.3.2.7.4 1 .6l.3 1.8h3.2l.3-1.8c.4-.2.7-.4 1-.6l2 .7 1.6-2.8-1.7-1.3L19 12Z',
}

export function Icon({ name, ...rest }: { name: IconName } & SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      {...rest}
    >
      <path d={paths[name]} />
    </svg>
  )
}
