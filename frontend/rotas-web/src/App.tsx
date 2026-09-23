import { useEffect, useState } from 'react'
import {
  calculateBatch,
  deleteSavedRoute,
  getGraphInfo,
  getSavedRoute,
  listSavedRoutes,
  saveRoute,
} from './api/routes'
import RouteMap from './components/RouteMap'
import RoutePanel from './components/RoutePanel'
import type {
  Algorithm,
  BatchRouteRequest,
  GraphInfo,
  LatLng,
  RouteResponse,
  SavedRouteSummary,
  VehicleRoute,
} from './types'
import './App.css'

function criarVeiculo(numero: number): VehicleRoute {
  return { id: crypto.randomUUID(), label: `Veículo ${numero}`, points: [] }
}

const primeiroVeiculo = criarVeiculo(1)

function App() {
  const [vehicles, setVehicles] = useState<VehicleRoute[]>(() => [primeiroVeiculo])
  const [activeVehicleId, setActiveVehicleId] = useState<string>(primeiroVeiculo.id)
  const [algorithm, setAlgorithm] = useState<Algorithm>('dijkstra')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [results, setResults] = useState<Record<string, RouteResponse> | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [graphInfo, setGraphInfo] = useState<GraphInfo | null>(null)
  const [savedRoutes, setSavedRoutes] = useState<SavedRouteSummary[]>([])
  const [saveName, setSaveName] = useState('')

  useEffect(() => {
    getGraphInfo().then(setGraphInfo).catch(() => setGraphInfo(null))
    listSavedRoutes().then(setSavedRoutes).catch(() => setSavedRoutes([]))
  }, [])

  const invalidate = () => {
    setResults(null)
    setError(null)
  }

  const addPoint = (position: LatLng) => {
    setVehicles((prev) =>
      prev.map((v) => (v.id === activeVehicleId ? { ...v, points: [...v.points, position] } : v)),
    )
    invalidate()
  }

  const movePoint = (vehicleId: string, index: number, position: LatLng) => {
    setVehicles((prev) =>
      prev.map((v) =>
        v.id === vehicleId
          ? { ...v, points: v.points.map((p, i) => (i === index ? position : p)) }
          : v,
      ),
    )
    invalidate()
  }

  const removePoint = (vehicleId: string, index: number) => {
    setVehicles((prev) =>
      prev.map((v) =>
        v.id === vehicleId ? { ...v, points: v.points.filter((_, i) => i !== index) } : v,
      ),
    )
    invalidate()
  }

  const addVehicle = () => {
    const v = criarVeiculo(vehicles.length + 1)
    setVehicles((prev) => [...prev, v])
    setActiveVehicleId(v.id)
    invalidate()
  }

  const removeVehicle = (id: string) => {
    const rest = vehicles.filter((v) => v.id !== id)
    const proximos = rest.length === 0 ? [criarVeiculo(1)] : rest
    setVehicles(proximos)
    if (activeVehicleId === id) {
      setActiveVehicleId(proximos[0].id)
    }
    setResults((prev) => {
      if (prev === null) return prev
      const copia = { ...prev }
      delete copia[id]
      return copia
    })
    setError(null)
  }

  const updateLabel = (id: string, label: string) => {
    setVehicles((prev) => prev.map((v) => (v.id === id ? { ...v, label } : v)))
  }

  const calculaveis = vehicles.filter((vehicle) => vehicle.points.length >= 2)
  const canCalculate = calculaveis.length > 0 && !loading

  const handleCalculate = async () => {
    if (!canCalculate) return
    const request: BatchRouteRequest = {
      algorithm,
      vehicles: calculaveis.map((v) => ({
        id: v.id,
        origin: v.points[0],
        destination: v.points[v.points.length - 1],
        waypoints: v.points.slice(1, -1),
      })),
    }
    setLoading(true)
    setError(null)
    setNotice(null)
    setResults(null)
    try {
      const response = await calculateBatch(request)
      const porId: Record<string, RouteResponse> = {}
      for (const item of response.routes) {
        porId[item.id] = item
      }
      setResults(porId)
      setVehicles((prev) =>
        prev.map((v) => {
          const item = response.routes.find((r) => r.id === v.id)
          if (item === undefined) return v
          return { ...v, points: [item.origin, ...item.waypoints, item.destination] }
        }),
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao calcular as rotas')
    } finally {
      setLoading(false)
    }
  }

  const handleClear = () => {
    const v = criarVeiculo(1)
    setVehicles([v])
    setActiveVehicleId(v.id)
    setResults(null)
    setError(null)
    setNotice(null)
  }

  const handleSave = async () => {
    if (results === null || saveName.trim() === '') return
    setSaving(true)
    setError(null)
    setNotice(null)
    try {
      const rotas = Object.entries(results).map(([id, rota]) => ({ ...rota, id }))
      await saveRoute({ name: saveName.trim(), algorithm, routes: rotas })
      setSaveName('')
      setSavedRoutes(await listSavedRoutes())
      setNotice('Rota salva com sucesso.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar a rota')
    } finally {
      setSaving(false)
    }
  }

  const loadSaved = async (id: number) => {
    setLoading(true)
    setError(null)
    setNotice(null)
    try {
      const detalhe = await getSavedRoute(id)
      const novos: VehicleRoute[] = detalhe.routes.map((item, i) => ({
        id: item.id,
        label: `Veículo ${i + 1}`,
        points: [item.origin, ...item.waypoints, item.destination],
      }))
      const porId: Record<string, RouteResponse> = {}
      for (const item of detalhe.routes) {
        porId[item.id] = item
      }
      setVehicles(novos.length > 0 ? novos : [criarVeiculo(1)])
      setActiveVehicleId(novos[0]?.id ?? '')
      setResults(porId)
      setAlgorithm(detalhe.algorithm)
      setSaveName(detalhe.name)
      setNotice('Rota carregada.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar a rota')
    } finally {
      setLoading(false)
    }
  }

  const deleteSaved = async (id: number) => {
    try {
      await deleteSavedRoute(id)
      setSavedRoutes((prev) => prev.filter((r) => r.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir a rota')
    }
  }

  return (
    <div className="app">
      <RoutePanel
        algorithm={algorithm}
        onAlgorithmChange={setAlgorithm}
        vehicles={vehicles}
        activeVehicleId={activeVehicleId}
        results={results}
        loading={loading}
        saving={saving}
        error={error}
        notice={notice}
        canCalculate={canCalculate}
        graphInfo={graphInfo}
        savedRoutes={savedRoutes}
        saveName={saveName}
        onSaveNameChange={setSaveName}
        onSelectVehicle={setActiveVehicleId}
        onAddVehicle={addVehicle}
        onRemoveVehicle={removeVehicle}
        onUpdateLabel={updateLabel}
        onRemovePoint={removePoint}
        onCalculate={handleCalculate}
        onClear={handleClear}
        onSave={handleSave}
        onLoadSaved={loadSaved}
        onDeleteSaved={deleteSaved}
      />
      <RouteMap
        vehicles={vehicles}
        activeVehicleId={activeVehicleId}
        results={results}
        onMapClick={addPoint}
        onPointDrag={movePoint}
      />
    </div>
  )
}

export default App