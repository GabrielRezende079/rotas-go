import type { NovoVeiculo, Veiculo } from '../types'

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

export async function createVehicle(req: NovoVeiculo): Promise<Veiculo> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/vehicles`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
  } catch {
    throw new Error('Não foi possível conectar ao servidor. Verifique se o backend está rodando.')
  }
  return handleResponse<Veiculo>(response)
}

export async function listVehicles(): Promise<Veiculo[]> {
  const response = await fetch(`${BASE_URL}/vehicles`)
  const data = await handleResponse<{ vehicles: Veiculo[] }>(response)
  return data.vehicles
}

export async function deleteVehicle(id: number): Promise<void> {
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/vehicles/${id}`, { method: 'DELETE' })
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