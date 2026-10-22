<script setup>
import { ref, onMounted } from 'vue'
import { api } from '@/common'
import { apiFetch } from '@/api'
import { session } from '@/auth'

const users = ref([])
const cargando = ref(true)
const error = ref('')
const aviso = ref('')

const page = ref(1)
const pages = ref(0)

// Nombre del rol en la UI, y el rol al que se pasa al tocar el botón.
const Rol = { admin: 'Administrador', client: 'Cliente' }

async function load(p = page.value) {
  cargando.value = true
  error.value = ''
  try {
    const data = await apiFetch(`/usuarios?page=${p}&size=10`)
    users.value = data.Data
    page.value = data.CurrentPage
    pages.value = data.TotalPages
  } catch (e) {
    error.value = e.message
  }
  cargando.value = false
}

onMounted(() => load())

// PATCH /user/{id}/role — el admin promueve o degrada a un usuario.
// El propio usuario no puede tocarse (el backend también lo rechaza).
async function cambiarRol(u) {
  const nuevo = u.Role === 'admin' ? 'client' : 'admin'
  error.value = ''
  aviso.value = ''
  try {
    await apiFetch(`/user/${u.Id}/role`, {
      method: 'PATCH',
      body: { role: nuevo },
    })
    aviso.value = `${u.Name} ${u.Surname} ahora es ${Rol[nuevo].toLowerCase()}.`
    await load()
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <h1>Usuarios</h1>
  <p class="intro">Asigná el rol de cada usuario. El administrador puede gestionar platos y pedidos; el cliente sólo ve lo suyo.</p>
  <p v-if="error" class="error">{{ error }}</p>
  <p v-if="aviso" class="ok">{{ aviso }}</p>

  <div v-if="cargando">Cargando...</div>

  <template v-else>
    <p v-if="!users.length" class="vacio">No hay usuarios.</p>

    <div v-for="u in users" :key="u.Id" class="card usuario">
      <div class="usuario-avatar" aria-hidden="true">
        <img v-if="u.Avatar" :src="`${api}/${u.Avatar}`" :alt="u.Name">
        <span v-else>{{ (u.Name?.[0] || '?') + (u.Surname?.[0] || '') }}</span>
      </div>
      <div class="usuario-info">
        <strong>{{ u.Name }} {{ u.Surname }}<template v-if="u.Id === session.user?.Id"> (vos)</template></strong>
        <small>{{ u.Email }}</small>
      </div>
      <span class="badge" :class="'badge-' + u.Role">{{ Rol[u.Role] || u.Role }}</span>
      <div class="acciones">
        <button
          class="btn-ghost"
          :disabled="u.Id === session.user?.Id"
          :title="u.Id === session.user?.Id ? 'No podés cambiar tu propio rol' : ''"
          @click="cambiarRol(u)"
        >
          {{ u.Role === 'admin' ? 'Quitar admin' : 'Hacer admin' }}
        </button>
      </div>
    </div>

    <p class="paginacion">
      <button class="btn-ghost" :disabled="page <= 1" @click="load(page - 1)">Anterior</button>
      Página: {{ page }} / {{ pages }}
      <button class="btn-ghost" :disabled="page >= pages" @click="load(page + 1)">Siguiente</button>
    </p>
  </template>
</template>

<style scoped>
.intro {
  color: var(--muted);
  margin-top: -0.25rem;
}

/* Fila de usuario: avatar, datos, badge de rol y acción. */
.usuario {
  display: grid;
  grid-template-columns: 2.75rem 1fr auto auto;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem 1rem;
  margin-bottom: 0.6rem;
}

.usuario-avatar {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 50%;
  overflow: hidden;
  background: var(--bg);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: var(--muted);
}

.usuario-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.usuario-info {
  min-width: 0;
  line-height: 1.25;
}

.usuario-info small {
  display: block;
  color: var(--muted);
}

/* Badge de rol. */
.badge-admin {
  background: #efe7ff;
  color: #6b3fa0;
}

.badge-client {
  background: #e6f2ff;
  color: #1f5fa8;
}

.vacio {
  color: var(--muted);
}

/* Chicas: la fila se apila (avatar + datos arriba, badge y acción abajo). */
@media (max-width: 640px) {
  .usuario {
    grid-template-columns: 2.75rem 1fr;
    grid-template-areas:
      'avatar info'
      'rol acciones';
  }

  .usuario-avatar {
    grid-area: avatar;
  }

  .usuario-info {
    grid-area: info;
  }

  .badge {
    grid-area: rol;
    justify-self: start;
  }

  .acciones {
    grid-area: acciones;
    justify-self: end;
  }
}
</style>
