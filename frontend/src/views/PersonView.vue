<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '@/auth'
import { apiFetch } from '@/api'

const router = useRouter()
const user = ref({
  name: "",
  surname: "",
  email: "",
  password: "",
  phone: "",
})
// La dirección es parte del registro: sin ella no se puede pedir.
// Campos anulables vacíos mandan null (piso, depto, detalle).
const direccion = ref({
  street: "",
  house: "",
  floor: "",
  door: "",
  detail: "",
})
const error = ref("")

function direccionPayload() {
  return {
    street: direccion.value.street,
    house: direccion.value.house,
    floor: direccion.value.floor === "" ? null : Number(direccion.value.floor),
    door: direccion.value.door || null,
    detail: direccion.value.detail || null,
  }
}

async function sendRequest() {
  error.value = ""
  try {
    await register({
      ...user.value,
      // Sin teléfono la columna queda en null.
      phone: user.value.phone || null,
    })
  } catch (e) {
    error.value = e.message
    return
  }

  // Ya tenemos sesión: la dirección se crea con el usuario del token
  // (el backend no acepta user_id en el body).
  try {
    await apiFetch("/direcciones", { method: "POST", body: direccionPayload() })
  } catch {
    // La cuenta ya existe: no la frenamos por esto. Si la dirección falló,
    // se carga desde su propia vista.
    router.push("/direcciones")
    return
  }
  router.push("/")
}
</script>

<template>
  <form id="personForm" class="card form-column" @submit.prevent="sendRequest">
    <h1>Crear cuenta</h1>
    <label for="name">Nombre:</label>
    <input type="text" name="name" id="name" v-model="user.name" required>
    <label for="surname">Apellido:</label>
    <input type="text" name="surname" id="surname" v-model="user.surname" required>
    <label for="email">Correo:</label>
    <input type="email" name="email" id="email" v-model="user.email" required>
    <label for="password">Contraseña:</label>
    <input type="password" name="password" id="password" v-model="user.password" required>
    <label for="phone">Teléfono (opcional):</label>
    <input type="tel" name="phone" id="phone" v-model="user.phone">

    <h2 class="seccion">Dirección de entrega</h2>
    <label for="street">Calle:</label>
    <input type="text" id="street" v-model="direccion.street" required>
    <label for="house">Altura:</label>
    <input type="text" id="house" v-model="direccion.house" required>
    <label for="floor">Piso (opcional):</label>
    <input type="number" id="floor" v-model="direccion.floor">
    <label for="door">Depto (opcional):</label>
    <input type="text" id="door" v-model="direccion.door">
    <label for="detail">Detalle (opcional):</label>
    <input type="text" id="detail" v-model="direccion.detail" placeholder="Timbre roto...">

    <button type="submit">Crear cuenta</button>
    <p v-if="error" class="error">{{ error }}</p>
  </form>
  <p class="login">
    ¿Ya tenés cuenta? <router-link to="/login">Iniciá sesión</router-link>
  </p>
</template>

<style scoped>
#personForm {
  margin: 2.5rem auto 0;
}

/* Separador visual entre la cuenta y la dirección. */
.seccion {
  width: 100%;
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border);
}

.login {
  text-align: center;
  margin-top: 1rem;
  color: var(--muted);
}
</style>
