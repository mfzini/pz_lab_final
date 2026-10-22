import { reactive, computed } from 'vue'

// Carrito en memoria: se pierde al recargar la página.
// Items: { id, name, price (centavos del plato), amount }
const state = reactive({ items: [] })

export const cartItems = computed(() => state.items)

export const totalCount = computed(() =>
  state.items.reduce((n, i) => n + i.amount, 0),
)

// Total en centavos.
export const totalPrice = computed(() =>
  state.items.reduce((n, i) => n + i.price * i.amount, 0),
)

export function addToCart(product, amount = 1) {
  const found = state.items.find((i) => i.id === product.Id)
  if (found) {
    found.amount += amount
    return
  }
  state.items.push({
    id: product.Id,
    name: product.Name,
    price: product.Price,
    amount,
  })
}

export function removeItem(id) {
  const index = state.items.findIndex((i) => i.id === id)
  if (index >= 0) state.items.splice(index, 1)
}

export function clearCart() {
  state.items = []
}
