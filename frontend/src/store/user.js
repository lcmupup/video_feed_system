import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const userId = computed(() => user.value?.user_id || user.value?.id || 0)
  const username = computed(() => user.value?.username || '')
  const avatar = computed(() => user.value?.avatar || '')
  const nickname = computed(() => user.value?.nickname || user.value?.username || '')

  function setAuth(t, u) {
    token.value = t
    user.value = u
    localStorage.setItem('token', t)
    localStorage.setItem('user', JSON.stringify(u))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function updateUser(u) {
    user.value = { ...user.value, ...u }
    localStorage.setItem('user', JSON.stringify(user.value))
  }

  return { token, user, isLoggedIn, userId, username, avatar, nickname, setAuth, logout, updateUser }
})
