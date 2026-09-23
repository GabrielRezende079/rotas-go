import { useState } from 'react'
import { Modal } from './Modal'

interface BaseNameModalProps {
  initialName: string
  onConfirm: (name: string) => void
  onClose: () => void
}

export function BaseNameModal({ initialName, onConfirm, onClose }: BaseNameModalProps) {
  const [name, setName] = useState(initialName)

  const confirmar = () => {
    const nome = name.trim()
    if (nome !== '') onConfirm(nome)
  }

  return (
    <Modal
      title="Nomear base"
      onClose={onClose}
      size="sm"
      footer={
        <>
          <button type="button" className="button button-secondary" onClick={onClose}>
            Cancelar
          </button>
          <button
            type="button"
            className="button button-primary"
            onClick={confirmar}
            disabled={name.trim() === ''}
          >
            Posicionar no mapa
          </button>
        </>
      }
    >
      <div className="modal-form">
        <div className="field">
          <label htmlFor="base-nome">Nome da base</label>
          <input
            id="base-nome"
            type="text"
            placeholder="Ex.: Depósito Centro"
            value={name}
            autoFocus
            onChange={(event) => setName(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') confirmar()
            }}
          />
        </div>
        <p className="hint">Depois de confirmar, clique no mapa para posicionar a base.</p>
      </div>
    </Modal>
  )
}