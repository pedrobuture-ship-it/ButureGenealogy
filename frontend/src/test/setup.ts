import '@testing-library/jest-dom/vitest'

// jsdom does not implement scrolling.
window.HTMLElement.prototype.scrollIntoView = function scrollIntoView() {}

// Node 26 ships an experimental global localStorage that shadows jsdom's; use a plain in-memory Storage.
function memoryStorage(): Storage {
  const m = new Map<string, string>()
  return {
    get length() { return m.size },
    clear: () => m.clear(),
    getItem: (k) => (m.has(k) ? m.get(k)! : null),
    key: (i) => [...m.keys()][i] ?? null,
    removeItem: (k) => { m.delete(k) },
    setItem: (k, v) => { m.set(k, String(v)) },
  }
}
Object.defineProperty(window, 'localStorage', { value: memoryStorage(), configurable: true })
