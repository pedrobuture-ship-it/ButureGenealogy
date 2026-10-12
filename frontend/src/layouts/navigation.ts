import type { IconName } from '../components/Icon'

export interface NavItem {
  path: string
  label: string
  icon: IconName
  description: string
}

export const navItems: NavItem[] = [
  { path: '/', label: 'Início', icon: 'home', description: 'Visão geral da propriedade e do estado do sistema.' },
  { path: '/rebanho', label: 'Rebanho', icon: 'herd', description: 'Cadastro e consulta dos animais.' },
  { path: '/genealogia', label: 'Genealogia', icon: 'tree', description: 'Pedigree e ancestrais.' },
  { path: '/semen', label: 'Sêmen', icon: 'vial', description: 'Touros e estoque de doses.' },
  { path: '/acasalamentos', label: 'Acasalamentos', icon: 'pairing', description: 'Ranking de touros por matriz.' },
  { path: '/bibliotecas', label: 'Bibliotecas', icon: 'library', description: 'Datasets e bases de avaliações instaláveis.' },
  { path: '/calculos', label: 'Cálculos', icon: 'calc', description: 'Documentação dos métodos usados pelo motor.' },
  { path: '/relatorios', label: 'Relatórios', icon: 'report', description: 'Relatórios com proveniência dos dados.' },
  { path: '/configuracoes', label: 'Configurações', icon: 'settings', description: 'Preferências, backup e integridade.' },
]
