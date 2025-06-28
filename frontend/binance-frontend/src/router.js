import { createRouter, createWebHistory } from 'vue-router';

// 路由组件懒加载
const Dashboard = () => import('./views/Dashboard.vue');
const Order = () => import('./views/Order.vue');
const Strategy = () => import('./views/Strategy.vue');
const Settings = () => import('./views/Settings.vue');
const Login = () => import('./views/Login.vue');
const Register = () => import('./views/Register.vue');
const Admin = () => import('./views/Admin.vue');
const AutoWithdrawal = () => import('./views/AutoWithdrawal.vue');
const DualInvestment = () => import('./views/DualInvestment.vue');

// 路由配置
const routes = [
    // 需要认证的路由
    {
        path: '/',
        name: 'Dashboard',
        component: Dashboard,
        meta: {
            requiresAuth: true,
            title: '仪表盘'
        }
    },
    {
        path: '/orders',
        name: 'Orders',
        component: Order,
        meta: {
            requiresAuth: true,
            title: '订单管理'
        }
    },
    {
        path: '/strategies',
        name: 'Strategies',
        component: Strategy,
        meta: {
            requiresAuth: true,
            title: '策略管理'
        }
    },
    {
        path: '/settings',
        name: 'Settings',
        component: Settings,
        meta: {
            requiresAuth: true,
            title: '系统设置'
        }
    },
    {
        path: '/dual-investment',
        name: 'DualInvestment',
        component: DualInvestment,
        meta: {
            requiresAuth: true,
            title: '双币投资'
        }
    },
    {
        path: '/auto-withdrawal',
        name: 'AutoWithdrawal',
        component: AutoWithdrawal,
        meta: {
            requiresAuth: true,
            title: '自动提币'
        }
    },

    // 管理员路由
    {
        path: '/admin',
        name: 'Admin',
        component: Admin,
        meta: {
            requiresAuth: true,
            requiresAdmin: true,
            title: '用户管理'
        }
    },

    // 公开路由
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: {
            title: '登录'
        }
    },
    {
        path: '/register',
        name: 'Register',
        component: Register,
        meta: {
            title: '注册'
        }
    },

    // 404 页面
    {
        path: '/:pathMatch(.*)*',
        name: 'NotFound',
        redirect: '/'
    }
];

// 创建路由实例
const router = createRouter({
    history: createWebHistory(),
    routes,
    // 滚动行为
    scrollBehavior(to, from, savedPosition) {
        if (savedPosition) {
            return savedPosition;
        } else {
            return { top: 0 };
        }
    }
});

// 导航守卫
router.beforeEach((to, from, next) => {
    // 设置页面标题
    document.title = to.meta.title ? `${to.meta.title} - 交易系统` : '交易系统';

    const token = localStorage.getItem('token');

    // 验证 token 有效性
    const isValidToken = token &&
        token !== 'undefined' &&
        token !== 'null' &&
        token !== '' &&
        token.split('.').length === 3;

    const requiresAuth = to.meta.requiresAuth;
    const requiresAdmin = to.meta.requiresAdmin;

    // 需要认证但 token 无效
    if (requiresAuth && !isValidToken) {
        // 清理无效 token
        if (token && !isValidToken) {
            console.log('清理无效的 token');
            localStorage.removeItem('token');
        }

        next({
            path: '/login',
            query: { redirect: to.fullPath } // 保存原始访问路径
        });
        return;
    }

    // 已登录用户访问登录/注册页面
    if (isValidToken && (to.path === '/login' || to.path === '/register')) {
        next('/');
        return;
    }

    // 需要管理员权限
    if (requiresAdmin && isValidToken) {
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            if (payload.role !== 'admin') {
                console.warn('🚫 无管理员权限');
                next('/');
                return;
            }
        } catch (e) {
            console.error('Token 解析失败:', e);
            localStorage.removeItem('token');
            next('/login');
            return;
        }
    }

    next();
});

// 导航后置守卫
router.afterEach((to, from) => {
    // 可以在这里添加页面访问统计等功能
    if (process.env.NODE_ENV === 'development') {
        console.log(`📍 导航到: ${to.path}`);
    }
});

// 路由错误处理
router.onError((error) => {
    console.error('路由错误:', error);
});

export default router;