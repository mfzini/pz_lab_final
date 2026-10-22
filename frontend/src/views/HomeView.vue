<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '@/common'
import { apiFetch } from '@/api'
import { session, isAuthenticated } from '@/auth'
import { cartItems, totalCount, totalPrice, addToCart, removeItem, clearCart } from '@/order'

// Catálogo público: no manda token (auth: false).
const Products = ref([])
const CurrentPage = ref(1)
const TotalPages = ref(0)
const cargando = ref(true)

// Checkout (sólo con sesión): dirección de entrega y detalle.
const addressId = ref('')
const detail = ref('')

// La dirección se elige con búsqueda asíncrona (el servidor filtra por
// calle): nunca se traen todas las direcciones de una. Al enfocar el campo ya
// se listan las del usuario; al tipear se filtran con debounce.
const dirSearch = ref('')
const dirElegida = ref(null)
const sugerencias = ref([])
const buscandoDir = ref(false)
const dirAbierta = ref(false)
let dirTimer = null
let dirCloseTimer = null
let dirSeq = 0

const error = ref('')

// Modal de pedido exitoso: se cierra solo a los 4 segundos.
const exito = ref(false)
let timerExito = null

const setError = (e) => { error.value = e.message }

async function loadCatalog(page = 1) {
  const data = await apiFetch(`/platos/list?page=${page}&size=8`, { auth: false })
  Products.value = data.Data
  CurrentPage.value = data.CurrentPage
  TotalPages.value = data.TotalPages
}

// Pide las direcciones al servidor. Con street vacío trae las primeras (para
// el focus); con street filtra por calle (ilike). seq desempata respuestas
// fuera de orden.
async function pedirDirecciones(street) {
  const seq = ++dirSeq
  buscandoDir.value = true
  try {
    const data = await apiFetch(
      `/usuarios/${session.user.Id}/direcciones?street=${encodeURIComponent(street)}&page=1&size=6`,
    )
    if (seq === dirSeq) sugerencias.value = data.Data
  } catch (e) {
    if (seq === dirSeq) setError(e)
  } finally {
    if (seq === dirSeq) buscandoDir.value = false
  }
}

// Al enfocar: se abren las sugerencias con las direcciones del usuario, sin
// tener que tipear nada (eso es lo que se espera de un selector).
function onDirFocus() {
  clearTimeout(dirCloseTimer)
  dirAbierta.value = true
  if (!dirSearch.value.trim()) pedirDirecciones('')
}

// Al tipear: debounce de 300ms contra ?street=. El valor sale del evento (no
// del v-model) para no depender del orden del handler con v-model.
function onDirInput(e) {
  const q = (e.target.value || '').trim()
  clearTimeout(dirTimer)
  clearTimeout(dirCloseTimer)
  dirAbierta.value = true
  if (!q) {
    pedirDirecciones('')
    return
  }
  dirTimer = setTimeout(() => pedirDirecciones(q), 300)
}

// Al salir del campo: se cierra sola, con un delay que deja alcanzar el
// click sobre una sugerencia.
function onDirBlur() {
  clearTimeout(dirTimer)
  dirCloseTimer = setTimeout(() => {
    dirAbierta.value = false
    sugerencias.value = []
  }, 180)
}

function elegirDireccion(a) {
  clearTimeout(dirTimer)
  clearTimeout(dirCloseTimer)
  addressId.value = a.Id
  dirElegida.value = a
  dirSearch.value = ''
  sugerencias.value = []
  dirAbierta.value = false
}

function cambiarDireccion() {
  addressId.value = ''
  dirElegida.value = null
  // Reabre el selector con las direcciones listadas.
  dirAbierta.value = true
  pedirDirecciones('')
}

function cerrarExito() {
  exito.value = false
  if (timerExito) {
    clearTimeout(timerExito)
    timerExito = null
  }
}

// Escape cierra el modal de éxito.
function onKey(e) {
  if (e.key === 'Escape' && exito.value) cerrarExito()
}

onMounted(async () => {
  window.addEventListener('keydown', onKey)
  await loadCatalog().catch(setError)
  cargando.value = false
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKey)
  if (timerExito) clearTimeout(timerExito)
  clearTimeout(dirTimer)
  clearTimeout(dirCloseTimer)
})

async function changePage(page) {
  cargando.value = true
  await loadCatalog(page).catch(setError)
  cargando.value = false
}

// Agrega al carrito la cantidad escrita en la tarjeta del plato.
function agregar(dish) {
  const input = document.getElementById(`${dish.Id}-amount`)
  const amount = Math.max(1, Number(input?.value) || 1)
  addToCart(dish, amount)
  if (input) input.value = 1
}

// Cantidades inválidas (vacías, 0, negativas) vuelven a 1.
function fixAmount(item) {
  item.amount = Math.max(1, Math.floor(Number(item.amount)) || 1)
}

async function submitOrder() {
  error.value = ''
  if (!cartItems.value.length) {
    error.value = 'Agregá al menos un plato.'
    return
  }
  if (!addressId.value) {
    error.value = 'Elegí una dirección de entrega.'
    return
  }
  try {
    await apiFetch('/ordenes', {
      method: 'POST',
      body: {
        detail: detail.value || null,
        address: { id: addressId.value },
        products: cartItems.value.map((i) => ({ id: i.id, amount: i.amount })),
      },
    })
    clearCart()
    detail.value = ''
    addressId.value = ''
    dirElegida.value = null
    // Modal de éxito: se cierra solo a los 4 segundos.
    exito.value = true
    if (timerExito) clearTimeout(timerExito)
    timerExito = setTimeout(cerrarExito, 4000)
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <p v-if="error" class="error">{{ error }}</p>

  <!-- El encabezado va fuera del grid: así la grilla de platos y el panel
       del carrito arrancan a la misma altura. -->
  <h1>Nuestros platos</h1>
  <p class="intro">Elegí tus favoritos, indicá la cantidad y agregalos al pedido.</p>

  <div class="tienda">
    <section class="catalogo">
      <div v-if="cargando">Cargando...</div>

      <template v-else>
        <div class="menu">
          <div v-for="dish in Products" :key="dish.Id" class="card plato">
            <div class="plato-foto">
              <img v-if="dish.Portrait" :src="`${api}/${dish.Portrait}`" :alt="dish.Name">
              <div v-else class="sin-foto">Sin foto</div>
            </div>
            <div class="plato-info">
              <span class="plato-nombre">{{ dish.Name }}</span>
              <span class="plato-precio">${{ (dish.Price / 100).toFixed(2) }}</span>
            </div>
            <div v-if="isAuthenticated()" class="plato-acciones">
              <label :for="`${dish.Id}-amount`">Cantidad</label>
              <input type="number" name="amount" :id="`${dish.Id}-amount`" value="1" min="1">
              <button @click="agregar(dish)">Agregar</button>
            </div>
          </div>
        </div>

        <p class="paginacion">
          <button class="btn-ghost" :disabled="CurrentPage <= 1" @click="changePage(CurrentPage - 1)">Anterior</button>
          Página {{ CurrentPage }} / {{ TotalPages }}
          <button class="btn-ghost" :disabled="CurrentPage >= TotalPages" @click="changePage(CurrentPage + 1)">Siguiente</button>
        </p>
      </template>
    </section>

    <aside class="carrito card">
      <h2>
        Tu pedido
        <small v-if="totalCount">({{ totalCount }} {{ totalCount === 1 ? 'plato' : 'platos' }})</small>
      </h2>

      <template v-if="!isAuthenticated()">
        <p class="vacio">Iniciá sesión para armar tu pedido.</p>
        <router-link to="/login" class="btn">Iniciar sesión</router-link>
      </template>

      <template v-else>
        <p v-if="!cartItems.length" class="vacio">El carrito está vacío: agregá platos del menú.</p>

        <ul v-else class="items">
          <li v-for="item in cartItems" :key="item.id">
            <span class="item-nombre">
              {{ item.name }}
              <small>${{ (item.price / 100).toFixed(2) }} c/u</small>
            </span>
            <input
              class="item-qty"
              type="number"
              min="1"
              v-model.number="item.amount"
              @change="fixAmount(item)"
              :aria-label="`Cantidad de ${item.name}`"
            >
            <span class="item-subtotal">${{ (item.price * item.amount / 100).toFixed(2) }}</span>
            <button
              class="btn-ghost item-quitar"
              @click="removeItem(item.id)"
              :aria-label="`Quitar ${item.name}`"
            >✕</button>
          </li>
        </ul>

        <div class="total">
          <span>Total</span>
          <strong>${{ (totalPrice / 100).toFixed(2) }}</strong>
        </div>

        <template v-if="cartItems.length">
          <div class="campo dir-campo">
            <label for="dir-buscar">Dirección de entrega</label>

            <!-- Dirección ya elegida: chip con la info y botón para cambiarla. -->
            <div v-if="addressId && dirElegida" class="dir-elegida">
              <span class="dir-icono" aria-hidden="true">📍</span>
              <span class="dir-elegida-texto">
                <strong>{{ dirElegida.Street }} {{ dirElegida.House }}</strong>
                <small v-if="dirElegida.Floor !== null || dirElegida.Door">
                  {{ dirElegida.Floor !== null ? `Piso ${dirElegida.Floor}` : '' }}<template v-if="dirElegida.Floor !== null && dirElegida.Door"> · </template><template v-if="dirElegida.Door">Depto {{ dirElegida.Door }}</template>
                </small>
              </span>
              <button type="button" class="btn-ghost dir-cambiar" @click="cambiarDireccion">Cambiar</button>
            </div>

            <!-- Selector con búsqueda: se abre al enfocar y filtra al tipear. -->
            <div v-else class="dir-selector">
              <span class="dir-icono dir-icono-input" aria-hidden="true">📍</span>
              <input
                id="dir-buscar"
                type="text"
                v-model="dirSearch"
                placeholder="Buscar por calle (ej: siempreviva)"
                autocomplete="off"
                @focus="onDirFocus"
                @input="onDirInput"
                @blur="onDirBlur"
              >

              <ul v-if="dirAbierta" class="sugerencias">
                <li v-if="buscandoDir" class="sugerencia-vacia">
                  <span class="dir-spinner" aria-hidden="true"></span> Buscando direcciones...
                </li>
                <template v-else>
                  <li v-for="a in sugerencias" :key="a.Id" class="sugerencia-item">
                    <button type="button" @click="elegirDireccion(a)">
                      <span class="dir-icono" aria-hidden="true">📍</span>
                      <span class="dir-elegida-texto">
                        <strong>{{ a.Street }} {{ a.House }}</strong>
                        <small v-if="a.Floor !== null || a.Door">
                          {{ a.Floor !== null ? `Piso ${a.Floor}` : '' }}<template v-if="a.Floor !== null && a.Door"> · </template><template v-if="a.Door">Depto {{ a.Door }}</template>
                        </small>
                      </span>
                    </button>
                  </li>
                  <li v-if="!sugerencias.length" class="sugerencia-vacia">
                    No tenés direcciones guardadas: <router-link to="/direcciones">creá una</router-link>.
                  </li>
                </template>
              </ul>
            </div>
          </div>

          <label for="detalle">Detalle (opcional):</label>
          <input id="detalle" type="text" v-model="detail" placeholder="Sin cebolla, timbre roto...">

          <button class="confirmar" @click="submitOrder">
            Realizar pedido
          </button>
        </template>
      </template>
    </aside>
  </div>

  <!-- Confirmación del pedido: se cierra solo a los 4 segundos. -->
  <div v-if="exito" class="modal-overlay" @click.self="cerrarExito">
    <div class="modal card" role="dialog" aria-modal="true" aria-label="Pedido realizado">
      <button class="btn-ghost modal-cerrar" aria-label="Cerrar" @click="cerrarExito">✕</button>

      <div class="exito-check" aria-hidden="true">✓</div>
      <h2>¡Pedido realizado!</h2>
      <p class="exito-texto">
        Tu pedido quedó guardado: <router-link to="/historial">seguilo en Historial</router-link>.
      </p>

      <div class="exito-acciones">
        <router-link to="/historial" class="btn">Ver historial</router-link>
        <button class="btn-ghost" @click="cerrarExito">Cerrar</button>
      </div>

      <div class="exito-progreso" aria-hidden="true"><span></span></div>
    </div>
  </div>
</template>

<style scoped>
.tienda {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 1.5rem;
  align-items: start;
}

.intro {
  color: var(--muted);
  margin-top: -0.25rem;
}

/* Controles de "Agregar" en la tarjeta del plato. margin-top: auto los
   ancla al pie de la tarjeta así se alinean entre columnas aunque el
   nombre del plato ocupe dos líneas. */
.plato-acciones {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-top: auto;
  padding: 0 1rem 1rem;
}

.plato-acciones label {
  font-size: 0.85rem;
  color: var(--muted);
}

.plato-acciones input {
  width: 4.5rem;
}

.plato-acciones button {
  flex: 1;
}

/* Panel del carrito. */
.carrito {
  position: sticky;
  top: 4.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.carrito h2 {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.carrito h2 small {
  font-weight: 400;
}

.carrito label {
  font-size: 0.9rem;
  color: var(--muted);
}

.carrito input[type='text'] {
  width: 100%;
}

.vacio {
  color: var(--muted);
  margin: 0;
}

.items {
  list-style: none;
  margin: 0;
  padding: 0;
}

.items li {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
}

.item-nombre {
  flex: 1;
  font-weight: 600;
  font-size: 0.95rem;
}

.item-nombre small {
  display: block;
  font-weight: 400;
}

.item-qty {
  width: 4rem;
  padding: 0.3rem 0.4rem;
}

.item-subtotal {
  min-width: 4.5rem;
  text-align: right;
  font-weight: 600;
}

.item-quitar {
  padding: 0.15rem 0.5rem;
  font-size: 0.85rem;
}

.total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1.05rem;
  padding-top: 0.25rem;
}

/* Dirección elegida: chip con el icono, la info y el botón cambiar. */
.dir-elegida {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 0.7rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg);
}

.dir-elegida-texto {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  line-height: 1.25;
}

.dir-elegida-texto small {
  color: var(--muted);
  font-size: 0.85rem;
}

.dir-cambiar {
  flex: none;
}

/* Selector de dirección con el input y el desplegable de sugerencias. */
.dir-selector {
  position: relative;
}

.dir-icono {
  flex: none;
  font-size: 1rem;
  line-height: 1;
}

.dir-icono-input {
  position: absolute;
  top: 50%;
  left: 0.7rem;
  transform: translateY(-50%);
  pointer-events: none;
  opacity: 0.65;
}

.dir-selector input {
  width: 100%;
  padding-left: 2rem;
}

/* Spinner para el estado "Buscando...". */
.dir-spinner {
  display: inline-block;
  width: 0.9rem;
  height: 0.9rem;
  border: 2px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: dir-gira 0.7s linear infinite;
  vertical-align: -0.15rem;
}

@keyframes dir-gira {
  to {
    transform: rotate(360deg);
  }
}

.sugerencia-item button {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  width: 100%;
}

.confirmar {
  width: 100%;
}

/* Modal de éxito. */
.exito-check {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  margin-bottom: 0.75rem;
  border-radius: 50%;
  background: #d1fae5;
  color: var(--ok);
  font-size: 1.4rem;
  font-weight: 700;
}

.exito-texto {
  color: var(--muted);
}

.exito-acciones {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 1.25rem;
}

/* Barra que marca los 4 segundos antes del cierre. */
.exito-progreso {
  height: 4px;
  margin-top: 1.25rem;
  background: var(--border);
  border-radius: 2px;
  overflow: hidden;
}

.exito-progreso span {
  display: block;
  height: 100%;
  background: var(--ok);
  animation: exito-cuenta 4s linear forwards;
}

@keyframes exito-cuenta {
  from {
    width: 100%;
  }

  to {
    width: 0;
  }
}

@media (max-width: 900px) {
  .tienda {
    grid-template-columns: 1fr;
  }

  .carrito {
    position: static;
  }
}

@media (max-width: 480px) {
  .items li {
    flex-wrap: wrap;
  }

  .item-nombre {
    min-width: 7rem;
  }
}
</style>
