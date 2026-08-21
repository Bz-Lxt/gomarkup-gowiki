import { createRouter, createWebHashHistory } from 'vue-router'
import LoginView from './views/LoginView.vue'
import AppShell from './views/AppShell.vue'
import WorkbenchView from './views/WorkbenchView.vue'
import EditorView from './views/EditorView.vue'

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login', component: LoginView },
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', component: WorkbenchView },
        { path: 'doc/:id', component: EditorView },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('gowiki.token')
  if (!token && to.path !== '/login') return '/login'
  if (token && to.path === '/login') return '/'
})
