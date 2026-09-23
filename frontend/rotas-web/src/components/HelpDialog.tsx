import { Modal } from './Modal'

const ATALHOS: Array<[string, string]> = [
  ['Ctrl/⌘ + Enter', 'Calcular rotas'],
  ['Ctrl/⌘ + Shift + A', 'Calcular alternativas'],
  ['Ctrl/⌘ + L', 'Limpar tudo'],
  ['Esc', 'Cancelar alternativas ou fechar'],
  ['Ctrl/⌘ + 1', 'Aba Rotas'],
  ['Ctrl/⌘ + 2', 'Aba Veículos'],
]

interface HelpDialogProps {
  open: boolean
  onClose: () => void
}

export function HelpDialog({ open, onClose }: HelpDialogProps) {
  return (
    <Modal open={open} title="Atalhos de teclado" onClose={onClose} size="sm">
      <ul className="shortcut-list">
        {ATALHOS.map(([tecla, acao]) => (
          <li key={tecla}>
            <kbd className="shortcut-key">{tecla}</kbd>
            <span className="shortcut-acao">{acao}</span>
          </li>
        ))}
      </ul>
      <p className="hint">Dica: a cada veículo calculado, suas medidas somam na barra de resumo.</p>
    </Modal>
  )
}