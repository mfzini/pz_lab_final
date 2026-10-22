<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login, isAdmin } from '@/auth'

const route = useRoute()
const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')

async function sendRequest() {
  error.value = ''
  try {
    await login(email.value, password.value)
    // Volvemos a la ruta que pidió el guard, o al home de cada rol.
    const next = route.query.next || (isAdmin.value ? '/admin/pedidos' : '/')
    router.push(next)
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <form id="loginForm" class="card form-column" @submit.prevent="sendRequest">
    <h1>Iniciar sesión</h1>
    <label for="email">Correo:</label>
    <input type="email" name="email" id="email" v-model="email" required>
    <label for="password">Contraseña:</label>
    <input type="password" name="password" id="password" v-model="password" required>
    <button type="submit">Entrar</button>
    <p v-if="error" class="error">{{ error }}</p>
  </form>
  <p class="registro">
    ¿No tenés cuenta? <router-link to="/persona">Registrate</router-link>
  </p>
</template>

<style scoped>
#loginForm {
  margin-top: 2.5rem;
}

.registro {
  text-align: center;
  margin-top: 1rem;
  color: var(--muted);
}
</style>
