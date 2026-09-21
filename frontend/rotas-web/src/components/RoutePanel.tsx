import type { Algorithm, LatLng, RouteResponse } from '../types'

interface RoutePanelProps {
  algorithm: Algorithm
  onAlgorithmChange: (algorithm: Algorithm) => void
  origin: LatLng | null
  destination: LatLng | null
  loading: boolean
  error: string | null
  result: RouteResponse | null
  onCalculate: () => void
  onClear: () => void
}

function RoutePanel({
  algorithm,
  onAlgorithmChange,
  origin,
  destination,
  loading,
  error,
  result,
  onCalculate,
  onClear,
}: RoutePanelProps) {
  const canCalculate = origin !== null && destination !== null && !loading

  return (
    <aside className="sidebar">
      <header className="sidebar-header">
        <h1>Rotas Go</h1>
        <p>
          Projeto acadêmico de Teoria dos Grafos que calcula a menor rota entre cidades do Espírito
          Santo com Dijkstra ou A*.
        </p>
      </header>

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

      <div className="route-status">
        <p>
          <strong>Origem:</strong>{' '}
          {origin === null
            ? 'Clique no mapa para definir'
            : `${origin.lat.toFixed(4)}, ${origin.lng.toFixed(4)}`}
        </p>
        <p>
          <strong>Destino:</strong>{' '}
          {destination === null
            ? 'Clique no mapa para definir'
            : `${destination.lat.toFixed(4)}, ${destination.lng.toFixed(4)}`}
        </p>
      </div>

      <div className="sidebar-actions">
        <button
          type="button"
          className="button button-primary"
          onClick={onCalculate}
          disabled={!canCalculate}
        >
          {loading ? 'Calculando…' : 'Calcular rota'}
        </button>
        <button type="button" className="button button-secondary" onClick={onClear}>
          Limpar
        </button>
      </div>

      {error !== null && <div className="error-box">{error}</div>}

      {result !== null && (
        <div className="result-card">
          <h2>Resultado</h2>
          <p className="result-distance">{result.distance_km.toFixed(1)} km</p>
          <p className="result-meta">
            <strong>Algoritmo:</strong> {result.algorithm === 'dijkstra' ? 'Dijkstra' : 'A*'}
          </p>
          <p className="result-meta">
            <strong>Nós visitados:</strong> {result.nodes_visited}
          </p>
          <p className="result-path">{result.path.map((node) => node.name).join(' → ')}</p>
        </div>
      )}
    </aside>
  )
}

export default RoutePanel
