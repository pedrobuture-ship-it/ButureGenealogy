import { Component, type ErrorInfo, type ReactNode } from 'react'

interface State { error: Error | null }

/** Catches rendering errors so a bug in one page never leaves a blank window. */
export class ErrorBoundary extends Component<{ children: ReactNode; resetKey?: string }, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State { return { error } }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Erro de renderização', error, info.componentStack)
  }

  componentDidUpdate(prev: { resetKey?: string }) {
    if (this.state.error && prev.resetKey !== this.props.resetKey) this.setState({ error: null })
  }

  render() {
    if (!this.state.error) return this.props.children
    return (
      <div role="alert" className="notice notice-error">
        <strong>Algo deu errado ao exibir esta página.</strong>
        <span>Seus dados não foram afetados. {this.state.error.message}</span>
        <span><button className="btn" onClick={() => window.location.reload()}>Recarregar</button></span>
      </div>
    )
  }
}
