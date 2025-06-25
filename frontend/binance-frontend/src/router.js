import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from './views/Dashboard.vue';
import Order from './views/Order.vue';
import Strategy from './views/Strategy.vue';
import Settings from './views/Settings.vue';
import Login from './views/Login.vue';
import Register from './views/Register.vue';
import Admin from './views/Admin.vue';
import AuthWithdrawal from './views/AutoWithdrawal.vue'

const routes = [
    { path: '/', component: Dashboard, meta: { requiresAuth: true } },
    { path: '/orders', component: Order, meta: { requiresAuth: true } },
    { path: '/strategies', component: Strategy, meta: { requiresAuth: true } },
    { path: '/settings', component: Settings, meta: { requiresAuth: true } },
    { path: '/admin', component: Admin, meta: { requiresAuth: true, requiresAdmin: true } },
    { path: '/login', component: Login },
    { path: '/register', component: Register },
    {
        path: '/authwithdrawal',
        name: 'AuthWithdrawal',
        component: AuthWithdrawal
    },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token');

    if (to.meta.requiresAuth && !token) {
        next('/login');
    } else if (to.meta.requiresAdmin) {
        // 检查是否为管理员
        if (token) {
            try {
                const payload = JSON.parse(atob(token.split('.')[1]));
                if (payload.role === 'admin') {
                    next();
                } else {
                    next('/'); // 非管理员重定向到首页
                }
            } catch (e) {
                next('/login');
            }
        } else {
            next('/login');
        }
    } else {
        next();
    }
});

export default router;