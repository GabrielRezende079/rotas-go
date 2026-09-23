import { useMemo, useState } from 'react'
import type { StatusVeiculo, Veiculo } from '../types'
import { Icon } from './Icon'

interface VehiclesViewProps {
  veiculos: Veiculo[]
  onNovoVeiculo: () => void
  onExcluir: (veiculo: Veiculo) => void
}

interface Kpi {
  chave: StatusVeiculo | 'total'
  rotulo: string
  contagem: number
  sufixo: string
}

function formatarNumero(valor: number): string {
  return valor.toLocaleString('pt-BR', { maximumFractionDigits: 1 })
}

function VehiclesView({ veiculos, onNovoVeiculo, onExcluir }: VehiclesViewProps) {
  const [busca, setBusca] = useState('')
  const [filtro, setFiltro] = useState<StatusVeiculo | null>(null)

  const kpis: Kpi[] = useMemo(() => {
    const contar = (status: StatusVeiculo) => veiculos.filter((v) => v.status === status).length
    return [
      { chave: 'total', rotulo: 'Total', contagem: veiculos.length, sufixo: 'veículos' },
      { chave: 'Disponível', rotulo: 'Disponíveis', contagem: contar('Disponível'), sufixo: 'prontos' },
      { chave: 'Em uso', rotulo: 'Em uso', contagem: contar('Em uso'), sufixo: 'na rota' },
      { chave: 'Em Manutenção', rotulo: 'Manutenção', contagem: contar('Em Manutenção'), sufixo: 'na oficina' },
      { chave: 'Indisponível', rotulo: 'Indisponíveis', contagem: contar('Indisponível'), sufixo: 'fora' },
    ]
  }, [veiculos])

  const filtrados = useMemo(() => {
    const termo = busca.trim().toLowerCase()
    return veiculos.filter((veiculo) => {
      if (filtro !== null && veiculo.status !== filtro) return false
      if (termo === '') return true
      return (
        veiculo.modelo.toLowerCase().includes(termo) ||
        veiculo.placa.toLowerCase().includes(termo) ||
        veiculo.categoria.toLowerCase().includes(termo)
      )
    })
  }, [veiculos, busca, filtro])

  const alternarFiltro = (chave: StatusVeiculo | 'total') => {
    setFiltro((prev) => (prev === chave || chave === 'total' ? null : chave))
  }

  return (
    <main className="vehicles-view">
      <header className="vehicles-header">
        <div>
          <h1>Frota de veículos</h1>
          <p>Cadastro e consulta dos veículos da empresa, persistidos no banco de dados.</p>
        </div>
        <button type="button" className="button button-primary" onClick={onNovoVeiculo}>
          <Icon name="plus" size={15} /> Novo veículo
        </button>
      </header>

      <div className="kpi-grid">
        {kpis.map((kpi) => (
          <button
            key={kpi.chave}
            type="button"
            className={`kpi-card ${filtro === kpi.chave ? 'kpi-card-active' : ''}`}
            onClick={() => alternarFiltro(kpi.chave)}
          >
            <span className="kpi-valor">{kpi.contagem}</span>
            <span className="kpi-rotulo">{kpi.rotulo}</span>
            <span className="kpi-sufixo">{kpi.sufixo}</span>
          </button>
        ))}
      </div>

      <div className="vehicles-toolbar">
        <div className="search-box">
          <Icon name="search" size={15} />
          <input
            type="text"
            placeholder="Buscar por modelo, placa ou categoria…"
            value={busca}
            onChange={(event) => setBusca(event.target.value)}
          />
          {busca !== '' && (
            <button
              type="button"
              className="search-clear"
              onClick={() => setBusca('')}
              aria-label="Limpar busca"
            >
              <Icon name="close" size={13} />
            </button>
          )}
        </div>
        <span className="vehicles-count">
          {filtrados.length} {filtrados.length === 1 ? 'veículo' : 'veículos'}
          {filtro !== null && ` · ${filtro}`}
        </span>
      </div>

      {veiculos.length === 0 ? (
        <div className="vehicles-empty">
          <Icon name="truck" size={28} />
          <p>Nenhum veículo cadastrado ainda.</p>
          <button type="button" className="button button-primary" onClick={onNovoVeiculo}>
            Cadastrar o primeiro veículo
          </button>
        </div>
      ) : filtrados.length === 0 ? (
        <div className="vehicles-empty">
          <Icon name="search" size={28} />
          <p>Nenhum veículo corresponde à busca atual.</p>
          <button
            type="button"
            className="button button-secondary"
            onClick={() => {
              setBusca('')
              setFiltro(null)
            }}
          >
            Limpar filtros
          </button>
        </div>
      ) : (
        <div className="vehicles-table-wrap">
          <table className="vehicles-table">
            <thead>
              <tr>
                <th>Modelo</th>
                <th>Categoria</th>
                <th>Placa</th>
                <th>Status</th>
                <th className="num">Kilometragem</th>
                <th className="num">Velocidade</th>
                <th>Cadastrado</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {filtrados.map((veiculo) => (
                <tr key={veiculo.id}>
                  <td className="veiculo-modelo">{veiculo.modelo}</td>
                  <td>{veiculo.categoria}</td>
                  <td className="veiculo-placa">{veiculo.placa}</td>
                  <td>
                    <span className={`status-badge status-${slugDoStatus(veiculo.status)}`}>
                      {veiculo.status}
                    </span>
                  </td>
                  <td className="num">{formatarNumero(veiculo.kilometragem)} km</td>
                  <td className="num">{formatarNumero(veiculo.velocidade)} km/h</td>
                  <td className="veiculo-data">
                    {new Date(veiculo.created_at).toLocaleDateString('pt-BR')}
                  </td>
                  <td>
                    <button
                      type="button"
                      className="icon-button"
                      title="Excluir veículo"
                      onClick={() => onExcluir(veiculo)}
                    >
                      <Icon name="trash" size={15} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
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