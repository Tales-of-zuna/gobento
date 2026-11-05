// Client of the JSON API that the Go server serves under /api/. In
// development, the Vite dev server forwards these requests to the server
// (see vite.config.ts).

export interface Item {
  id: number
  name: string
  created_at: string
}

const apiPrefix = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(apiPrefix + path, init)
  const body: unknown = await res.json().catch(() => null)
  if (!res.ok) {
    const message =
      typeof body === 'object' && body !== null && 'error' in body && typeof body.error === 'string'
        ? body.error
        : `${res.status} ${res.statusText}`
    throw new Error(message)
  }
  return body as T
}

export function listItems(): Promise<Item[]> {
  return request<Item[]>('/items')
}

export function createItem(name: string): Promise<Item> {
  return request<Item>('/items', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
}

export function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
