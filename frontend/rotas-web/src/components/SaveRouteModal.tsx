import { useState } from 'react'
import { Modal } from './Modal'
import { Spinner } from './Spinner'
import { useToasts } from '../toast'

interface SaveRouteModalProps {
  initialName: string
  totalVeiculos: number
  totalKm: number
  totalMin: number
  onConfirm: (name: string) => Promise<void>
  onClose: () => void
}

export function SaveRouteModal({
  initialName,
  totalVeiculos,
  totalKm,
  totalMin,
  onConfirm,
  onClose,
}: SaveRouteModalProps) {
  const toasts = useToasts()
  const [name, setName] = useState(initialName)
  const [salvando, setSalvando] = useState(false)

  const handleSubmit = async () => {
    const nome = name.trim()
    if (nome === '' || salvando) return
    setSalvando(true)
    try {
      await onConfirm(nome)
      toasts.success('Rota salva com sucesso.')
      onClose()
    } catch (err) {
      toasts.error(err instanceof Error ? err.message : 'Erro ao salvar a rota')
    } finally {
      setSalvando(false)
    }
  }

  return (
    <Modal
      title="Salvar rota"
      onClose={onClose}
      size="sm"
      footer={
        <>
          <button type="button" className="button button-secondary" onClick={onClose} disabled={salvando}>
            Cancelar
          </button>
          <button
            type="button"
            className="button button-primary"
            onClick={handleSubmit}
            disabled={salvando || name.trim() === ''}
          >
            {salvando ? (
              <>
                <Spinner size={14} /> Salvando…
              </>
            ) : (
              'Salvar'
            )}
          </button>
        </>
      }
    >
      <div className="modal-form">
        <div className="field">
          <label htmlFor="save-name">Nome da rota</label>
          <input
            id="save-name"
            type="text"
            placeholder="Ex.: entrega norte"
            value={name}
            autoFocus
            onChange={(event) => setName(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') void handleSubmit()
            }}
          />
        </div>
        <p className="save-summary">
          {totalVeiculos} {totalVeiculos === 1 ? 'veículo' : 'veículos'} · {totalKm.toFixed(1)} km ·{' '}
          {totalMin.toFixed(1)} min
        </p>
      </div>
    </Modal>
  )
}