import { useEffect, useMemo } from 'react'
import type { ReactNode } from 'react'
import 'leaflet/dist/leaflet.css'
import * as L from 'leaflet'
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'
import {
  MapContainer,
  Marker,
  Polyline,
  TileLayer,
  Tooltip,
  useMap,
  useMapEvents,
} from 'react-leaflet'
import type { Base, LatLng, RouteResponse, VehicleRoute } from '../types'
import type { AltState } from './AltStepper'

L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
})

const CORES = [
  '#2563eb',
  '#7c3aed',
  '#059669',
  '#ea580c',
  '#0d9488',
  '#be123c',
  '#4f46e5',
  '#ca8a04',
]

const COR_ORIGEM = '#16a34a'
const COR_PARADA = '#f59e0b'
const COR_DESTINO = '#dc2626'

const CORES_OPCOES = ['#2563eb', '#7c3aed', '#059669']

// iconoNumerado cria um marcador circular com o número da parada dentro.
function iconoNumerado(numero: number, cor: string) {
  return L.divIcon({
    className: '',
    html: `<div class="map-marker-number" style="background-color:${cor}">${numero}</div>`,
    iconSize: [22, 22],
    iconAnchor: [11, 11],
  })
}

// iconoBase cria o marcador das bases (localizações padrão).
function iconoBase() {
  return L.divIcon({
    className: 'base-marker',
    html: '<div class="base-marker-icon">B</div>',
    iconSize: [26, 26],
    iconAnchor: [13, 13],
  })
}

function corDoPonto(index: number, total: number): string {
  if (index === 0) return COR_ORIGEM
  if (index === total - 1) return COR_DESTINO
  return COR_PARADA
}

interface MapClickHandlerProps {
  onMapClick: (position: LatLng) => void
}

function MapClickHandler({ onMapClick }: MapClickHandlerProps) {
  useMapEvents({
    click: (event) => {
      const alvo = event.originalEvent.target as HTMLElement | null
      if (alvo?.closest('.base-marker') != null) return
      onMapClick({ lat: event.latlng.lat, lng: event.latlng.lng })
    },
  })
  return null
}

function FitRoute({ geometries }: { geometries: LatLng[][] }) {
  const map = useMap()
  useEffect(() => {
    const pontos: [number, number][] = []
    for (const geometry of geometries) {
      for (const ponto of geometry) {
        pontos.push([ponto.lat, ponto.lng])
      }
    }
    if (pontos.length > 1) {
      map.fitBounds(pontos, { padding: [32, 32] })
    }
  }, [geometries, map])
  return null
}

interface RouteMapProps {
  vehicles: VehicleRoute[]
  activeVehicleId: string | null
  results: Record<string, RouteResponse> | null
  alt: AltState
  altVehicleId: string | null
  bases: Base[]
  onMapClick: (position: LatLng) => void
  onPointDrag: (vehicleId: string, index: number, position: LatLng) => void
  onSelectAlternative: (step: number, alternativeIndex: number) => void
  onSelectBase: (base: Base) => void
}

function RouteMap({
  vehicles,
  activeVehicleId,
  results,
  alt,
  altVehicleId,
  bases,
  onMapClick,
  onPointDrag,
  onSelectAlternative,
  onSelectBase,
}: RouteMapProps) {
  const ativo = vehicles.find((v) => v.id === activeVehicleId) ?? null

  const geometriasParaFocus = useMemo(() => {
    const lista: LatLng[][] = []
    if (alt.active && alt.legs !== null) {
      for (const leg of alt.legs) {
        for (const alternativa of leg.alternatives) {
          if (alternativa.geometry.length > 0) lista.push(alternativa.geometry)
        }
      }
      return lista
    }
    if (results === null) return lista
    for (const vehicle of vehicles) {
      const resultado = results[vehicle.id]
      if (resultado === undefined) continue
      if (resultado.legs.length > 0) {
        for (const leg of resultado.legs) {
          if (leg.geometry.length > 0) lista.push(leg.geometry)
        }
      } else if (resultado.geometry.length > 0) {
        lista.push(resultado.geometry)
      }
    }
    return lista
  }, [results, vehicles, alt])

  const polilinhasAlternativas = (() => {
    if (!alt.active || alt.legs === null || altVehicleId === null) return null
    const trechos: ReactNode[] = []
    alt.legs.forEach((leg, legIndex) => {
      const escolha = alt.selections[legIndex]
      if (escolha !== undefined) {
        const alternativa = leg.alternatives[escolha]
        if (alternativa !== undefined && alternativa.geometry.length > 0) {
          trechos.push(
            <Polyline
              key={`alt-${legIndex}-escolhida`}
              positions={alternativa.geometry.map((point) => [point.lat, point.lng])}
              pathOptions={{
                color: CORES[legIndex % CORES.length],
                weight: 5,
                opacity: 0.95,
              }}
            />,
          )
        }
        return
      }
      if (legIndex !== alt.step) return
      leg.alternatives.forEach((alternativa, opcaoIndex) => {
        if (alternativa.geometry.length === 0) return
        trechos.push(
          <Polyline
            key={`alt-${legIndex}-opcao-${opcaoIndex}`}
            positions={alternativa.geometry.map((point) => [point.lat, point.lng])}
            pathOptions={{
              color: CORES_OPCOES[opcaoIndex % CORES_OPCOES.length],
              weight: 5,
              dashArray: '8 8',
              opacity: 0.9,
            }}
            eventHandlers={{
              click: () => onSelectAlternative(legIndex, opcaoIndex),
            }}
          />,
        )
      })
      trechos.push(
        <Marker
          key={`alt-${legIndex}-origem`}
          position={[leg.origin.lat, leg.origin.lng]}
          icon={iconoNumerado(legIndex + 1, COR_ORIGEM)}
        />,
      )
      trechos.push(
        <Marker
          key={`alt-${legIndex}-destino`}
          position={[leg.destination.lat, leg.destination.lng]}
          icon={iconoNumerado(legIndex + 2, COR_DESTINO)}
        />,
      )
    })
    if (trechos.length === 0) return null
    return trechos
  })()

  return (
    <div className="map-container">
      <MapContainer center={[-19.6, -40.65]} zoom={8}>
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapClickHandler onMapClick={onMapClick} />
        {polilinhasAlternativas !== null
          ? polilinhasAlternativas
          : vehicles.map((vehicle, vehicleIndex) => {
              const resultado = results?.[vehicle.id]
              if (resultado === undefined) return null
              if (resultado.legs.length > 0) {
                return resultado.legs.map((leg, legIndex) => (
                  <Polyline
                    key={`${vehicle.id}-leg-${legIndex}`}
                    positions={leg.geometry.map((point) => [point.lat, point.lng])}
                    pathOptions={{ color: CORES[(vehicleIndex + legIndex) % CORES.length], weight: 5 }}
                  />
                ))
              }
              return (
                <Polyline
                  key={`${vehicle.id}-route`}
                  positions={resultado.geometry.map((point) => [point.lat, point.lng])}
                  pathOptions={{ color: CORES[vehicleIndex % CORES.length], weight: 5 }}
                />
              )
            })}
        {bases.map((base) => (
          <Marker
            key={`base-${base.id}`}
            position={[base.lat, base.lng]}
            icon={iconoBase()}
            eventHandlers={{
              click: () => onSelectBase(base),
            }}
          >
            <Tooltip>
              <strong>{base.name}</strong>
              <span> clicar adiciona como ponto da rota</span>
            </Tooltip>
          </Marker>
        ))}
        {ativo?.points.map((point, index) => (
          <Marker
            key={`${ativo.id}-${index}`}
            position={[point.lat, point.lng]}
            icon={iconoNumerado(index + 1, corDoPonto(index, ativo.points.length))}
            draggable
            eventHandlers={{
              dragend: (event) => {
                const position = event.target.getLatLng()
                onPointDrag(ativo.id, index, { lat: position.lat, lng: position.lng })
              },
            }}
          />
        ))}
        <FitRoute geometries={geometriasParaFocus} />
      </MapContainer>
    </div>
  )
}

export default RouteMap