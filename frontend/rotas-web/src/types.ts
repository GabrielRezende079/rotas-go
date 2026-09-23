export interface LatLng {
  lat: number
  lng: number
}

export type Algorithm = 'dijkstra' | 'astar'

export type Cost = 'duration' | 'distance'

export interface RouteRequest {
  origin: LatLng
  destination: LatLng
  algorithm: Algorithm
  waypoints?: LatLng[]
  cost?: Cost
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
  cost?: Cost
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

export interface AlternativesRequest {
  origin: LatLng
  destination: LatLng
  waypoints?: LatLng[]
  cost?: Cost
  max_alternatives?: number
}

export interface RouteAlternative {
  distance_km: number
  estimated_duration_minutes: number
  nodes_visited: number
  total_cost: number
  geometry: LatLng[]
}

export interface LegAlternatives {
  origin: SnappedPoint
  destination: SnappedPoint
  alternatives: RouteAlternative[]
}

export interface AlternativesResponse {
  cost: Cost
  legs: LegAlternatives[]
}

export interface Base {
  id: number
  name: string
  lat: number
  lng: number
  created_at: string
}