import type { TreeNode } from '../../types/domain'
import { sexLabel } from '../../utils/labels'

interface BranchProps {
  repeatedIds: Set<string>
  onOpen: (id: string) => void
  onCreate: (placeholder: TreeNode) => void
}
interface Props extends BranchProps {
  root: TreeNode
}

/**
 * Horizontal pedigree: the animal on the left, ancestors to the right.
 * Unknown ancestors collapse into a single compact placeholder, so sparse
 * pedigrees stay small; deep ones scroll inside the container.
 */
export function PedigreeTree({ root, repeatedIds, onOpen, onCreate }: Props) {
  return (
    <div className="ped-scroll" role="tree" aria-label="Árvore genealógica">
      <Branch node={root} repeatedIds={repeatedIds} onOpen={onOpen} onCreate={onCreate} />
    </div>
  )
}

function Branch({ node, repeatedIds, onOpen, onCreate }: BranchProps & { node: TreeNode }) {
  const hasParents = node.sire !== null || node.dam !== null
  return (
    <div className="ped-branch" role="group">
      <div className={hasParents ? 'ped-node-cell has-parents' : 'ped-node-cell'}>
        <NodeCard node={node} repeated={node.known && repeatedIds.has(node.id)} onOpen={onOpen} onCreate={onCreate} />
      </div>
      {hasParents && (
        <div className="ped-parents">
          {node.sire && <Branch node={node.sire} repeatedIds={repeatedIds} onOpen={onOpen} onCreate={onCreate} />}
          {node.dam && <Branch node={node.dam} repeatedIds={repeatedIds} onOpen={onOpen} onCreate={onCreate} />}
        </div>
      )}
    </div>
  )
}

function NodeCard({ node, repeated, onOpen, onCreate }: { node: TreeNode; repeated: boolean; onOpen: (id: string) => void; onCreate: (n: TreeNode) => void }) {
  if (!node.known) {
    const role = node.role === 'sire' ? 'pai' : 'mãe'
    const of = node.childName ? ` de ${node.childName}` : ''
    return (
      <button className="ped-node ped-unknown" role="treeitem" onClick={() => onCreate(node)} aria-label={`Cadastrar ${role}${of}`}>
        <span className="ped-unknown-label">{node.role === 'sire' ? 'Pai' : 'Mãe'}{of} não informado(a)</span>
        <span className="ped-unknown-action">+ Cadastrar</span>
      </button>
    )
  }
  const title = node.name || node.registration || 'Sem identificação'
  const sexClass = node.sex === 'male' ? 'ped-male' : node.sex === 'female' ? 'ped-female' : ''
  const isSelf = node.role === 'self'
  return (
    <button
      className={`ped-node ${sexClass} ${isSelf ? 'ped-self' : ''} ${repeated ? 'ped-repeated' : ''}`}
      role="treeitem"
      onClick={() => !isSelf && onOpen(node.id)}
      aria-current={isSelf ? 'true' : undefined}
      title={isSelf ? 'Animal atual' : 'Abrir ficha'}
    >
      <span className="ped-name">{title}</span>
      <span className="ped-meta">
        {node.sex ? sexLabel[node.sex] : ''}{node.breedName ? ` · ${node.breedName}` : ''}
      </span>
      {node.registration && node.name && <span className="ped-meta mono">{node.registration}</span>}
      {repeated && <span className="ped-flag">ancestral repetido</span>}
      {node.hasMore && <span className="ped-more" title="Há ancestrais além das gerações exibidas">+ mais ancestrais</span>}
    </button>
  )
}
