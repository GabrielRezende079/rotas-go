import type {
  AlternativesRequest,
  AlternativesResponse,
  Base,
  BatchRouteRequest,
  BatchRouteResponse,
  GraphInfo,
  RouteRequest,
  RouteResponse,
  SavedRouteDetail,
  SavedRouteSummary,
} from '../types'

const BASE_URL = '/api/v1'

async function readJson(response: Response): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    return null
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  const data = await readJson(response)
  if (!response.ok) {
    if (data !== null && typeof data === 'object' && 'error' in data) {
      throw new Error(String(data.error))
    }
    throw new Error(`Erro HTTP ${response.status}`)
  }
  if (data === null) {
    throw new Error('Resposta vazia do servidor')
  }
  return data as T
}

async function postJson<T>(path: string, body: unknown): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
  } catch {
    throw new Error('Não foi possível conectar ao servidor. Verifique se o backend está rodando.')
  }
  return handleResponse<T>(response)
}

export async function getGraphInfo(): Promise<GraphInfo> {
  const response = await fetch(`${BASE_URL}/info`)
  return handleResponse<GraphInfo>(response)
}

export async function calculateRoute(req: RouteRequest): Promise<RouteResponse> {
  return postJson<RouteResponse>('/route', req)
}

export async function calculateBatch(req: BatchRouteRequest): Promise<BatchRouteResponse> {
  return postJson<BatchRouteResponse>('/routes', req)
}

export async function calculateAlternatives(
  req: AlternativesRequest,
): Promise<AlternativesResponse> {
  return postJson<AlternativesResponse>('/route/alternatives', req)
}

export async function saveRoute(req: {
  name: string
  algorithm: RouteRequest['algorithm']
  routes: BatchRouteResponse['routes']
}): Promise<SavedRouteSummary> {
  return postJson<SavedRouteSummary>('/routes/saved', req)
}

export async function listSavedRoutes(): Promise<SavedRouteSummary[]> {
  const response = await fetch(`${BASE_URL}/routes/saved`)
  return handleResponse<SavedRouteSummary[]>(response)
}

export async function getSavedRoute(id: number): Promise<SavedRouteDetail> {
  const response = await fetch(`${BASE_URL}/routes/saved/${id}`)
  return handleResponse<SavedRouteDetail>(response)
}

export async function deleteSavedRoute(id: number): Promise<void> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/routes/saved/${id}`, { method: 'DELETE' })
  } catch {
    throw new Error('Não foi possível conectar ao servidor. Verifique se o backend está rodando.')
  }
  if (!response.ok) {
    const data = await readJson(response)
    if (data !== null && typeof data === 'object' && 'error' in data) {
      throw new Error(String(data.error))
    }
    throw new Error(`Erro HTTP ${response.status}`)
  }
}

export async function createBase(req: { name: string; lat: number; lng: number }): Promise<Base> {
  return postJson<Base>('/bases', req)
}

export async function listBases(): Promise<Base[]> {
  const response = await fetch(`${BASE_URL}/bases`)
  const data = await handleResponse<{ bases: Base[] }>(response)
  return data.bases
}

export async function deleteBase(id: number): Promise<void> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/bases/${id}`, { method: 'DELETE' })
  } catch {
    throw new Error('Não foi possível conectar ao servidor. Verifique se o backend está rodando.')
  }
  if (!response.ok) {
    const data = await readJson(response)
    if (data !== null && typeof data === 'object' && 'error' in data) {
      throw new Error(String(data.error))
    }
    throw new Error(`Erro HTTP ${response.status}`)
  }
}