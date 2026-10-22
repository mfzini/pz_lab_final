import { api } from '@/common'
import { session, setSession, clearSession } from '@/auth'

async function refreshSession() {
  if (!session.refreshToken) return false
  try {
    const r = await fetch(`${api}/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: session.refreshToken }),
    })
    if (!r.ok) return false
    setSession(await r.json())
    return true
  } catch {
    return false
  }
}

/**
 * Fetch contra el backend con token y refresh automático.
 *  - body: objeto a mandar como JSON
 *  - form: FormData (no se toca el Content-Type: lo pone el browser)
 *  - auth: false sólo para endpoints públicos
 * Si el access venció (401) renueva con el refresh token y reintenta una vez;
 * si no se puede, cierra la sesión y manda al login.
 */
export async function apiFetch(path, { method = 'GET', body, form, auth = true } = {}) {
  const headers = {}
  if (form === undefined && body !== undefined) headers['Content-Type'] = 'application/json'
  if (auth && session.token) headers['Authorization'] = `Bearer ${session.token}`
  const payload = form !== undefined ? form : body !== undefined ? JSON.stringify(body) : undefined

  // El body sólo si hay algo que mandar: GET/HEAD no pueden llevar body.
  const send = () => {
    const options = { method, headers }
    if (payload !== undefined) options.body = payload
    return fetch(`${api}${path}`, options)
  }

  let r = await send()
  if (r.status === 401 && auth && session.token) {
    if (await refreshSession()) {
      headers['Authorization'] = `Bearer ${session.token}`
      r = await send()
    } else {
      clearSession()
      window.location.assign('/login')
      throw new Error('Sesión expirada.')
    }
  }

  const text = await r.text()
  let data = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    // respuesta sin JSON
  }
  if (!r.ok) {
    const error = new Error(data?.error || `Error ${r.status}`)
    error.status = r.status
    throw error
  }
  return data
}
