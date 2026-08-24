const API = (import.meta.env.VITE_API_BASE as string) || '/api'

export type Envelope<T> = { code: number; message: string; data: T; trace_id: string }

let token = localStorage.getItem('gc_token') || ''

export function setToken(t: string) {
  token = t
  if (t) localStorage.setItem('gc_token', t)
  else localStorage.removeItem('gc_token')
}

export function getToken() { return token }

async function req<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(init.headers as any) }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${API}${path}`, { ...init, headers })
  const json = (await res.json()) as Envelope<T>
  if (json.code !== 0) {
    const err = new Error(json.message) as Error & { code: number }
    err.code = json.code
    throw err
  }
  return json.data
}

export const api = {
  login: (username: string, password: string) =>
    req<{ token: string; user: any }>('/v1/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  register: (username: string, password: string, nickname: string) =>
    req<{ token: string; user: any }>('/v1/auth/register', { method: 'POST', body: JSON.stringify({ username, password, nickname }) }),
  me: () => req<any>('/v1/me'),
  routes: (scope?: string) => req<any[]>(`/v1/route-books${scope ? '?scope=' + scope : ''}`),
  getRoute: (id: number) => req<any>(`/v1/route-books/${id}`),
  saveRoute: (body: any) => req<any>('/v1/route-books', { method: 'POST', body: JSON.stringify(body) }),
  updateRoute: (id: number, body: any) => req<any>(`/v1/route-books/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  elevation: (id: number) => req<any>(`/v1/route-books/${id}/elevation`, { method: 'POST', body: '{}' }),
  teams: () => req<any[]>('/v1/teams'),
  createTeam: (name: string, route_id?: number) =>
    req<any>('/v1/teams', { method: 'POST', body: JSON.stringify({ name, route_id }) }),
  joinTeam: (code: string) => req<any>('/v1/teams/join', { method: 'POST', body: JSON.stringify({ code }) }),
  getTeam: (id: number) => req<any>(`/v1/teams/${id}`),
  createTrip: (team_id: number) => req<any>('/v1/trips', { method: 'POST', body: JSON.stringify({ team_id }) }),
  getTrip: (id: number) => req<any>(`/v1/trips/${id}`),
  listTrips: (teamId: number) => req<any[]>(`/v1/teams/${teamId}/trips`),
  startTrip: (id: number) => req<any>(`/v1/trips/${id}/start`, { method: 'POST', body: '{}' }),
  pauseTrip: (id: number) => req<any>(`/v1/trips/${id}/pause`, { method: 'POST', body: '{}' }),
  finishTrip: (id: number) => req<any>(`/v1/trips/${id}/finish`, { method: 'POST', body: '{}' }),
  position: (id: number, lat: number, lon: number) =>
    req<any>(`/v1/trips/${id}/positions`, { method: 'POST', body: JSON.stringify({ lat, lon }) }),
  batch: (id: number, points: any[]) =>
    req<any>(`/v1/trips/${id}/tracks/batch`, { method: 'POST', body: JSON.stringify({ points }) }),
  tracks: (id: number) => req<any[]>(`/v1/trips/${id}/tracks`),
  eta: (id: number) => req<any>(`/v1/trips/${id}/eta`),
  sos: (id: number, lat: number, lon: number, reason: string) =>
    req<any>(`/v1/trips/${id}/sos`, { method: 'POST', body: JSON.stringify({ lat, lon, reason }) }),
  listSos: (id: number) => req<any[]>(`/v1/trips/${id}/sos`),
  resolveSos: (id: number) => req<any>(`/v1/sos/${id}/resolve`, { method: 'POST', body: '{}' }),
  risk: (id: number) => req<any>(`/v1/trips/${id}/risk`),
  backtrack: (id: number) => req<any>(`/v1/trips/${id}/backtrack`),
  simulate: (id: number) => req<any>(`/v1/trips/${id}/simulate`, { method: 'POST', body: '{}' }),
}

export function wsURL(tripId: number) {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${location.host}/api/v1/ws?trip_id=${tripId}&token=${encodeURIComponent(token)}`
}

export const TILE_URL = (import.meta.env.VITE_TILE_URL as string) || '/tiles/{z}/{x}/{y}.png'
