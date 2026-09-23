import { useEffect, useRef, useState } from 'react'
import type { ReactNode } from 'react'
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
import { ConfirmDialog } from './components/ConfirmDialog'
import { HelpDialog } from './components/HelpDialog'
import { SaveRouteModal } from './components/SaveRouteModal'
import { BaseNameModal } from './components/BaseNameModal'
import { VehicleFormModal } from './components/VehicleFormModal'
import { Icon } from './components/Icon'
import { Spinner } from './components/Spinner'
import { ToastProvider } from './components/Toaster'
import { useToasts } from './toast'
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

interface Confirmacao {
  titulo: string
  mensagem: ReactNode
  confirmarLabel: string
  executar: () => Promise<void>
}

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
  const [results, setResults] = useState<Record<string, RouteResponse> | null>(null)
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

  const [salvarAberta, setSalvarAberta] = useState(false)
  const [ajudaAberta, setAjudaAberta] = useState(false)
  const [nomearBaseAberto, setNomearBaseAberto] = useState(false)
  const [formVeiculoAberto, setFormVeiculoAberto] = useState(false)
  const [confirmacao, setConfirmacao] = useState<Confirmacao | null>(null)
  const [confirmLoading, setConfirmLoading] = useState(false)

  const toasts = useToasts()

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
      toasts.error(err instanceof Error ? err.message : 'Erro ao calcular as rotas')
    } finally {
      setLoading(false)
    }
  }

  const handleAlternatives = async () => {
    const ativo = vehicles.find((v) => v.id === activeVehicleId)
    if (ativo === undefined || ativo.points.length < 2 || loading) return
    setLoading(true)
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
      toasts.success('Alternativas calculadas.')
    } catch (err) {
      toasts.error(err instanceof Error ? err.message : 'Erro ao calcular alternativas')
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
    toasts.success('Rota montada com as alternativas escolhidas.')
  }

  const iniciarAdicaoDeBase = (nome: string) => {
    setBaseName(nome)
    setNomearBaseAberto(false)
    setAddingBase(true)
  }

  const cancelAddBase = () => {
    setAddingBase(false)
    setBaseName('')
  }

  const handleMapClick = async (position: LatLng) => {
    if (addingBase) {
      try {
        const base = await createBase({
          name: baseName.trim(),
          lat: position.lat,
          lng: position.lng,
        })
        setBases((prev) => [...prev, base].sort((a, b) => a.name.localeCompare(b.name)))
        toasts.success(`Base "${base.name}" criada.`)
        cancelAddBase()
      } catch (err) {
        toasts.error(err instanceof Error ? err.message : 'Erro ao criar a base')
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

  const pedirExclusaoDeBase = (base: Base) => {
    setConfirmacao({
      titulo: 'Excluir base',
      mensagem: (
        <>
          Excluir a base <strong>{base.name}</strong>? Os veículos que usam esta localização não
          serão alterados.
        </>
      ),
      confirmarLabel: 'Excluir',
      executar: async () => {
        await deleteBase(base.id)
        setBases((prev) => prev.filter((b) => b.id !== base.id))
        toasts.success('Base excluída.')
      },
    })
  }

  const pedirRemocaoDeVeiculoDeRota = (id: string) => {
    const veiculo = vehicles.find((v) => v.id === id)
    setConfirmacao({
      titulo: 'Remover veículo',
      mensagem: (
        <>
          Remover <strong>{veiculo?.label ?? 'o veículo'}</strong> desta rota?
        </>
      ),
      confirmarLabel: 'Remover',
      executar: async () => {
        removeVehicle(id)
        toasts.success('Veículo removido da rota.')
      },
    })
  }

  const pedirExclusaoDeRotaSalva = (saved: SavedRouteSummary) => {
    setConfirmacao({
      titulo: 'Excluir rota salva',
      mensagem: (
        <>
          Excluir a rota <strong>{saved.name}</strong>? Esta ação não pode ser desfeita.
        </>
      ),
      confirmarLabel: 'Excluir',
      executar: async () => {
        await deleteSavedRoute(saved.id)
        setSavedRoutes((prev) => prev.filter((r) => r.id !== saved.id))
        toasts.success('Rota salva excluída.')
      },
    })
  }

  const pedirExclusaoDeVeiculo = (veiculo: Veiculo) => {
    setConfirmacao({
      titulo: 'Excluir veículo',
      mensagem: (
        <>
          Excluir <strong>{veiculo.modelo}</strong> ({veiculo.placa})? Esta ação não pode ser
          desfeita.
        </>
      ),
      confirmarLabel: 'Excluir',
      executar: async () => {
        await deleteVehicle(veiculo.id)
        setVeiculosCadastrados((prev) => prev.filter((v) => v.id !== veiculo.id))
        toasts.success('Veículo excluído.')
      },
    })
  }

  const executarConfirmacao = async () => {
    if (confirmacao === null) return
    setConfirmLoading(true)
    try {
      await confirmacao.executar()
      setConfirmacao(null)
    } catch (err) {
      toasts.error(err instanceof Error ? err.message : 'Não foi possível concluir a ação')
      setConfirmacao(null)
    } finally {
      setConfirmLoading(false)
    }
  }

  const cadastrarVeiculo = async (req: NovoVeiculo) => {
    const veiculo = await createVehicle(req)
    setVeiculosCadastrados((prev) => [veiculo, ...prev])
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
    sairDasAlternativas()
    toasts.success('Mapa e rotas limpos.')
  }

  const handleSave = async (nome: string) => {
    if (results === null) throw new Error('Calcule as rotas antes de salvar.')
    const rotas = Object.entries(results).map(([id, rota]) => ({ ...rota, id }))
    await saveRoute({ name: nome, algorithm, routes: rotas })
    setSavedRoutes(await listSavedRoutes())
  }

  const loadSaved = async (id: number) => {
    setLoading(true)
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
      toasts.success('Rota carregada.')
    } catch (err) {
      toasts.error(err instanceof Error ? err.message : 'Erro ao carregar a rota')
    } finally {
      setLoading(false)
    }
  }

  const modalAberto =
    salvarAberta || ajudaAberta || nomearBaseAberto || formVeiculoAberto || confirmacao !== null

  const acoesRef = useRef({
    calcular: () => void handleCalculate(),
    alternativas: () => void handleAlternatives(),
    limpar: handleClear,
    cancelarEsc: () => {
      if (addingBase) {
        cancelAddBase()
      } else if (alt.active) {
        sairDasAlternativas()
      }
    },
    alternarAjuda: () => setAjudaAberta(true),
  })

  const cenarioRef = useRef({
    canCalculate: false,
    loading: false,
    altActive: false,
    addingBase: false,
    modalAberto: false,
    view,
  })

  useEffect(() => {
    acoesRef.current = {
      calcular: () => void handleCalculate(),
      alternativas: () => void handleAlternatives(),
      limpar: handleClear,
      cancelarEsc: () => {
        if (addingBase) {
          cancelAddBase()
        } else if (alt.active) {
          sairDasAlternativas()
        }
      },
      alternarAjuda: () => setAjudaAberta(true),
    }
  })

  useEffect(() => {
    cenarioRef.current = {
      canCalculate,
      loading,
      altActive: alt.active,
      addingBase,
      modalAberto,
      view,
    }
  })

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const c = cenarioRef.current
      if (c.modalAberto && event.key !== '?') {
        if (event.key === 'Escape') return
      }
      const alvo = event.target as HTMLElement | null
      const digitando =
        alvo != null && alvo.tagName === 'INPUT' ||
        alvo?.tagName === 'TEXTAREA' ||
        alvo?.tagName === 'SELECT' ||
        (alvo?.isContentEditable ?? false)
      const meta = event.ctrlKey || event.metaKey
      if (digitando && !meta && event.key !== 'Escape') return
      if (meta && event.key === 'Enter') {
        event.preventDefault()
        if (c.canCalculate && !c.loading && !c.modalAberto) acoesRef.current.calcular()
        return
      }
      if (meta && event.shiftKey && (event.key === 'a' || event.key === 'A')) {
        event.preventDefault()
        if (c.canCalculate && !c.loading && !c.modalAberto) acoesRef.current.alternativas()
        return
      }
      if (meta && (event.key === 'l' || event.key === 'L')) {
        event.preventDefault()
        if (!c.modalAberto) acoesRef.current.limpar()
        return
      }
      if (meta && event.key === '1') {
        event.preventDefault()
        if (!c.modalAberto) setView('rotas')
        return
      }
      if (meta && event.key === '2') {
        event.preventDefault()
        if (!c.modalAberto) setView('veiculos')
        return
      }
      if (event.key === 'Escape') {
        if (c.modalAberto) return
        acoesRef.current.cancelarEsc()
        return
      }
      if (event.key === '?') {
        if (c.modalAberto) return
        acoesRef.current.alternarAjuda()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  return (
    <div className="app">
      <header className="app-topbar">
        <span className="app-brand">
          <span className="app-brand-mark">
            <Icon name="route" size={16} />
          </span>
          Rotas Go
        </span>
        <nav className="view-tabs" aria-label="Seções">
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
        {view === 'rotas' && (
          <div className="topbar-actions">
            <button
              type="button"
              className="button button-primary"
              onClick={() => void handleCalculate()}
              disabled={!canCalculate}
            >
              {loading ? (
                <>
                  <Spinner size={14} /> Calculando…
                </>
              ) : (
                <>
                  <Icon name="zap" size={15} /> Calcular
                </>
              )}
            </button>
            <button
              type="button"
              className="button button-secondary"
              onClick={() => void handleAlternatives()}
              disabled={!canCalculate || alt.active}
            >
              <Icon name="route" size={15} /> Alternativas
            </button>
            <button
              type="button"
              className="button button-secondary"
              onClick={() => setSalvarAberta(true)}
              disabled={results === null}
            >
              <Icon name="save" size={15} /> Salvar
            </button>
            <button type="button" className="button button-ghost" onClick={handleClear}>
              <Icon name="clear" size={15} /> Limpar
            </button>
          </div>
        )}
        <button
          type="button"
          className="topbar-help"
          onClick={() => setAjudaAberta(true)}
          aria-label="Ajuda e atalhos de teclado"
          title="Atalhos de teclado"
        >
          <Icon name="keyboard" size={17} />
        </button>
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
              graphInfo={graphInfo}
              savedRoutes={savedRoutes}
              alt={alt}
              onSelectVehicle={selecionarVeiculo}
              onAddVehicle={addVehicle}
              onRemoveVehicle={pedirRemocaoDeVeiculoDeRota}
              onUpdateLabel={updateLabel}
              onRemovePoint={removePoint}
              onSelectAlternative={selectAlternative}
              onApplyAlternatives={applyAlternatives}
              onCancelAlternatives={sairDasAlternativas}
              onLoadSaved={loadSaved}
              onDeleteSaved={pedirExclusaoDeRotaSalva}
              bases={bases}
              addingBase={addingBase}
              onStartAddBase={() => setNomearBaseAberto(true)}
              onSelectBase={selectBaseAsPoint}
              onDeleteBase={pedirExclusaoDeBase}
            />
            <RouteMap
              vehicles={vehicles}
              activeVehicleId={activeVehicleId}
              results={results}
              alt={alt}
              altVehicleId={altVehicleId}
              bases={bases}
              loading={loading}
              addingBase={addingBase}
              baseName={baseName}
              onMapClick={handleMapClick}
              onPointDrag={movePoint}
              onSelectAlternative={selectAlternative}
              onSelectBase={selectBaseAsPoint}
              onClear={handleClear}
              onCancelAddBase={cancelAddBase}
            />
          </>
        ) : (
          <VehiclesView
            veiculos={veiculosCadastrados}
            onNovoVeiculo={() => setFormVeiculoAberto(true)}
            onExcluir={pedirExclusaoDeVeiculo}
          />
        )}
      </div>

      {salvarAberta && (
        <SaveRouteModal
          initialName={saveName}
          totalVeiculos={Object.keys(results ?? {}).length}
          totalKm={Object.values(results ?? {}).reduce((acc, r) => acc + r.distance_km, 0)}
          totalMin={Object.values(results ?? {}).reduce(
            (acc, r) => acc + r.estimated_duration_minutes,
            0,
          )}
          onConfirm={handleSave}
          onClose={() => setSalvarAberta(false)}
        />
      )}

      {nomearBaseAberto && (
        <BaseNameModal
          initialName=""
          onConfirm={iniciarAdicaoDeBase}
          onClose={() => setNomearBaseAberto(false)}
        />
      )}

      {formVeiculoAberto && (
        <VehicleFormModal
          onCadastrar={cadastrarVeiculo}
          onClose={() => setFormVeiculoAberto(false)}
        />
      )}

      <HelpDialog open={ajudaAberta} onClose={() => setAjudaAberta(false)} />

      <ConfirmDialog
        open={confirmacao !== null}
        title={confirmacao?.titulo ?? ''}
        message={confirmacao?.mensagem}
        confirmLabel={confirmacao?.confirmarLabel}
        loading={confirmLoading}
        onConfirm={() => void executarConfirmacao()}
        onCancel={() => {
          if (!confirmLoading) setConfirmacao(null)
        }}
      />
    </div>
  )
}

export default function AppRoot() {
  return (
    <ToastProvider>
      <App />
    </ToastProvider>
  )
}