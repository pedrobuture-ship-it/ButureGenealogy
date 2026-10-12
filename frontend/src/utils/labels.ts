// Presentation labels for backend enums.
export const sexLabel: Record<string, string> = { male: 'Macho', female: 'Fêmea' }

export const categoryLabel: Record<string, string> = {
  calf: 'Bezerro(a)', heifer: 'Novilha', cow: 'Vaca', steer: 'Boi', bull: 'Touro', other: 'Outra',
}

export const statusLabel: Record<string, string> = {
  active: 'Ativo', archived: 'Arquivado', sold: 'Vendido', deceased: 'Morto',
}

export const categoriesForSex = (sex: string): string[] =>
  sex === 'male' ? ['calf', 'steer', 'bull', 'other'] : sex === 'female' ? ['calf', 'heifer', 'cow', 'other'] : Object.keys(categoryLabel)

export const traitCategoryLabel: Record<string, string> = {
  growth: 'Crescimento', reproduction: 'Reprodução', maternal: 'Materna', carcass: 'Carcaça', dairy: 'Leite',
  conformation: 'Conformação', health: 'Saúde', temperament: 'Temperamento', efficiency: 'Eficiência', other: 'Outras',
}

export const valueKindLabel: Record<string, string> = {
  PHENOTYPE: 'Fenótipo', DEP: 'DEP', EPD: 'EPD', EBV: 'EBV', PTA: 'PTA', INDEX: 'Índice',
}

export const directionLabel: Record<string, string> = {
  MAXIMIZE: 'Maximizar', MINIMIZE: 'Minimizar', TARGET_RANGE: 'Faixa-alvo', INFORMATION_ONLY: 'Apenas informativa',
}

export const purposeLabel: Record<string, string> = { beef: 'Corte', dairy: 'Leite', dual: 'Dupla aptidão', any: 'Qualquer' }

export const licenseStatusLabel: Record<string, string> = {
  unknown: 'Licença desconhecida',
  no_redistribute: 'Sem redistribuição',
  redistributable: 'Redistribuível',
}

export const stockStatusLabel: Record<string, string> = { in_stock: 'Em estoque', depleted: 'Esgotado', no_semen: 'Sem sêmen' }
export const movementTypeLabel: Record<string, string> = { entry: 'Entrada', withdrawal: 'Retirada', adjustment: 'Ajuste' }
export const entryReasonLabel: Record<string, string> = { purchase: 'Compra', donation: 'Doação', transfer: 'Transferência', other: 'Outro' }
export const withdrawalReasonLabel: Record<string, string> = { insemination: 'Inseminação', sale: 'Venda', loss: 'Perda / descarte', transfer: 'Transferência', other: 'Outro' }
export const reasonLabel = (r: string): string =>
  ({ ...entryReasonLabel, ...withdrawalReasonLabel, count_correction: 'Correção de contagem' } as Record<string, string>)[r] ?? r
