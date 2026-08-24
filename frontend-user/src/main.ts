import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import Login from './views/Login.vue'
import Studio from './views/Studio.vue'
import './style.css'
import 'leaflet/dist/leaflet.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: Login },
    { path: '/', component: Studio },
    { path: '/privacy', component: Studio },
    { path: '/terms', component: Studio },
  ],
})

createApp(App).use(createPinia()).use(router).mount('#app')
