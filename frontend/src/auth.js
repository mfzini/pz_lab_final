import { reactive, computed } from 'vue'
import { api } from '@/common'

// Un localStorage sucio no tiene que romper el arranque de la app.
function loadStoredUser() {
  try {
    return JSON.parse(localStorage.getItem('user') || 'null')
  } catch {
    return null
  }
}

// Sesión guardada en localStorage (diseño del proyecto): access + refresh +
// usuario. El access viaja en cada request; el refresh sólo para renovarlo.
export const session = reactive({
  token: localStorage.getItem('token') || '',
  refreshToken: localStorage.getItem('refreshToken') || '',
  user: loadStoredUser(),
})

function persist() {
  if (session.token) {
    localStorage.setItem('token', session.token)
    localStorage.setItem('refreshToken', session.refreshToken)
    localStorage.setItem('user', JSON.stringify(session.user))
  } else {
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
  }
}

// setSession recibe la respuesta de /login o /refresh:
// { token, refreshToken, user }
export function setSession(data) {
  session.token = data.token
  session.refreshToken = data.refreshToken
  session.user = data.user
  persist()
}

export function clearSession() {
  session.token = ''
  session.refreshToken = ''
  session.user = null
  persist()
}

export const isAuthenticated = () => !!session.token
export const isAdmin = computed(() => session.user?.Role === 'admin')

async function post(path, body) {
  const r = await fetch(`${api}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const text = await r.text()
  let data = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    // respuesta sin JSON
  }
  if (!r.ok) throw new Error(data?.error || `Error ${r.status}`)
  return data
}

export async function login(email, password) {
  setSession(await post('/login', { email, password }))
}

// POST /user devuelve 201 sin body; con las mismas credenciales entramos.
export async function register(payload) {
  await post('/user', payload)
  await login(payload.email, payload.password)
}
