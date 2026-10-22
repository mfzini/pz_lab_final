<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { apiFetch } from '@/api'

// Historial propio: GET /ordenes (el usuario sale del token).
const historial = ref([])
const page = ref(1)
const pages = ref(0)
const cargando = ref(true)
const error = ref('')

// Pedido abierto en el modal de detalle.
const seleccionado = ref(null)

const Estados = { pending: 'Pendiente', preparing: 'En preparación', shipped: 'Despachado' }

async function load(p = page.value) {
  cargando.value = true
  error.value = ''
  try {
    const data = await apiFetch(`/ordenes?page=${p}&size=10`)
    historial.value = data.Data
    page.value = data.CurrentPage
    pages.value = data.TotalPages
  } catch (e) {
    error.value = e.message
  }
  cargando.value = false
}

// Escape cierra el modal.
function onKey(e) {
  if (e.key === 'Escape') seleccionado.value = null
}

onMounted(() => {
  load()
  window.addEventListener('keydown', onKey)
})

onUnmounted(() => window.removeEventListener('keydown', onKey))

function abrir(o) {
  seleccionado.value = o
}

function cerrar() {
  seleccionado.value = null
}

function fecha(value) {
  return value ? new Date(value).toLocaleDateString() : '-'
}

function fechaHora(value) {
  return value ? new Date(value).toLocaleString() : '-'
}

// El backend manda Price como importe de la línea (precio unitario * amount),
// en centavos; Total es la suma de los importes.
function importe(cents) {
  return (cents / 100).toFixed(2)
}
</script>

<template>
  <h1>Mis pedidos</h1>
  <p v-if="error" class="error">{{ error }}</p>

  <div v-if="cargando">Cargando...</div>

  <template v-else>
    <p v-if="!historial.length" class="vacio">
      Todavía no hiciste pedidos: <router-link to="/">mirá el menú</router-link> y hacé el primero.
    </p>

    <div
      v-for="o in historial"
      :key="o.Id"
      class="card pedido"
      role="button"
      tabindex="0"
      @click="abrir(o)"
      @keydown.enter="abrir(o)"
    >
      <span>{{ fecha(o.CreatedAt) }}</span>
      <span :class="['badge', 'badge-' + o.Status]">{{ Estados[o.Status] || o.Status }}</span>
      <span>{{ o.Products?.length }} platos</span>
      <span v-if="o.Detail" class="pedido-detalle">{{ o.Detail }}</span>
      <span class="pedido-total">${{ importe(o.Total) }}</span>
    </div>

    <p class="paginacion">
      <button class="btn-ghost" :disabled="page <= 1" @click="load(page - 1)">Anterior</button>
      Página {{ page }} / {{ pages }}
      <button class="btn-ghost" :disabled="page >= pages" @click="load(page + 1)">Siguiente</button>
    </p>
  </template>

  <!-- Detalle del pedido: platos, subtotales y total. -->
  <div v-if="seleccionado" class="modal-overlay" @click.self="cerrar">
    <div class="modal card" role="dialog" aria-modal="true" aria-label="Detalle del pedido">
      <button class="btn-ghost modal-cerrar" aria-label="Cerrar" @click="cerrar">✕</button>

      <h2>Pedido del {{ fechaHora(seleccionado.CreatedAt) }}</h2>
      <p>
        <span :class="['badge', 'badge-' + seleccionado.Status]">
          {{ Estados[seleccionado.Status] || seleccionado.Status }}
        </span>
      </p>

      <table class="detalle-tabla">
        <thead>
          <tr>
            <th>Plato</th>
            <th>Cant.</th>
            <th>Subtotal</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in seleccionado.Products" :key="p.Id">
            <td>{{ p.Name }}</td>
            <td>{{ p.Amount }}</td>
            <td>${{ importe(p.Price) }}</td>
          </tr>
        </tbody>
      </table>

      <p v-if="seleccionado.Detail" class="modal-detalle">Detalle: {{ seleccionado.Detail }}</p>

      <p class="modal-total">
        <span>Total</span>
        <strong>${{ importe(seleccionado.Total) }}</strong>
      </p>
    </div>
  </div>
</template>

<style scoped>
/* Grilla con columnas fijas: las filas se alinean entre sí aunque una
   tenga detalle y otra no (detalle y total tienen columna explícita). */
.pedido {
  display: grid;
  grid-template-columns: 7rem 9.5rem 6.5rem 1fr auto;
  gap: 0.5rem 1rem;
  align-items: center;
  padding: 0.75rem 1rem;
  margin-bottom: 0.5rem;
  cursor: pointer;
}

.pedido:hover {
  background: #faf8f4;
}

.pedido-detalle {
  grid-column: 4;
  font-style: italic;
  color: var(--muted);
}

.pedido-total {
  grid-column: 5;
  justify-self: end;
  font-weight: 700;
}

.vacio {
  color: var(--muted);
}

@media (max-width: 700px) {
  .pedido {
    grid-template-columns: max-content max-content 1fr;
  }

  .pedido-detalle,
  .pedido-total {
    grid-column: 1 / -1;
    justify-self: start;
  }
}

/* ---------- Modal de detalle ---------- */

.detalle-tabla {
  width: 100%;
  border-collapse: collapse;
  margin: 0.75rem 0;
}

.detalle-tabla th,
.detalle-tabla td {
  text-align: left;
  padding: 0.4rem 0;
  border-bottom: 1px solid var(--border);
}

.detalle-tabla th {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--muted);
}

.detalle-tabla td:nth-child(2),
.detalle-tabla th:nth-child(2) {
  text-align: center;
}

.detalle-tabla td:last-child,
.detalle-tabla th:last-child {
  text-align: right;
}

.modal-detalle {
  color: var(--muted);
  font-style: italic;
}

.modal-total {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 1.1rem;
  margin-bottom: 0;
}
</style>
