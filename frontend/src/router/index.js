import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/login', name: 'Login', component: () => import('../views/Login.vue') },
  { path: '/register', name: 'Register', component: () => import('../views/Register.vue') },
  { path: '/', name: 'Feed', component: () => import('../views/Feed.vue') },
  { path: '/search', name: 'Search', component: () => import('../views/Search.vue') },
  { path: '/upload', name: 'Upload', component: () => import('../views/Upload.vue'), meta: { auth: true } },
  { path: '/ranking', name: 'Ranking', component: () => import('../views/Ranking.vue') },
  { path: '/chat', name: 'Chat', component: () => import('../views/Chat.vue'), meta: { auth: true } },
  { path: '/chat/:userId', name: 'ChatConversation', component: () => import('../views/ChatConversation.vue'), meta: { auth: true } },
  { path: '/profile', name: 'Profile', component: () => import('../views/Profile.vue'), meta: { auth: true } },
  { path: '/user/:id', name: 'UserProfile', component: () => import('../views/UserProfile.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫：需要登录的页面自动跳转
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.auth && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
