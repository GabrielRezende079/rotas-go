import { useEffect } from 'react'
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
  useMap,
  useMapEvents,
} from 'react-leaflet'
import type { LatLng, RouteResponse } from '../types'

L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
})

const originIcon = L.divIcon({
  className: '',
  html: '<div class="map-marker-dot map-marker-dot-origin"></div>',
  iconSize: [18, 18],
  iconAnchor: [9, 9],
})

const destinationIcon = L.divIcon({
  className: '',
  html: '<div class="map-marker-dot map-marker-dot-destination"></div>',
  iconSize: [18, 18],
  iconAnchor: [9, 9],
})

interface MapClickHandlerProps {
  origin: LatLng | null
  destination: LatLng | null
  onOriginChange: (origin: LatLng) => void
  onDestinationChange: (destination: LatLng) => void
  onClear: () => void
}

function MapClickHandler({
  origin,
  destination,
  onOriginChange,
  onDestinationChange,
  onClear,
}: MapClickHandlerProps) {
  useMapEvents({
    click: (event) => {
      const position: LatLng = { lat: event.latlng.lat, lng: event.latlng.lng }
      if (origin === null) {
        onOriginChange(position)
      } else if (destination === null) {
        onDestinationChange(position)
      } else {
        onClear()
        onOriginChange(position)
      }
    },
  })
  return null
}

function FitRoute({ geometry }: { geometry: LatLng[] }) {
  const map = useMap()
  useEffect(() => {
    if (geometry.length > 1) {
      map.fitBounds(
        geometry.map((point) => [point.lat, point.lng] as [number, number]),
        { padding: [32, 32] },
      )
    }
  }, [geometry, map])
  return null
}

interface RouteMapProps {
  origin: LatLng | null
  destination: LatLng | null
  response: RouteResponse | null
  onOriginChange: (origin: LatLng) => void
  onDestinationChange: (destination: LatLng) => void
  onClear: () => void
}

function RouteMap({
  origin,
  destination,
  response,
  onOriginChange,
  onDestinationChange,
  onClear,
}: RouteMapProps) {
  return (
    <div className="map-container">
      <MapContainer center={[-19.6, -40.65]} zoom={8}>
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapClickHandler
          origin={origin}
          destination={destination}
          onOriginChange={onOriginChange}
          onDestinationChange={onDestinationChange}
          onClear={onClear}
        />
        {origin !== null && (
          <Marker
            position={[origin.lat, origin.lng]}
            icon={originIcon}
            draggable
            eventHandlers={{
              dragend: (event) => {
                const position = event.target.getLatLng()
                onOriginChange({ lat: position.lat, lng: position.lng })
              },
            }}
          />
        )}
        {destination !== null && (
          <Marker
            position={[destination.lat, destination.lng]}
            icon={destinationIcon}
            draggable
            eventHandlers={{
              dragend: (event) => {
                const position = event.target.getLatLng()
                onDestinationChange({ lat: position.lat, lng: position.lng })
              },
            }}
          />
        )}
        {response !== null && (
          <>
            <Polyline
              positions={response.geometry.map((point) => [point.lat, point.lng])}
              pathOptions={{ color: '#2563eb', weight: 5 }}
            />
            <FitRoute geometry={response.geometry} />
          </>
        )}
      </MapContainer>
    </div>
  )
}

export default RouteMap
