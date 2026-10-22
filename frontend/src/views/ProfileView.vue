<script setup>
import { ref, computed, onUnmounted } from 'vue'
import { api } from '@/common'
import { apiFetch } from '@/api'
import { session, setSession } from '@/auth'

const error = ref('')
const aviso = ref('')
const subiendo = ref(false)
const form = ref({ name: session.user.Name, surname: session.user.Surname })

// Foto nueva: se previsualiza con un objectURL y recién se sube al clickear.
const nuevaFoto = ref('')

const avatarSrc = computed(() => {
  if (nuevaFoto.value) return nuevaFoto.value
  if (session.user?.Avatar) return `${api}/${session.user.Avatar}`
  return ''
})

const iniciales = computed(() => {
  const n = session.user?.Name?.[0] ?? ''
  const a = session.user?.Surname?.[0] ?? ''
  return (n + a).toUpperCase() || '?'
})

function limpiarPreview() {
  if (nuevaFoto.value) URL.revokeObjectURL(nuevaFoto.value)
  nuevaFoto.value = ''
}

onUnmounted(limpiarPreview)

function onFoto(e) {
  limpiarPreview()
  const file = e.target.files && e.target.files[0]
  if (file) nuevaFoto.value = URL.createObjectURL(file)
}

// La sesión se refresca con los datos del usuario que devuelve el server,
// así el header muestra el avatar recién subido.
function guardarSesion(user) {
  setSession({ token: session.token, refreshToken: session.refreshToken, user })
}

async function subirFoto() {
  const input = document.getElementById('avatar-archivo')
  const file = input?.files?.[0]
  error.value = ''
  aviso.value = ''
  if (!file) {
    error.value = 'Elegí una foto primero.'
    return
  }
  subiendo.value = true
  try {
    const fd = new FormData()
    fd.append('picture', file)
    const res = await apiFetch(`/user/${session.user.Id}/avatar`, {
      method: 'POST',
      form: fd,
    })
    guardarSesion({ ...session.user, Avatar: res.avatar })
    limpiarPreview()
    if (input) input.value = ''
    aviso.value = 'Foto de perfil actualizada.'
  } catch (e) {
    error.value = e.message
  }
  subiendo.value = false
}

async function quitarFoto() {
  error.value = ''
  aviso.value = ''
  try {
    await apiFetch(`/user/${session.user.Id}/avatar`, {
      method: 'PATCH',
      body: { avatar: null },
    })
    guardarSesion({ ...session.user, Avatar: null })
    aviso.value = 'Foto de perfil quitada.'
  } catch (e) {
    error.value = e.message
  }
}

async function guardarDatos() {
  error.value = ''
  aviso.value = ''
  const name = form.value.name.trim()
  const surname = form.value.surname.trim()
  if (!name || !surname) {
    error.value = 'Nombre y apellido no pueden estar vacíos.'
    return
  }
  try {
    const user = await apiFetch(`/user/${session.user.Id}`, {
      method: 'PATCH',
      body: { name, surname },
    })
    guardarSesion(user)
    form.value = { name: user.Name, surname: user.Surname }
    aviso.value = 'Perfil actualizado.'
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <div class="form-column">
    <h1>Mi perfil</h1>
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="aviso" class="ok">{{ aviso }}</p>

    <div class="perfil-avatar">
      <img v-if="avatarSrc" :src="avatarSrc" alt="Foto de perfil">
      <span v-else class="avatar-iniciales" aria-hidden="true">{{ iniciales }}</span>
    </div>

    <div class="campo">
      <label for="avatar-archivo">Foto de perfil:</label>
      <input type="file" id="avatar-archivo" accept="image/*" @change="onFoto">
      <div class="avatar-acciones">
        <button class="btn" :disabled="subiendo || !nuevaFoto" @click="subirFoto">
          {{ subiendo ? 'Subiendo...' : 'Subir foto' }}
        </button>
        <button
          v-if="session.user?.Avatar && !nuevaFoto"
          class="btn-ghost"
          @click="quitarFoto"
        >Quitar foto</button>
      </div>
    </div>

    <div class="campo">
      <label for="perfil-nombre">Nombre:</label>
      <input id="perfil-nombre" type="text" v-model="form.name" required>
    </div>
    <div class="campo">
      <label for="perfil-apellido">Apellido:</label>
      <input id="perfil-apellido" type="text" v-model="form.surname" required>
    </div>
    <div class="campo">
      <label for="perfil-email">Email:</label>
      <input id="perfil-email" type="email" :value="session.user?.Email" disabled>
    </div>
    <div class="campo">
      <label>Rol:</label>
      <span class="badge" :class="session.user?.Role === 'admin' ? 'badge-admin' : 'badge-client'">
        {{ session.user?.Role === 'admin' ? 'Administrador' : 'Cliente' }}
      </span>
    </div>

    <button class="btn perfil-guardar" @click="guardarDatos">Guardar datos</button>
  </div>
</template>

<style scoped>
.perfil-avatar {
  display: flex;
  justify-content: center;
}

.perfil-avatar img,
.avatar-iniciales {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
  background: var(--bg);
}

.avatar-iniciales {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
  color: var(--muted);
}

.avatar-acciones {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.35rem;
}

.perfil-guardar {
  align-self: center;
}

.badge-admin {
  background: #fdf3d7;
  color: #97660d;
}

.badge-client {
  background: #dbeafe;
  color: #1e5b9a;
}
</style>
