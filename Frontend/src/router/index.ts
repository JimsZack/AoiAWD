import { createRouter, createWebHistory } from 'vue-router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

NProgress.configure({ showSpinner: false })

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { hidden: true }
    },
    {
      path: '/',
      component: () => import('@/views/Home.vue'),
      redirect: '/main',
      children: [
        {
          path: 'main',
          name: 'Main',
          component: () => import('@/views/Main.vue'),
          meta: { title: '首页', icon: 'HomeFilled' }
        },
        {
          path: 'webLog',
          name: 'WebLog',
          component: () => import('@/views/logSystem/webLog.vue'),
          meta: { title: 'Web日志', icon: 'Monitor' }
        },
        {
          path: 'pwnLog',
          name: 'PwnLog',
          component: () => import('@/views/logSystem/pwnLog.vue'),
          meta: { title: 'PWN日志', icon: 'Warning' }
        },
        {
          path: 'fileLog',
          name: 'FileLog',
          component: () => import('@/views/logSystem/fileLog.vue'),
          meta: { title: '文件日志', icon: 'Document' }
        },
        {
          path: 'processLog',
          name: 'ProcessLog',
          component: () => import('@/views/logSystem/processLog.vue'),
          meta: { title: '进程日志', icon: 'Cpu' }
        },
        {
          path: 'warnLog',
          name: 'WarnLog',
          component: () => import('@/views/logSystem/warnLog.vue'),
          meta: { title: '系统警告', icon: 'Bell' }
        }
      ]
    },
    {
      path: '/webLog/detail/:id',
      name: 'WebDetail',
      component: () => import('@/views/detailPages/webLog.vue'),
      meta: { hidden: true }
    },
    {
      path: '/pwnLog/detail/:id',
      name: 'PwnDetail',
      component: () => import('@/views/detailPages/pwnLog.vue'),
      meta: { hidden: true }
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/main'
    }
  ]
})

router.beforeEach((to, _from, next) => {
  NProgress.start()
  const token = sessionStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

router.afterEach(() => {
  NProgress.done()
})

export default router
