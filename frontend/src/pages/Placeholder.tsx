import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { navItems } from '../layouts/navigation'

/** Temporary page for modules implemented in later stages. */
export function Placeholder({ path }: { path: string }) {
  const item = navItems.find((n) => n.path === path)
  return (
    <>
      <PageHeader title={item?.label ?? 'Página'} subtitle={item?.description} />
      <EmptyState title="Módulo ainda não implementado">
        Esta seção será construída em uma etapa posterior do MVP.
      </EmptyState>
    </>
  )
}
