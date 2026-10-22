<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { apiFetch } from '@/api'
import { session } from '@/auth'

const addresses = ref([])
const cargando = ref(true)
const error = ref('')
const aviso = ref('')

// Modal de alta/edición: el formulario vive acá, no en la vista.
const abierto = ref(false)

// Form vacío: null en los campos anulables (floor, door, coordinates, detail).
const emptyForm = () => ({
  street: '',
  house: '',
  floor: '',
  door: '',
  coordinates: '',
  detail: '',
})
const form = ref(emptyForm())
const editId = ref(null) // null = creando; con id = editando

async function load() {
  cargando.value = true
  error.value = ''
  try {
    const data = await apiFetch(`/usuarios/${session.user.Id}/direcciones?page=1&size=50`)
    addresses.value = data.Data
  } catch (e) {
    error.value = e.message
  }
  cargando.value = false
}

// Escape cierra el modal.
function onKey(e) {
  if (e.key === 'Escape') cerrar()
}

onMounted(() => {
  load()
  window.addEventListener('keydown', onKey)
})

onUnmounted(() => window.removeEventListener('keydown', onKey))

function nueva() {
  editId.value = null
  form.value = emptyForm()
  abierto.value = true
}

function edit(a) {
  editId.value = a.Id
  form.value = {
    street: a.Street,
    house: a.House,
    floor: a.Floor ?? '',
    door: a.Door ?? '',
    coordinates: a.Coordinates ?? '',
    detail: a.Detail ?? '',
  }
  abierto.value = true
}

function cerrar() {
  abierto.value = false
  editId.value = null
  form.value = emptyForm()
}

function payload() {
  return {
    street: form.value.street,
    house: form.value.house,
    floor: form.value.floor === '' ? null : Number(form.value.floor),
    door: form.value.door || null,
    coordinates: form.value.coordinates || null,
    detail: form.value.detail || null,
  }
}

async function save() {
  error.value = ''
  aviso.value = ''
  try {
    if (editId.value) {
      await apiFetch(`/direcciones/${editId.value}`, { method: 'PATCH', body: payload() })
      aviso.value = 'Dirección actualizada.'
    } else {
      await apiFetch('/direcciones', { method: 'POST', body: payload() })
      aviso.value = 'Dirección creada.'
    }
    cerrar()
    await load()
  } catch (e) {
    error.value = e.message
  }
}

async function remove(id) {
  error.value = ''
  aviso.value = ''
  try {
    await apiFetch(`/direcciones/${id}`, { method: 'DELETE' })
    if (editId.value === id) cerrar()
    await load()
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <!-- Toda la vista en una columna centrada, como login y registro. -->
  <div class="columna">
    <h1>Mis direcciones</h1>
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="aviso" class="ok">{{ aviso }}</p>

    <div v-if="cargando">Cargando...</div>

    <template v-else>
      <div class="barra">
        <button class="btn" @click="nueva">+ Agregar dirección</button>
      </div>

      <ul class="lista">
        <li v-for="a in addresses" :key="a.Id" class="card direccion">
          <span class="direccion-texto">
            {{ a.Street }} {{ a.House }}
            <template v-if="a.Floor !== null"> · Piso {{ a.Floor }}</template>
            <template v-if="a.Door"> · Depto {{ a.Door }}</template>
            <small v-if="a.Detail"> — {{ a.Detail }}</small>
          </span>
          <span class="acciones">
            <button class="btn-ghost" @click="edit(a)">Editar</button>
            <button class="btn-ghost" @click="remove(a.Id)">Eliminar</button>
          </span>
        </li>
        <li v-if="!addresses.length" class="vacio">Todavía no tenés direcciones cargadas.</li>
      </ul>
    </template>
  </div>

  <!-- Alta y edición en un modal, para que la vista quede sólo la lista. -->
  <div v-if="abierto" class="modal-overlay" @click.self="cerrar">
    <form class="modal card" @submit.prevent="save">
      <button type="button" class="btn-ghost modal-cerrar" aria-label="Cerrar" @click="cerrar">✕</button>

      <h2>{{ editId ? 'Editar dirección' : 'Nueva dirección' }}</h2>

      <div class="campo">
        <label for="street">Calle:</label>
        <input type="text" id="street" v-model="form.street" required>
      </div>
      <div class="campo">
        <label for="house">Altura:</label>
        <input type="text" id="house" v-model="form.house" required>
      </div>
      <div class="campo">
        <label for="floor">Piso:</label>
        <input type="number" id="floor" v-model="form.floor">
      </div>
      <div class="campo">
        <label for="door">Depto:</label>
        <input type="text" id="door" v-model="form.door">
      </div>
      <div class="campo">
        <label for="detail">Detalle:</label>
        <input type="text" id="detail" v-model="form.detail" placeholder="Timbre roto...">
      </div>
      <div class="campo">
        <label for="coordinates">Coordenadas:</label>
        <input type="text" id="coordinates" v-model="form.coordinates" placeholder="(-58.38,-34.60)">
      </div>

      <div class="modal-actions">
        <button type="submit" class="btn">{{ editId ? 'Guardar' : 'Crear dirección' }}</button>
        <button type="button" class="btn-ghost" @click="cerrar">Cancelar</button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.columna {
  width: min(100%, 640px);
  margin: 0 auto;
}

.barra {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
}

.lista {
  list-style: none;
  margin: 0;
  padding: 0;
}

.direccion {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 0.75rem 1rem;
  margin-bottom: 0.5rem;
}

.vacio {
  color: var(--muted);
}

.acciones {
  display: flex;
  gap: 0.5rem;
  /* En pantallas chicas la fila se parte: los botones quedan a la derecha. */
  margin-left: auto;
}

@media (max-width: 480px) {
  .direccion-texto {
    width: 100%;
  }
}
</style>
