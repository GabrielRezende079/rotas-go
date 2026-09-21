export interface LatLng { lat: number; lng: number }
export interface GraphNode extends LatLng { name: string }
export type Algorithm = 'dijkstra' | 'astar'
export interface RouteRequest { origin: LatLng; destination: LatLng; algorithm: Algorithm }
export interface RouteResponse { origin: GraphNode; destination: GraphNode; algorithm: Algorithm; distance_km: number; nodes_visited: number; path: GraphNode[] }
