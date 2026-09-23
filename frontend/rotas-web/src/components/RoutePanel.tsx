import type {
  Algorithm,
  Base,
  Cost,
  GraphInfo,
  RouteResponse,
  SavedRouteSummary,
  VehicleRoute,
} from '../types'
import { CORES_VEICULO } from '../colors'
import AltStepper, { type AltState } from './AltStepper'

interface RoutePanelProps {
  algorithm: Algorithm
  onAlgorithmChange: (algorithm: Algorithm) => void
  cost: Cost
  onCostChange: (cost: Cost) => void
  vehicles: VehicleRoute[]
  activeVehicleId: string | null
  results: Record<string, RouteResponse> | null
  graphInfo: GraphInfo | null
  savedRoutes: SavedRouteSummary[]
  alt: AltState
  onSelectVehicle: (id: string) => void
  onAddVehicle: () => void
  onRemoveVehicle: (id: string) => void
  onUpdateLabel: (id: string, label: string) => void
  onRemovePoint: (vehicleId: string, index: number) => void
  onSelectAlternative: (step: number, alternativeIndex: number) => void
  onApplyAlternatives: () => void
  onCancelAlternatives: () => void
  onLoadSaved: (id: number) => void
  onDeleteSaved: (saved: SavedRouteSummary) => void
  bases: Base[]
  addingBase: boolean
  onStartAddBase: () => void
  onSelectBase: (base: Base) => void
  onDeleteBase: (base: Base) => void
}

function nomeDoPonto(index: number, total: number): string {
  if (index === 0) return 'Origem'
  if (index === total - 1) return 'Destino'
  return `Parada ${index}`
}

function corDoVeiculo(index: number): string {
  return CORES_VEICULO[index % CORES_VEICULO.length]
}

function RoutePanel({
  algorithm,
  onAlgorithmChange,
  cost,
  onCostChange,
  vehicles,
  activeVehicleId,
  results,
  graphInfo,
  savedRoutes,
  alt,
  onSelectVehicle,
  onAddVehicle,
  onRemoveVehicle,
  onUpdateLabel,
  onRemovePoint,
  onSelectAlternative,
  onApplyAlternatives,
  onCancelAlternatives,
  onLoadSaved,
  onDeleteSaved,
  bases,
  addingBase,
  onStartAddBase,
  onSelectBase,
  onDeleteBase,
}: RoutePanelProps) {
  const ativo = vehicles.find((v) => v.id === activeVehicleId) ?? null
  const resultados = results !== null ? Object.values(results) : []
  const totalKm = resultados.reduce((acc, r) => acc + r.distance_km, 0)
  const totalMin = resultados.reduce((acc, r) => acc + r.estimated_duration_minutes, 0)

  return (
    <aside className="sidebar">
      <div className="sidebar-scroll">
        {graphInfo !== null && (
          <div
            className={
              graphInfo.real_road_graph ? 'graph-status graph-status-real' : 'graph-status'
            }
          >
            <strong>
              {graphInfo.real_road_graph ? 'Malha OSM real carregada' : 'Modo didático'}
            </strong>
            <span>
              {graphInfo.vertices.toLocaleString('pt-BR')} vértices ·{' '}
              {graphInfo.edges.toLocaleString('pt-BR')} arestas
            </span>
          </div>
        )}

        <div className="field-grid-2">
          <div className="field">
            <label htmlFor="algorithm">Algoritmo</label>
            <select
              id="algorithm"
              value={algorithm}
              onChange={(event) => onAlgorithmChange(event.target.value as Algorithm)}
            >
              <option value="dijkstra">Dijkstra</option>
              <option value="astar">A*</option>
            </select>
          </div>

          <div className="field">
            <label htmlFor="cost">Custo a minimizar</label>
            <select
              id="cost"
              value={cost}
              onChange={(event) => onCostChange(event.target.value as Cost)}
            >
              <option value="duration">Duração</option>
              <option value="distance">Distância</option>
            </select>
          </div>
        </div>

        <section className="vehicles-section">
          <div className="section-title">
            <label>Veículos</label>
            <button type="button" className="button button-small button-secondary" onClick={onAddVehicle}>
              <span className="button-plus">+</span> Adicionar
            </button>
          </div>
          <ul className="vehicle-list">
            {vehicles.map((vehicle, index) => (
              <li
                key={vehicle.id}
                className={`vehicle-card ${
                  vehicle.id === activeVehicleId ? 'vehicle-card-active' : ''
                }`}
                onClick={() => onSelectVehicle(vehicle.id)}
              >
                <span className="vehicle-swatch" style={{ backgroundColor: corDoVeiculo(index) }} />
                <input
                  type="text"
                  value={vehicle.label}
                  className="vehicle-label"
                  onClick={(event) => event.stopPropagation()}
                  onChange={(event) => onUpdateLabel(vehicle.id, event.target.value)}
                />
                <span className="vehicle-count">
                  {vehicle.points.length} {vehicle.points.length === 1 ? 'ponto' : 'pontos'}
                </span>
                <button
                  type="button"
                  className="icon-button"
                  title="Remover veículo"
                  disabled={vehicles.length <= 1}
                  onClick={(event) => {
                    event.stopPropagation()
                    onRemoveVehicle(vehicle.id)
                  }}
                >
                  ×
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section className="points-section">
          <div className="section-title">
            <label>Pontos do veículo ativo</label>
          </div>
          {ativo === null ? (
            <p className="hint">Adicione um veículo para começar.</p>
          ) : (
            <>
              <p className="hint">
                Clique no mapa para adicionar pontos: o 1º é a origem, o último é o destino e os
                demais são paradas intermediárias.
              </p>
              {ativo.points.length === 0 ? (
                <p className="hint">Nenhum ponto definido ainda.</p>
              ) : (
                <ul className="point-list">
                  {ativo.points.map((point, index) => (
                    <li key={`${index}-${point.lat}-${point.lng}`} className="point-row">
                      <span className="point-name">{nomeDoPonto(index, ativo.points.length)}</span>
                      <span className="point-coords">
                        {point.lat.toFixed(4)}, {point.lng.toFixed(4)}
                      </span>
                      <button
                        type="button"
                        className="icon-button"
                        title="Remover ponto"
                        onClick={() => onRemovePoint(ativo.id, index)}
                      >
                        ×
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}
        </section>

        <AltStepper
          alt={alt}
          cost={cost}
          onSelectAlternative={onSelectAlternative}
          onConfirm={onApplyAlternatives}
          onCancel={onCancelAlternatives}
        />

        <section className="bases-section">
          <div className="section-title">
            <label>Bases</label>
            <button
              type="button"
              className="button button-small button-secondary"
              onClick={onStartAddBase}
              disabled={addingBase}
            >
              <span className="button-plus">+</span> Adicionar
            </button>
          </div>
          {addingBase ? (
            <p className="hint">Modo de posicionamento ativo: clique no mapa.</p>
          ) : bases.length === 0 ? (
            <p className="hint">Nenhuma base cadastrada.</p>
          ) : (
            <ul className="bases-list">
              {bases.map((base) => (
                <li key={base.id} className="saved-card">
                  <div className="saved-info">
                    <strong>{base.name}</strong>
                    <span>
                      {base.lat.toFixed(4)}, {base.lng.toFixed(4)}
                    </span>
                  </div>
                  <div className="saved-actions">
                    <button
                      type="button"
                      className="button button-small button-soft"
                      title="Adiciona a base como próximo ponto do veículo ativo"
                      onClick={() => onSelectBase(base)}
                    >
                      + Rota
                    </button>
                    <button
                      type="button"
                      className="button button-small button-secondary"
                      onClick={() => onDeleteBase(base)}
                    >
                      Excluir
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="saved-section">
          <div className="section-title">
            <label>Rotas salvas</label>
          </div>
          {savedRoutes.length === 0 ? (
            <p className="hint">Nenhuma rota salva ainda.</p>
          ) : (
            <ul className="saved-list">
              {savedRoutes.map((saved) => (
                <li key={saved.id} className="saved-card">
                  <div className="saved-info">
                    <strong>{saved.name}</strong>
                    <span>
                      {saved.vehicle_count} {saved.vehicle_count === 1 ? 'veículo' : 'veículos'} ·{' '}
                      {saved.total_distance_km.toFixed(1)} km
                    </span>
                    <span>{new Date(saved.created_at).toLocaleString('pt-BR')}</span>
                  </div>
                  <div className="saved-actions">
                    <button
                      type="button"
                      className="button button-small button-soft"
                      onClick={() => onLoadSaved(saved.id)}
                    >
                      Carregar
                    </button>
                    <button
                      type="button"
                      className="button button-small button-secondary"
                      onClick={() => onDeleteSaved(saved)}
                    >
                      Excluir
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>

      <footer className="sidebar-footer">
        <span className="summary-item">
          <strong>{resultados.length}</strong>
          <span>rotas</span>
        </span>
        <span className="summary-item">
          <strong>{totalKm.toFixed(1)}</strong>
          <span>km</span>
        </span>
        <span className="summary-item">
          <strong>{totalMin.toFixed(1)}</strong>
          <span>min</span>
        </span>
      </footer>
    </aside>
  )
}

export default RoutePanel