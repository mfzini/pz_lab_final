<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api } from '@/common'
import { apiFetch } from '@/api'

const productos = ref([])
const cargando = ref(true)
const error = ref('')
const aviso = ref('')
const page = ref(1)
const totalPages = ref(0)

// Modal de alta/edición: el formulario vive acá, igual que en direcciones.
const abierto = ref(false)
const editId = ref(null) // null = creando; con id = editando
const form = ref({ name: '', price: '', categoryId: null, categoryName: '' }) // precio en pesos (decimal)
const previewUrl = ref('') // objectURL de la foto recién elegida
const actualFoto = ref('') // portrait actual, para mostrarlo si no se cambia

// La categoría se elige con búsqueda asíncrona al servidor (nunca un select
// con todas las categorías de una).
const catSearch = ref('')
const catSugerencias = ref([])
const buscandoCat = ref(false)
let catTimer = null
let catSeq = 0

// Para ofrecer "Crear" sólo cuando todavía no existe ese nombre exacto.
const catExacta = computed(() =>
  catSugerencias.value.some(
    (c) => c.Name.toLowerCase() === catSearch.value.trim().toLowerCase(),
  ),
)

// Foto a mostrar: la elegida ahora, o la que el plato ya tiene.
const fotoSrc = computed(() => {
  if (previewUrl.value) return previewUrl.value
  if (actualFoto.value) return `${api}/${actualFoto.value}`
  return ''
})

async function load(p = 1) {
  cargando.value = true
  error.value = ''
  try {
    const data = await apiFetch(`/platos/list?page=${p}&size=10`, { auth: false })
    productos.value = data.Data
    page.value = data.CurrentPage
    totalPages.value = data.TotalPages
  } catch (e) {
    error.value = e.message
  }
  cargando.value = false
}

// Escape cierra el modal.
function onKey(e) {
  if (e.key === 'Escape' && abierto.value) cerrar()
}

function limpiarPreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

function limpiarBusquedaCat() {
  clearTimeout(catTimer)
  catSearch.value = ''
  catSugerencias.value = []
}

// Debounce de 300ms contra GET /categorias?name=.
function buscarCategoria() {
  const q = catSearch.value.trim()
  clearTimeout(catTimer)
  if (!q) {
    catSugerencias.value = []
    return
  }
  catTimer = setTimeout(async () => {
    const seq = ++catSeq
    buscandoCat.value = true
    try {
      const data = await apiFetch(
        `/categorias?name=${encodeURIComponent(q)}&page=1&size=5`,
        { auth: false },
      )
      if (seq === catSeq) catSugerencias.value = data.Data
    } catch (e) {
      if (seq === catSeq) error.value = e.message
    } finally {
      if (seq === catSeq) buscandoCat.value = false
    }
  }, 300)
}

function elegirCategoria(c) {
  form.value.categoryId = c.Id
  form.value.categoryName = c.Name
  limpiarBusquedaCat()
}

function quitarCategoria() {
  form.value.categoryId = null
  form.value.categoryName = ''
}

// Si no aparece en las sugerencias, se crea al toque y queda elegida.
async function crearCategoria() {
  const name = catSearch.value.trim()
  if (!name) return
  error.value = ''
  aviso.value = ''
  try {
    const cat = await apiFetch('/categorias', { method: 'POST', body: { name } })
    elegirCategoria(cat)
  } catch (e) {
    error.value = e.message
  }
}

onMounted(() => {
  load(1)
  window.addEventListener('keydown', onKey)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKey)
  limpiarPreview()
  clearTimeout(catTimer)
})

function nueva() {
  editId.value = null
  form.value = { name: '', price: '', categoryId: null, categoryName: '' }
  actualFoto.value = ''
  limpiarPreview()
  limpiarBusquedaCat()
  abierto.value = true
}

function edit(p) {
  editId.value = p.Id
  form.value = {
    name: p.Name,
    price: (p.Price / 100).toFixed(2),
    categoryId: p.CategoryId ?? null,
    categoryName: p.CategoryName ?? '',
  }
  actualFoto.value = p.Portrait || ''
  limpiarPreview()
  limpiarBusquedaCat()
  abierto.value = true
}

function cerrar() {
  abierto.value = false
  editId.value = null
  form.value = { name: '', price: '', categoryId: null, categoryName: '' }
  actualFoto.value = ''
  limpiarPreview()
  limpiarBusquedaCat()
}

function onFoto(e) {
  limpiarPreview()
  const file = e.target.files && e.target.files[0]
  if (file) previewUrl.value = URL.createObjectURL(file)
}

async function save() {
  error.value = ''
  aviso.value = ''
  const name = form.value.name.trim()
  const cents = Math.round(Number(form.value.price) * 100)
  if (!name) {
    error.value = 'Falta el nombre.'
    return
  }
  if (!Number.isFinite(cents) || cents < 1) {
    error.value = 'El precio debe ser mayor a 0.'
    return
  }
  const file = document.getElementById('plato-foto').files[0]
  if (!file && !editId.value) {
    error.value = 'Falta la foto.'
    return
  }

  const fd = new FormData()
  fd.append(
    'product',
    JSON.stringify({ name, price: cents, categoryId: form.value.categoryId }),
  )
  if (file) fd.append('picture', file)

  try {
    if (editId.value) {
      await apiFetch(`/platos/${editId.value}`, { method: 'PATCH', form: fd })
      aviso.value = `Plato "${name}" actualizado.`
    } else {
      await apiFetch('/platos', { method: 'POST', form: fd })
      aviso.value = `Plato "${name}" creado.`
    }
    cerrar()
    await load(page.value)
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <div class="columna">
    <h1>Platos</h1>
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="aviso" class="ok">{{ aviso }}</p>

    <div v-if="cargando">Cargando...</div>

    <template v-else>
      <div class="barra">
        <button class="btn" @click="nueva">+ Agregar plato</button>
      </div>

      <ul class="lista">
        <li v-for="p in productos" :key="p.Id" class="card plato-fila">
          <img
            v-if="p.Portrait"
            class="miniatura"
            :src="`${api}/${p.Portrait}`"
            :alt="p.Name"
          >
          <span v-else class="miniatura miniatura-vacia" aria-hidden="true"></span>

          <span class="plato-info">
            <strong>{{ p.Name }}</strong>
            <small>
              ${{ (p.Price / 100).toFixed(2) }}
              <template v-if="p.CategoryName"> · {{ p.CategoryName }}</template>
            </small>
          </span>

          <span class="acciones">
            <button class="btn-ghost" @click="edit(p)">Editar</button>
          </span>
        </li>
        <li v-if="!productos.length" class="vacio">No hay platos cargados.</li>
      </ul>

      <div v-if="totalPages > 1" class="paginacion">
        <button class="btn-ghost" :disabled="page <= 1" @click="load(page - 1)">← Anterior</button>
        <span>Página {{ page }} de {{ totalPages }}</span>
        <button class="btn-ghost" :disabled="page >= totalPages" @click="load(page + 1)">Siguiente →</button>
      </div>
    </template>
  </div>

  <!-- Alta y edición en modal. -->
  <div v-if="abierto" class="modal-overlay" @click.self="cerrar">
    <form class="modal card" @submit.prevent="save">
      <button type="button" class="btn-ghost modal-cerrar" aria-label="Cerrar" @click="cerrar">✕</button>

      <h2>{{ editId ? 'Editar plato' : 'Nuevo plato' }}</h2>

      <div class="campo">
        <label for="plato-nombre">Nombre:</label>
        <input type="text" id="plato-nombre" v-model="form.name" required>
      </div>
      <div class="campo">
        <label for="plato-precio">Precio ($):</label>
        <input type="number" id="plato-precio" v-model="form.price" min="0.01" step="0.01" required>
      </div>

      <!-- Categoría: búsqueda asíncrona, con "Crear" si todavía no existe. -->
      <div class="campo">
        <label>Categoría:</label>
        <div v-if="form.categoryId" class="cat-elegida">
          <span class="badge">{{ form.categoryName }}</span>
          <button type="button" class="btn-ghost" @click="quitarCategoria">Quitar</button>
        </div>
        <template v-else>
          <input
            type="text"
            v-model="catSearch"
            placeholder="Buscar categoría..."
            autocomplete="off"
            @input="buscarCategoria"
          >
          <ul v-if="catSearch.trim()" class="sugerencias">
            <li v-for="c in catSugerencias" :key="c.Id">
              <button type="button" @click="elegirCategoria(c)">{{ c.Name }}</button>
            </li>
            <li v-if="buscandoCat" class="sugerencia-vacia">Buscando...</li>
            <li v-else-if="!catSugerencias.length" class="sugerencia-vacia">
              No existe todavía.
            </li>
          </ul>
          <button
            v-if="catSearch.trim() && !buscandoCat && !catExacta"
            type="button"
            class="btn-ghost crear-cat"
            @click="crearCategoria"
          >+ Crear “{{ catSearch.trim() }}”</button>
        </template>
      </div>

      <div class="campo">
        <label for="plato-foto">
          Foto{{ editId ? ' (vacía = mantener la actual)' : '' }}:
        </label>
        <input type="file" id="plato-foto" accept="image/*" @change="onFoto">
        <div v-if="fotoSrc" class="foto-preview">
          <img :src="fotoSrc" alt="Vista previa de la foto">
        </div>
      </div>

      <div class="modal-actions">
        <button type="submit" class="btn">{{ editId ? 'Guardar' : 'Crear plato' }}</button>
        <button type="button" class="btn-ghost" @click="cerrar">Cancelar</button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.columna {
  width: min(100%, 720px);
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

.plato-fila {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  flex-wrap: wrap;
  padding: 0.6rem 0.85rem;
  margin-bottom: 0.5rem;
}

.miniatura {
  width: 48px;
  height: 48px;
  border-radius: var(--radius);
  object-fit: cover;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

.miniatura-vacia {
  background: var(--bg);
}

.plato-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 8rem;
}

.acciones {
  display: flex;
  gap: 0.5rem;
  margin-left: auto;
}

.vacio {
  color: var(--muted);
}

/* Categoría elegida: chip + botón para sacarla. */
.cat-elegida {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.crear-cat {
  align-self: flex-start;
  margin-top: 0.35rem;
}

/* Foto en el modal: miniatura cuadrada con la imagen actual o la nueva. */
.foto-preview {
  width: 120px;
  height: 120px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}

.foto-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

@media (max-width: 480px) {
  .acciones {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
