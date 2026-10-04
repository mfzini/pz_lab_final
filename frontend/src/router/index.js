import { createRouter, createWebHistory } from 'vue-router'

import Home from '@/views/HomeView.vue'
import About from '@/views/About.vue'
import Persona from '@/views/PersonView.vue'
import ProductsView from '@/views/ProductsView.vue'
import OrderView from '@/views/OrderView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {path: '/',name: 'Home',component: Home},
    {path: '/about',name: 'About',component: About},
    {path: '/persona',name: 'Persona',component: Persona},
    {path: '/platos', name: 'Dishes', component: ProductsView},
    {path: '/pedidos', name: 'Orders', component: OrderView}
  ]
});

export default router;