import { useState } from 'react'
import type { FormEvent } from 'react'
import {
  CATEGORIAS_VEICULO,
  STATUS_VEICULO,
  type CategoriaVeiculo,
  type NovoVeiculo,
  type StatusVeiculo,
  type Veiculo,
} from '../types'

interface VehiclesViewProps {
  veiculos: Veiculo[]
  onCadastrar: (req: NovoVeiculo) => Promise<void>
  onExcluir: (id: number) => Promise<void>
}

const FORM_INICIAL: NovoVeiculo = {
  modelo: '',
  categoria: 'Carro',
  placa: '',
  status: 'Disponível',
  kilometragem: 0,
  velocidade: 0,
}

function formatarNumero(valor: number): string {
  return valor.toLocaleString('pt-BR', { maximumFractionDigits: 1 })
}

function VehiclesView({ veiculos, onCadastrar, onExcluir }: VehiclesViewProps) {
  const [form, setForm] = useState<NovoVeiculo>(FORM_INICIAL)
  const [salvando, setSalvando] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const atualizar = (campo: keyof NovoVeiculo, valor: string | number) => {
    setForm((prev) => ({ ...prev, [campo]: valor }))
  }

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)
    setNotice(null)
    setSalvando(true)
    try {
      await onCadastrar({
        ...form,
        modelo: form.modelo.trim(),
        placa: form.placa.trim().toUpperCase(),
        kilometragem: Number(form.kilometragem),
        velocidade: Number(form.velocidade),
      })
      setForm(FORM_INICIAL)
      setNotice('Veículo cadastrado com sucesso.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao cadastrar o veículo')
    } finally {
      setSalvando(false)
    }
  }

  const handleExcluir = async (veiculo: Veiculo) => {
    if (!window.confirm(`Excluir o veículo ${veiculo.modelo} (${veiculo.placa})?`)) return
    setError(null)
    setNotice(null)
    try {
      await onExcluir(veiculo.id)
      setNotice('Veículo excluído.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir o veículo')
    }
  }

  const podeCadastrar = form.modelo.trim() !== '' && form.placa.trim() !== ''

  return (
    <div className="vehicles-view">
      <aside className="vehicles-sidebar">
        <header className="sidebar-header">
          <h1>Frota de veículos</h1>
          <p>Cadastro e consulta dos veículos da empresa, persistidos no banco de dados.</p>
        </header>

        <form className="vehicles-form" onSubmit={handleSubmit}>
          <div className="field">
            <label htmlFor="veh-modelo">Modelo</label>
            <input
              id="veh-modelo"
              type="text"
              placeholder="Ex.: Fiorino 2021"
              value={form.modelo}
              onChange={(event) => atualizar('modelo', event.target.value)}
            />
          </div>

          <div className="field">
            <label htmlFor="veh-categoria">Categoria</label>
            <select
              id="veh-categoria"
              value={form.categoria}
              onChange={(event) =>
                atualizar('categoria', event.target.value as CategoriaVeiculo)
              }
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

          <div className="vehicles-form-grid">
            <div className="field">
              <label htmlFor="veh-km">Kilometragem (km)</label>
              <input
                id="veh-km"
                type="number"
                min={0}
                step="0.1"
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
                value={form.velocidade}
                onChange={(event) => atualizar('velocidade', event.target.value)}
              />
            </div>
          </div>

          <button
            type="submit"
            className="button button-primary"
            disabled={salvando || !podeCadastrar}
          >
            {salvando ? 'Cadastrando…' : 'Cadastrar veículo'}
          </button>
        </form>

        {error !== null && <div className="error-box">{error}</div>}
        {notice !== null && <div className="notice-box">{notice}</div>}
      </aside>

      <main className="vehicles-list">
        <div className="section-title">
          <label>
            Frota cadastrada ({veiculos.length} {veiculos.length === 1 ? 'veículo' : 'veículos'})
          </label>
        </div>
        {veiculos.length === 0 ? (
          <p className="hint">Nenhum veículo cadastrado ainda. Use o formulário ao lado.</p>
        ) : (
          <table className="vehicles-table">
            <thead>
              <tr>
                <th>Modelo</th>
                <th>Categoria</th>
                <th>Placa</th>
                <th>Status</th>
                <th>Kilometragem</th>
                <th>Velocidade</th>
                <th>Cadastrado</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {veiculos.map((veiculo) => (
                <tr key={veiculo.id}>
                  <td>{veiculo.modelo}</td>
                  <td>{veiculo.categoria}</td>
                  <td className="veiculo-placa">{veiculo.placa}</td>
                  <td>
                    <span className={`status-badge status-${slugDoStatus(veiculo.status)}`}>
                      {veiculo.status}
                    </span>
                  </td>
                  <td>{formatarNumero(veiculo.kilometragem)} km</td>
                  <td>{formatarNumero(veiculo.velocidade)} km/h</td>
                  <td className="veiculo-data">{new Date(veiculo.created_at).toLocaleDateString('pt-BR')}</td>
                  <td>
                    <button
                      type="button"
                      className="button button-small button-secondary"
                      onClick={() => handleExcluir(veiculo)}
                    >
                      Excluir
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </main>
    </div>
  )
}

function slugDoStatus(status: StatusVeiculo): string {
  switch (status) {
    case 'Disponível':
      return 'disponivel'
    case 'Em uso':
      return 'em-uso'
    case 'Em Manutenção':
      return 'manutencao'
    case 'Indisponível':
      return 'indisponivel'
  }
}

export default VehiclesView