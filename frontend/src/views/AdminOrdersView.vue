<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '@/api'

const orders = ref([])
const cargando = ref(true)
const error = ref("")
const aviso = ref("")

// from/to van como "AAAA-MM-DD" (el backend los trata inclusive; hasta = día + 1).
const filtros = ref({ status: "", from: "", to: "" })
const page = ref(1)
const pages = ref(0)

const Estados = { pending: "Pendiente", preparing: "En preparación", shipped: "Despachado" }
const statusOptions = ["pending", "preparing", "shipped"]

async function load(p = page.value) {
  cargando.value = true
  error.value = ""
  try {
    const params = new URLSearchParams({ page: String(p), size: "10" })
    if (filtros.value.status) params.set("status", filtros.value.status)
    if (filtros.value.from) params.set("from", filtros.value.from)
    if (filtros.value.to) params.set("to", filtros.value.to)

    const data = await apiFetch(`/ordenes/admin?${params}`)
    orders.value = data.Data
    page.value = data.CurrentPage
    pages.value = data.TotalPages
  } catch (e) {
    error.value = e.message
  }
  cargando.value = false
}

onMounted(() => load())

function aplicar() {
  load(1)
}

async function cambiarEstado(order, status) {
  error.value = ""
  aviso.value = ""
  try {
    await apiFetch(`/ordenes/${order.Id}/estado`, {
      method: "PATCH",
      body: { status },
    })
    aviso.value = "Estado actualizado."
    await load()
  } catch (e) {
    error.value = e.message
  }
}

function fechaHora(value) {
  return value ? new Date(value).toLocaleString() : "-"
}
</script>

<template>
  <h1>Gestionar pedidos</h1>
  <p v-if="error" class="error">{{ error }}</p>
  <p v-if="aviso" class="ok">{{ aviso }}</p>

  <form class="filtros" @submit.prevent="aplicar">
    <label for="status">Estado:</label>
    <select id="status" v-model="filtros.status">
      <option value="">Todos</option>
      <option v-for="s in statusOptions" :key="s" :value="s">{{ Estados[s] }}</option>
    </select>
    <label for="from">Desde:</label>
    <input type="date" id="from" v-model="filtros.from">
    <label for="to">Hasta:</label>
    <input type="date" id="to" v-model="filtros.to">
    <button type="submit">Buscar</button>
  </form>

  <div v-if="cargando">Cargando...</div>

  <template v-else>
    <p v-if="!orders.length" class="vacio">No hay pedidos con esos filtros.</p>

    <div v-for="o in orders" :key="o.Id" class="card orden">
      <div class="fecha">{{ fechaHora(o.CreatedAt) }}</div>
      <div class="cliente">
        {{ o.Client?.Name }} {{ o.Client?.Surname }}
        <small>{{ o.Client?.Email }}</small>
      </div>
      <div class="estado">
        <span :class="['badge', 'badge-' + o.Status]">{{ Estados[o.Status] || o.Status }}</span>
      </div>
      <div class="total">
        ${{ (o.Total / 100).toFixed(2) }}
        <small>{{ o.Products?.length }} platos</small>
      </div>
      <div v-if="o.Detail" class="detalle">{{ o.Detail }}</div>
      <div class="acciones">
        <button
          v-for="s in statusOptions"
          :key="s"
          class="btn-ghost"
          :disabled="o.Status === s"
          @click="cambiarEstado(o, s)"
        >
          {{ Estados[s] }}
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
.filtros {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 1.25rem;
}

.filtros label {
  font-size: 0.9rem;
  color: var(--muted);
}

/* Grilla con columnas fijas: fecha, cliente, estado y total alineados
   entre filas; detalle y acciones van en fila propia (columna total). */
.orden {
  display: grid;
  grid-template-columns: 10rem 1fr 9.5rem 7rem;
  gap: 0.5rem 1rem;
  align-items: center;
  padding: 0.9rem 1rem;
  margin-bottom: 0.6rem;
}

.fecha {
  grid-column: 1;
  color: var(--muted);
  font-size: 0.9rem;
}

.cliente {
  grid-column: 2;
  min-width: 0;
}

.cliente small {
  display: block;
}

.estado {
  grid-column: 3;
}

.total {
  grid-column: 4;
  font-weight: 700;
  text-align: right;
}

.total small {
  display: block;
  font-weight: 400;
}

.detalle {
  grid-column: 1 / -1;
  font-style: italic;
  color: var(--muted);
}

.acciones {
  grid-column: 1 / -1;
  display: flex;
  gap: 0.5rem;
}

.vacio {
  color: var(--muted);
}

@media (max-width: 800px) {
  .orden {
    grid-template-columns: 1fr 1fr;
  }

  .fecha {
    grid-column: 1;
  }

  .estado {
    grid-column: 2;
    justify-self: end;
  }

  .cliente,
  .total,
  .detalle,
  .acciones {
    grid-column: 1 / -1;
  }

  .total {
    text-align: left;
  }
}

/* Chicas: los filtros se apilan (cada label queda sobre su control). */
@media (max-width: 640px) {
  .filtros {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
