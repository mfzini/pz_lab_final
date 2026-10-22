<script setup>
import { useRouter } from 'vue-router'
import { api } from '@/common'
import { session, isAdmin, clearSession } from '@/auth'

const router = useRouter()

function logout() {
  clearSession()
  router.push('/')
}
</script>

<template>
  <header>
    <router-link to="/" class="logo">PZ Lab</router-link>
    <nav>
      <router-link to="/">Inicio</router-link>
      <template v-if="!session.token">
        <router-link to="/login">Iniciar sesión</router-link>
        <router-link to="/persona">Registrarse</router-link>
      </template>
      <template v-else>
        <router-link to="/historial">Historial</router-link>
        <router-link to="/direcciones">Direcciones</router-link>
        <router-link to="/perfil">Perfil</router-link>
        <template v-if="isAdmin">
          <router-link to="/admin/platos">Platos</router-link>
          <router-link to="/admin/pedidos">Gestionar pedidos</router-link>
          <router-link to="/admin/usuarios">Usuarios</router-link>
        </template>
        <router-link v-if="session.user?.Avatar" to="/perfil" class="avatar-link" aria-label="Mi perfil">
          <img class="avatar-nav" :src="`${api}/${session.user.Avatar}`" alt="">
        </router-link>
        <span class="usuario">{{ session.user?.Name }}</span>
        <button class="btn-ghost" @click="logout">Salir</button>
      </template>
    </nav>
  </header>

  <main>
    <router-view></router-view>
  </main>

  <footer class="pie">PZ Lab — pedidos online</footer>
</template>

<style scoped>
header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1.25rem;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
}

.logo {
  font-weight: 800;
  font-size: 1.1rem;
  color: var(--text);
}

.logo:hover {
  color: var(--primary);
  text-decoration: none;
}

nav {
  display: flex;
  gap: 0.35rem;
  align-items: center;
  flex-wrap: wrap;
}

nav a {
  color: var(--text);
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
}

nav a:hover {
  background: #efece6;
  text-decoration: none;
}

/* La ruta actual, en pastilla del color primario. */
nav a.router-link-exact-active {
  background: var(--primary);
  color: #fff;
}

nav .usuario {
  margin-left: 0.5rem;
  color: var(--muted);
  font-size: 0.9rem;
}

/* Avatar en el header: chiquito, redondo, sin romper el layout. */
.avatar-link {
  display: flex;
  padding: 0;
}

.avatar-nav {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
}

/* Chicas: el header se apila (logo arriba, links abajo) y no queda apretado. */
@media (max-width: 640px) {
  header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.6rem 0.9rem;
  }

  nav .usuario {
    margin-left: 0;
  }
}
</style>
