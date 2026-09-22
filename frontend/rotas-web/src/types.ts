export interface LatLng { lat: number; lng: number }
export type Algorithm = 'dijkstra' | 'astar'
export interface RouteRequest { origin: LatLng; destination: LatLng; algorithm: Algorithm }
export interface SnappedPoint extends LatLng { snap_distance_km: number }
export interface RouteResponse {
  origin: SnappedPoint
  destination: SnappedPoint
  algorithm: Algorithm
  distance_km: number
  estimated_duration_minutes: number
  nodes_visited: number
  search_time_ms: number
  geometry: LatLng[]
  data_source: string
}
export interface GraphInfo {
  data_source: string
  generated_at: string
  vertices: number
  edges: number
  real_road_graph: boolean
}
