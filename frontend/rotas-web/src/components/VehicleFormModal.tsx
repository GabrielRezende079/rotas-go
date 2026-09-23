import { useState } from 'react'
import type { FormEvent } from 'react'
import {
  CATEGORIAS_VEICULO,
  STATUS_VEICULO,
  type CategoriaVeiculo,
  type NovoVeiculo,
  type StatusVeiculo,
} from '../types'
import { Modal } from './Modal'
import { Spinner } from './Spinner'
import { useToasts } from '../toast'

interface VehicleFormModalProps {
  onCadastrar: (req: NovoVeiculo) => Promise<void>
  onClose: () => void
}

const FORM_INICIAL: NovoVeiculo = {
  modelo: '',
  categoria: 'Carro',
  placa: '',
  status: 'Disponível',
  kilometragem: 0,
  velocidade: 0,
}

export function VehicleFormModal({ onCadastrar, onClose }: VehicleFormModalProps) {
  const toasts = useToasts()
  const [form, setForm] = useState<NovoVeiculo>(FORM_INICIAL)
  const [salvando, setSalvando] = useState(false)

  const atualizar = (campo: keyof NovoVeiculo, valor: string | number) => {
    setForm((prev) => ({ ...prev, [campo]: valor }))
  }

  const handleSubmit = async (event?: FormEvent) => {
    event?.preventDefault()
    setSalvando(true)
    try {
      await onCadastrar({
        ...form,
        modelo: form.modelo.trim(),
        placa: form.placa.trim().toUpperCase(),
        kilometragem: Number(form.kilometragem) || 0,
        velocidade: Number(form.velocidade) || 0,
      })
      toasts.success('Veículo cadastrado com sucesso.')
      onClose()
    } catch (err) {
      toasts.error(err instanceof Error ? err.message : 'Erro ao cadastrar o veículo')
    } finally {
      setSalvando(false)
    }
  }

  const podeCadastrar = form.modelo.trim() !== '' && form.placa.trim() !== ''

  return (
    <Modal
      title="Cadastrar veículo"
      onClose={onClose}
      size="md"
      footer={
        <>
          <button
            type="button"
            className="button button-secondary"
            onClick={onClose}
            disabled={salvando}
          >
            Cancelar
          </button>
          <button
            type="button"
            className="button button-primary"
            onClick={() => void handleSubmit()}
            disabled={salvando || !podeCadastrar}
          >
            {salvando ? (
              <>
                <Spinner size={14} /> Cadastrando…
              </>
            ) : (
              'Cadastrar veículo'
            )}
          </button>
        </>
      }
    >
      <form className="modal-form modal-form-grid" onSubmit={(event) => void handleSubmit(event)}>
        <div className="field">
          <label htmlFor="veh-modelo">Modelo</label>
          <input
            id="veh-modelo"
            type="text"
            placeholder="Ex.: Fiorino 2021"
            value={form.modelo}
            autoFocus
            onChange={(event) => atualizar('modelo', event.target.value)}
          />
        </div>

        <div className="field">
          <label htmlFor="veh-categoria">Categoria</label>
          <select
            id="veh-categoria"
            value={form.categoria}
            onChange={(event) => atualizar('categoria', event.target.value as CategoriaVeiculo)}
          >
            {CATEGORIAS_VEICULO.map((categoria) => (
              <option key={categoria} value={categoria}>
                {categoria}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label htmlFor="veh-placa">Placa</label>
          <input
            id="veh-placa"
            type="text"
            placeholder="ABC1D23 ou ABC-1234"
            value={form.placa}
            onChange={(event) => atualizar('placa', event.target.value.toUpperCase())}
          />
        </div>

        <div className="field">
          <label htmlFor="veh-status">Status</label>
          <select
            id="veh-status"
            value={form.status}
            onChange={(event) => atualizar('status', event.target.value as StatusVeiculo)}
          >
            {STATUS_VEICULO.map((status) => (
              <option key={status} value={status}>
                {status}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label htmlFor="veh-km">Kilometragem (km)</label>
          <input
            id="veh-km"
            type="number"
            min={0}
            step="0.1"
            placeholder="0"
            value={form.kilometragem}
            onChange={(event) => atualizar('kilometragem', event.target.value)}
          />
        </div>

        <div className="field">
          <label htmlFor="veh-vel">Velocidade (km/h)</label>
          <input
            id="veh-vel"
            type="number"
            min={0}
            step="1"
            placeholder="0"
            value={form.velocidade}
            onChange={(event) => atualizar('velocidade', event.target.value)}
          />
        </div>
      </form>
    </Modal>
  )
}