import { useEffect, useState } from 'react'
import {
  calculateAlternatives,
  calculateBatch,
  createBase,
  deleteBase,
  deleteSavedRoute,
  getGraphInfo,
  getSavedRoute,
  listBases,
  listSavedRoutes,
  saveRoute,
} from './api/routes'
import RouteMap from './components/RouteMap'
import RoutePanel from './components/RoutePanel'
import VehiclesView from './components/VehiclesView'
import type { AltState } from './components/AltStepper'
import type {
  Algorithm,
  Base,
  BatchRouteRequest,
  Cost,
  GraphInfo,
  LatLng,
  NovoVeiculo,
  RouteLeg,
  RouteResponse,
  SavedRouteSummary,
  Veiculo,
  VehicleRoute,
} from './types'
import { createVehicle, deleteVehicle, listVehicles } from './api/vehicles'
import './App.css'

type Visao = 'rotas' | 'veiculos'

function criarVeiculo(numero: number): VehicleRoute {
  return { id: crypto.randomUUID(), label: `Veículo ${numero}`, points: [] }
}

const primeiroVeiculo = criarVeiculo(1)

function alternativasIniciais(): AltState {
  return { active: false, step: 0, total: 0, legs: null, selections: {} }
}

function App() {
  const [vehicles, setVehicles] = useState<VehicleRoute[]>(() => [primeiroVeiculo])
  const [activeVehicleId, setActiveVehicleId] = useState<string>(primeiroVeiculo.id)
  const [algorithm, setAlgorithm] = useState<Algorithm>('dijkstra')
  const [cost, setCost] = useState<Cost>('duration')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [results, setResults] = useState<Record<string, RouteResponse> | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [graphInfo, setGraphInfo] = useState<GraphInfo | null>(null)
  const [savedRoutes, setSavedRoutes] = useState<SavedRouteSummary[]>([])
  const [saveName, setSaveName] = useState('')
  const [alt, setAlt] = useState<AltState>(alternativasIniciais)
  const [altVehicleId, setAltVehicleId] = useState<string | null>(null)
  const [bases, setBases] = useState<Base[]>([])
  const [addingBase, setAddingBase] = useState(false)
  const [baseName, setBaseName] = useState('')
  const [view, setView] = useState<Visao>('rotas')
  const [veiculosCadastrados, setVeiculosCadastrados] = useState<Veiculo[]>([])

  useEffect(() => {
    getGraphInfo().then(setGraphInfo).catch(() => setGraphInfo(null))
    listSavedRoutes().then(setSavedRoutes).catch(() => setSavedRoutes([]))
    listBases().then(setBases).catch(() => setBases([]))
    listVehicles().then(setVeiculosCadastrados).catch(() => setVeiculosCadastrados([]))
  }, [])

  const sairDasAlternativas = () => {
    setAlt(alternativasIniciais())
    setAltVehicleId(null)
  }

  const invalidate = () => {
    sairDasAlternativas()
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
      cost,
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

  const handleAlternatives = async () => {
    const ativo = vehicles.find((v) => v.id === activeVehicleId)
    if (ativo === undefined || ativo.points.length < 2 || loading) return
    setLoading(true)
    setError(null)
    setNotice(null)
    try {
      const resposta = await calculateAlternatives({
        origin: ativo.points[0],
        destination: ativo.points[ativo.points.length - 1],
        waypoints: ativo.points.slice(1, -1),
        cost,
        max_alternatives: 3,
      })
      setAltVehicleId(ativo.id)
      setAlt({
        active: true,
        step: 0,
        total: resposta.legs.length,
        legs: resposta.legs,
        selections: {},
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao calcular alternativas')
    } finally {
      setLoading(false)
    }
  }

  const selectAlternative = (step: number, alternativeIndex: number) => {
    setAlt((prev) => {
      if (!prev.active || prev.legs === null) return prev
      const novaSelecao = { ...prev.selections, [step]: alternativeIndex }
      const proximoPasso = step + 1 < prev.total ? step + 1 : step
      return { ...prev, selections: novaSelecao, step: proximoPasso }
    })
  }

  const applyAlternatives = () => {
    if (!alt.active || alt.legs === null || altVehicleId === null) return
    const legs: RouteLeg[] = alt.legs.map((leg, index) => {
      const escolha = alt.selections[index] ?? 0
      const alternativa = leg.alternatives[escolha] ?? leg.alternatives[0]
      return {
        origin: leg.origin,
        destination: leg.destination,
        distance_km: alternativa.distance_km,
        estimated_duration_minutes: alternativa.estimated_duration_minutes,
        nodes_visited: alternativa.nodes_visited,
        geometry: alternativa.geometry,
      }
    })
    const geometria: LatLng[] = []
    legs.forEach((leg, index) => {
      if (index === 0) {
        geometria.push(...leg.geometry)
      } else {
        geometria.push(...leg.geometry.slice(1))
      }
    })
    const totalDistancia = legs.reduce((acc, leg) => acc + leg.distance_km, 0)
    const totalDuracao = legs.reduce((acc, leg) => acc + leg.estimated_duration_minutes, 0)
    const totalVisitados = legs.reduce((acc, leg) => acc + leg.nodes_visited, 0)
    const rota: RouteResponse = {
      origin: legs[0].origin,
      destination: legs[legs.length - 1].destination,
      waypoints: legs.slice(1, -1).map((leg) => leg.origin),
      algorithm,
      distance_km: totalDistancia,
      estimated_duration_minutes: totalDuracao,
      nodes_visited: totalVisitados,
      search_time_ms: 0,
      geometry: geometria,
      legs,
      data_source: graphInfo?.data_source ?? 'OSM',
    }
    setResults((prev) => ({ ...(prev ?? {}), [altVehicleId]: rota }))
    setVehicles((prev) =>
      prev.map((v) =>
        v.id === altVehicleId
          ? { ...v, points: [rota.origin, ...rota.waypoints, rota.destination] }
          : v,
      ),
    )
    sairDasAlternativas()
    setNotice('Rota montada com as alternativas escolhidas.')
  }

  const startAddBase = () => {
    setBaseName('')
    setAddingBase(true)
  }

  const cancelAddBase = () => {
    setAddingBase(false)
    setBaseName('')
  }

  const handleMapClick = async (position: LatLng) => {
    if (addingBase) {
      if (baseName.trim() === '') {
        setError('Dê um nome à base antes de clicar no mapa.')
        return
      }
      try {
        const base = await createBase({ name: baseName.trim(), lat: position.lat, lng: position.lng })
        setBases((prev) => [...prev, base].sort((a, b) => a.name.localeCompare(b.name)))
        setNotice(`Base "${base.name}" criada.`)
        cancelAddBase()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao criar a base')
      }
      return
    }
    addPoint(position)
  }

  const selectBaseAsPoint = (base: Base) => {
    if (addingBase) return
    const ponto: LatLng = { lat: base.lat, lng: base.lng }
    setVehicles((prev) =>
      prev.map((v) =>
        v.id === activeVehicleId ? { ...v, points: [...v.points, ponto] } : v,
      ),
    )
    invalidate()
  }

  const handleDeleteBase = async (id: number) => {
    try {
      await deleteBase(id)
      setBases((prev) => prev.filter((b) => b.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir a base')
    }
  }

  const cadastrarVeiculo = async (req: NovoVeiculo) => {
    const veiculo = await createVehicle(req)
    setVeiculosCadastrados((prev) => [veiculo, ...prev])
  }

  const excluirVeiculo = async (id: number) => {
    await deleteVehicle(id)
    setVeiculosCadastrados((prev) => prev.filter((v) => v.id !== id))
  }

  const selecionarVeiculo = (id: string) => {
    if (id !== activeVehicleId) {
      sairDasAlternativas()
    }
    setActiveVehicleId(id)
  }

  const handleClear = () => {
    const v = criarVeiculo(1)
    setVehicles([v])
    setActiveVehicleId(v.id)
    setResults(null)
    setError(null)
    setNotice(null)
    sairDasAlternativas()
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
      <header className="app-topbar">
        <span className="app-brand">Rotas Go</span>
        <nav className="view-tabs">
          <button
            type="button"
            className={view === 'rotas' ? 'view-tab view-tab-active' : 'view-tab'}
            onClick={() => setView('rotas')}
          >
            Rotas
          </button>
          <button
            type="button"
            className={view === 'veiculos' ? 'view-tab view-tab-active' : 'view-tab'}
            onClick={() => setView('veiculos')}
          >
            Veículos
          </button>
        </nav>
      </header>
      <div className="app-body">
        {view === 'rotas' ? (
          <>
            <RoutePanel
              algorithm={algorithm}
              onAlgorithmChange={setAlgorithm}
              cost={cost}
              onCostChange={setCost}
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
              onSelectVehicle={selecionarVeiculo}
              onAddVehicle={addVehicle}
              onRemoveVehicle={removeVehicle}
              onUpdateLabel={updateLabel}
              onRemovePoint={removePoint}
              onCalculate={handleCalculate}
              onAlternatives={handleAlternatives}
              onClear={handleClear}
              onSave={handleSave}
              onLoadSaved={loadSaved}
              onDeleteSaved={deleteSaved}
              alt={alt}
              onSelectAlternative={selectAlternative}
              onApplyAlternatives={applyAlternatives}
              onCancelAlternatives={sairDasAlternativas}
              bases={bases}
              addingBase={addingBase}
              baseName={baseName}
              onBaseNameChange={setBaseName}
              onStartAddBase={startAddBase}
              onCancelAddBase={cancelAddBase}
              onSelectBase={selectBaseAsPoint}
              onDeleteBase={handleDeleteBase}
            />
            <RouteMap
              vehicles={vehicles}
              activeVehicleId={activeVehicleId}
              results={results}
              alt={alt}
              altVehicleId={altVehicleId}
              bases={bases}
              onMapClick={handleMapClick}
              onPointDrag={movePoint}
              onSelectAlternative={selectAlternative}
              onSelectBase={selectBaseAsPoint}
            />
          </>
        ) : (
          <VehiclesView
            veiculos={veiculosCadastrados}
            onCadastrar={cadastrarVeiculo}
            onExcluir={excluirVeiculo}
          />
        )}
      </div>
    </div>
  )
}

export default App