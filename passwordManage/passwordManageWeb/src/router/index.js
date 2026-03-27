import { createRouter, createWebHistory } from 'vue-router'
import Login from '../components/Login.vue';
import Register from '../components/Register.vue';
import Manager from '../components/Manager.vue';
import PasswordManage from '../views/passwordManage.vue'
import About from '../views/about.vue'
import Account from '../views/account.vue'
import Dashboard from '../views/dashboard.vue'
import ShareManage from '../views/shareManage.vue'
import Setting from '../views/set.vue'
import Introduce from '../views/introduce.vue'
import Tools from '../views/tools.vue'
import LedgerManage from '../views/ledgerManage.vue'
import { useAuthStore } from '../store/auth';
import { ElMessageBox } from 'element-plus';


const routes = [
  {
    path: '/',
    name: 'Login',
    component: Login,
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
  },
  {
    path: '/manager',
    name: 'Manager',
    component: Manager,
    children: [
      {
        path: '/passwordManage',
        name: 'PasswordManage',
        meta: {
          requiresAuth: true,
          permissions: ["user.isLogin"]
        },
        component: PasswordManage,
      },
      {
        path: '/about',
        name: 'About',
        component: About,
      },
      {
        path: '/account',
        name: 'Account',
        meta: {
          requiresAuth: true,
          permissions: ['user.isLogin']
        },
        component: Account,
      },
      {
        path: '/dashboard',
        name: 'Dashboard',
        meta: {
          requiresAuth: true,
          permissions: ['user.isLogin']
        },
        component: Dashboard,
      },
      {
        path: '/shareManage',
        name: 'ShareManage',
        meta: {
          requiresAuth: true,
          permissions: ['user.isLogin']
        },
        component: ShareManage,
      },
      {
        path: '/ledgerManage',
        name: 'LedgerManage',
        meta: {
          requiresAuth: true,
          permissions: ['user.isLogin']
        },
        component: LedgerManage,
      },
      {
        path: '/set',
        name: 'Setting',
        meta: {
          requiresAuth: true,
          permissions: ["user.isLogin"]
        },
        component: Setting,
      },
      {
        path: '/introduce',
        name: 'Introduce',
        component: Introduce,
      },
      {
        path: '/tools',
        name: 'Tools',
        component: Tools,
      },
      {
        path: '/:pathMatch(.*)*',
        redirect: '/404'
      }
    ]
  }

];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  // 需要认证页面
  if (to.meta.requiresAuth) {
    // 检查会话状态
    await authStore.checkSession();
    if (!authStore.username) {
      // 未登录，重定向到登录页
      return '/'
    } else {
      const requiredPermissions = to.meta.permissions;
      // 已登录，不需要权限验证的路由直接放行
      if (requiredPermissions === undefined) {
        return true;
      }
      if (!Array.isArray(requiredPermissions)) {
        ElMessageBox.error("权限配置错误");
        return '/tools'; // 跳转到错误页面
      }
      // 检查权限
      const hasPermission = requiredPermissions.every(perm =>
        authStore.hasPermission(perm)
      )
      if (!hasPermission) {
        return '/tools'
      }
      return true
    }
  }
  return true

})

export default router;
