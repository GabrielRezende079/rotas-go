import type {
  Algorithm,
  GraphInfo,
  RouteResponse,
  SavedRouteSummary,
  VehicleRoute,
} from '../types'

interface RoutePanelProps {
  algorithm: Algorithm
  onAlgorithmChange: (algorithm: Algorithm) => void
  vehicles: VehicleRoute[]
  activeVehicleId: string | null
  results: Record<string, RouteResponse> | null
  loading: boolean
  saving: boolean
  error: string | null
  notice: string | null
  canCalculate: boolean
  graphInfo: GraphInfo | null
  savedRoutes: SavedRouteSummary[]
  saveName: string
  onSaveNameChange: (name: string) => void
  onSelectVehicle: (id: string) => void
  onAddVehicle: () => void
  onRemoveVehicle: (id: string) => void
  onUpdateLabel: (id: string, label: string) => void
  onRemovePoint: (vehicleId: string, index: number) => void
  onCalculate: () => void
  onClear: () => void
  onSave: () => void
  onLoadSaved: (id: number) => void
  onDeleteSaved: (id: number) => void
}

function nomeDoPonto(index: number, total: number): string {
  if (index === 0) return 'Origem'
  if (index === total - 1) return 'Destino'
  return `Parada ${index}`
}

function RoutePanel({
  algorithm,
  onAlgorithmChange,
  vehicles,
  activeVehicleId,
  results,
  loading,
  saving,
  error,
  notice,
  canCalculate,
  graphInfo,
  savedRoutes,
  saveName,
  onSaveNameChange,
  onSelectVehicle,
  onAddVehicle,
  onRemoveVehicle,
  onUpdateLabel,
  onRemovePoint,
  onCalculate,
  onClear,
  onSave,
  onLoadSaved,
  onDeleteSaved,
}: RoutePanelProps) {
  const ativo = vehicles.find((v) => v.id === activeVehicleId) ?? null
  const resultadoAtivo =
    results !== null && ativo !== null ? results[ativo.id] ?? null : null
  const rotasSalvasVisiveis = results !== null

  return (
    <aside className="sidebar">
      <header className="sidebar-header">
        <h1>Rotas Go</h1>
        <p>
          Menor rota na malha viária real do Espírito Santo, calculada pelo nosso próprio Dijkstra
          ou A*.
        </p>
      </header>

      {graphInfo !== null && (
        <div className={graphInfo.real_road_graph ? 'graph-status graph-status-real' : 'graph-status'}>
          <strong>{graphInfo.real_road_graph ? 'Malha OSM real carregada' : 'Modo didático'}</strong>
          <span>
            {graphInfo.vertices.toLocaleString('pt-BR')} vértices ·{' '}
            {graphInfo.edges.toLocaleString('pt-BR')} arestas
          </span>
        </div>
      )}

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

      <section className="vehicles-section">
        <div className="section-title">
          <label>Veículos</label>
          <button type="button" className="button button-small button-secondary" onClick={onAddVehicle}>
            + Adicionar
          </button>
        </div>
        <ul className="vehicle-list">
          {vehicles.map((vehicle) => (
            <li
              key={vehicle.id}
              className={`vehicle-card ${vehicle.id === activeVehicleId ? 'vehicle-card-active' : ''}`}
              onClick={() => onSelectVehicle(vehicle.id)}
            >
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
          </>
        )}
      </section>

      <div className="sidebar-actions">
        <button
          type="button"
          className="button button-primary"
          onClick={onCalculate}
          disabled={!canCalculate}
        >
          {loading ? 'Calculando…' : 'Calcular rotas'}
        </button>
        <button type="button" className="button button-secondary" onClick={onClear}>
          Limpar
        </button>
      </div>

      {error !== null && <div className="error-box">{error}</div>}
      {notice !== null && <div className="notice-box">{notice}</div>}

      {rotasSalvasVisiveis && results !== null && (
        <section className="save-section">
          <div className="section-title">
            <label>Salvar rota</label>
          </div>
          <div className="save-form">
            <input
              type="text"
              placeholder="Nome da rota (ex.: entrega norte)"
              value={saveName}
              onChange={(event) => onSaveNameChange(event.target.value)}
            />
            <button
              type="button"
              className="button button-primary"
              onClick={onSave}
              disabled={saving || saveName.trim() === ''}
            >
              {saving ? 'Salvando…' : 'Salvar'}
            </button>
          </div>
        </section>
      )}

      {results !== null && (
        <section className="results-section">
          <div className="section-title">
            <label>Resultados</label>
          </div>
          {vehicles.map((vehicle) => {
            const resultado = results[vehicle.id]
            if (resultado === undefined) return null
            return (
              <div key={vehicle.id} className="result-card result-compact">
                <p className="result-title">{vehicle.label}</p>
                <p className="result-distance">{resultado.distance_km.toFixed(1)} km</p>
                <p className="result-meta">
                  {resultado.estimated_duration_minutes.toFixed(1)} min ·{' '}
                  {resultado.geometry.length.toLocaleString('pt-BR')} pontos na malha
                </p>
              </div>
            )
          })}
          {resultadoAtivo !== null && (
            <p className="result-source">Fonte: {resultadoAtivo.data_source}</p>
          )}
        </section>
      )}

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
                    className="button button-small button-primary"
                    onClick={() => onLoadSaved(saved.id)}
                  >
                    Carregar
                  </button>
                  <button
                    type="button"
                    className="button button-small button-secondary"
                    onClick={() => onDeleteSaved(saved.id)}
                  >
                    Excluir
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </aside>
  )
}

export default RoutePanel