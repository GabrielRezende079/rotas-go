import type { GraphNode, RouteRequest, RouteResponse } from '../types'

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

export async function getNodes(): Promise<GraphNode[]> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/nodes`)
  } catch {
    throw new Error('Não foi possível conectar ao servidor. Verifique se o backend está rodando.')
  }
  return handleResponse<GraphNode[]>(response)
}

export async function calculateRoute(req: RouteRequest): Promise<RouteResponse> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/route`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
  } catch {
    throw new Error('Não foi possível conectar ao servidor. Verifique se o backend está rodando.')
  }
  return handleResponse<RouteResponse>(response)
}
