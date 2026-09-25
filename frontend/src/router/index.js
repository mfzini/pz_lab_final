import { createRouter, createWebHistory } from 'vue-router'

import Home from '@/views/HomeView.vue'
import About from '@/views/About.vue'
import Persona from '@/views/PersonView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/about',
      name: 'About',
      component: About
    },
    {
      path: '/persona',
      name: 'Persona',
      component: Persona
    }
  ]
});

export default router;