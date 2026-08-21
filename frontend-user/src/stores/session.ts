import { defineStore } from 'pinia'
import { api, type User } from '../api/wiki'

export const useSession = defineStore('session', {
  state: () => ({
    token: localStorage.getItem('gowiki.token') || '',
    user: null as User | null,
  }),
  actions: {
    async login(email: string, password: string) {
      const res = await api.login(email, password)
      this.token = res.data.accessToken
      this.user = res.data.user
      localStorage.setItem('gowiki.token', this.token)
      localStorage.setItem('gowiki.refresh', res.data.refreshToken)
    },
    async register(email: string, password: string, name: string) {
      const res = await api.register(email, password, name)
      this.token = res.data.accessToken
      this.user = res.data.user
      localStorage.setItem('gowiki.token', this.token)
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('gowiki.token')
    },
  },
})
