import { createRouter, createWebHistory } from 'vue-router'

import Home from '@/views/HomeView.vue'
import About from '@/views/AboutPage.vue'
import Login from '@/views/LoginView.vue'
import Persona from '@/views/PersonView.vue'
import HistorialView from '@/views/HistorialView.vue'
import AddressesView from '@/views/AddressesView.vue'
import ProfileView from '@/views/ProfileView.vue'
import AdminOrdersView from '@/views/AdminOrdersView.vue'
import AdminProductsView from '@/views/AdminProductsView.vue'
import AdminUsersView from '@/views/AdminUsersView.vue'
import { isAuthenticated, isAdmin } from '@/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Home', component: Home },
    { path: '/about', name: 'About', component: About },
    { path: '/login', name: 'Login', component: Login, meta: { guestOnly: true } },
    { path: '/persona', name: 'Persona', component: Persona, meta: { guestOnly: true } },
    // Platos: alta y edición en modal: sólo admin.
    { path: '/admin/platos', name: 'AdminProducts', component: AdminProductsView, meta: { requiresAuth: true, requiresAdmin: true } },
    // La vista vieja de alta quedó unificada acá.
    { path: '/platos', redirect: '/admin/platos' },
    // El pedido se arma con el carrito del índice.
    { path: '/pedidos', redirect: '/' },
    // Historial propio de pedidos.
    { path: '/historial', name: 'Historial', component: HistorialView, meta: { requiresAuth: true } },
    { path: '/direcciones', name: 'Addresses', component: AddressesView, meta: { requiresAuth: true } },
    // Perfil: datos propios + avatar (subida de archivo).
    { path: '/perfil', name: 'Profile', component: ProfileView, meta: { requiresAuth: true } },
    // Todas las órdenes con filtros de fecha y estado: sólo admin.
    { path: '/admin/pedidos', name: 'AdminOrders', component: AdminOrdersView, meta: { requiresAuth: true, requiresAdmin: true } },
    // Usuarios con su rol: el admin puede ascender o degradar (no a sí mismo).
    { path: '/admin/usuarios', name: 'AdminUsers', component: AdminUsersView, meta: { requiresAuth: true, requiresAdmin: true } },
  ],
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    // guardamos a dónde iba para volver después del login.
    return { name: 'Login', query: { next: to.fullPath } }
  }
  if (to.meta.requiresAdmin && !isAdmin.value) {
    return { name: 'Home' }
  }
  if (to.meta.guestOnly && isAuthenticated()) {
    return { name: 'Home' }
  }
})

export default router
