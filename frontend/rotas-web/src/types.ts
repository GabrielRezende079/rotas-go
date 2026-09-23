export interface LatLng {
  lat: number
  lng: number
}

export type Algorithm = 'dijkstra' | 'astar'

export interface RouteRequest {
  origin: LatLng
  destination: LatLng
  algorithm: Algorithm
  waypoints?: LatLng[]
}

export interface SnappedPoint extends LatLng {
  snap_distance_km: number
}

export interface RouteLeg {
  origin: SnappedPoint
  destination: SnappedPoint
  distance_km: number
  estimated_duration_minutes: number
  nodes_visited: number
  geometry: LatLng[]
}

export interface RouteResponse {
  origin: SnappedPoint
  destination: SnappedPoint
  waypoints: SnappedPoint[]
  algorithm: Algorithm
  distance_km: number
  estimated_duration_minutes: number
  nodes_visited: number
  search_time_ms: number
  geometry: LatLng[]
  legs: RouteLeg[]
  data_source: string
}

export interface VehicleRoute {
  id: string
  label: string
  points: LatLng[]
}

export interface BatchVehicleRequest {
  id: string
  origin: LatLng
  destination: LatLng
  waypoints?: LatLng[]
}

export interface BatchRouteRequest {
  algorithm: Algorithm
  vehicles: BatchVehicleRequest[]
}

export interface BatchRouteItem extends RouteResponse {
  id: string
}

export interface BatchRouteResponse {
  routes: BatchRouteItem[]
}

export interface SavedRouteSummary {
  id: number
  name: string
  algorithm: Algorithm
  created_at: string
  vehicle_count: number
  total_distance_km: number
  total_duration_minutes: number
}

export interface SavedRouteDetail extends SavedRouteSummary {
  routes: BatchRouteItem[]
}

export interface GraphInfo {
  data_source: string
  generated_at: string
  vertices: number
  edges: number
  real_road_graph: boolean
}