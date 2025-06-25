import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from './views/Dashboard.vue';
import Order from './views/Order.vue';
import Strategy from './views/Strategy.vue';
import Settings from './views/Settings.vue';
import Login from './views/Login.vue';

const routes = [
    { path: '/', component: Dashboard, meta: { requiresAuth: true } },
    { path: '/orders', component: Order, meta: { requiresAuth: true } },
    { path: '/strategies', component: Strategy, meta: { requiresAuth: true } },
    { path: '/settings', component: Settings, meta: { requiresAuth: true } },
    { path: '/login', component: Login },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token');
    if (to.meta.requiresAuth && !token) {
        next('/login');
    } else {
        next();
    }
});

export default router;