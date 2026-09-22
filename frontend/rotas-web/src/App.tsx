import { useEffect, useState } from 'react'
import { calculateRoute, getGraphInfo } from './api/routes'
import RouteMap from './components/RouteMap'
import RoutePanel from './components/RoutePanel'
import type { Algorithm, GraphInfo, LatLng, RouteResponse } from './types'
import './App.css'

function App() {
  const [origin, setOrigin] = useState<LatLng | null>(null)
  const [destination, setDestination] = useState<LatLng | null>(null)
  const [algorithm, setAlgorithm] = useState<Algorithm>('dijkstra')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<RouteResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [graphInfo, setGraphInfo] = useState<GraphInfo | null>(null)

  useEffect(() => {
    getGraphInfo().then(setGraphInfo).catch(() => setGraphInfo(null))
  }, [])

  const changeOrigin = (position: LatLng) => {
    setOrigin(position)
    setResult(null)
    setError(null)
  }

  const changeDestination = (position: LatLng) => {
    setDestination(position)
    setResult(null)
    setError(null)
  }

  const handleCalculate = async () => {
    if (origin === null || destination === null) return
    setLoading(true)
    setError(null)
    setResult(null)
    try {
      const response = await calculateRoute({ origin, destination, algorithm })
      setOrigin(response.origin)
      setDestination(response.destination)
      setResult(response)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao calcular a rota')
    } finally {
      setLoading(false)
    }
  }

  const handleClear = () => {
    setOrigin(null)
    setDestination(null)
    setResult(null)
    setError(null)
  }

  return (
    <div className="app">
      <RoutePanel
        algorithm={algorithm}
        onAlgorithmChange={setAlgorithm}
        origin={origin}
        destination={destination}
        loading={loading}
        error={error}
        result={result}
        graphInfo={graphInfo}
        onCalculate={handleCalculate}
        onClear={handleClear}
      />
      <RouteMap
        origin={origin}
        destination={destination}
        response={result}
        onOriginChange={changeOrigin}
        onDestinationChange={changeDestination}
        onClear={handleClear}
      />
    </div>
  )
}

export default App
