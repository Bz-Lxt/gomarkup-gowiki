<template>
  <div class="page-shell" style="display:grid;place-items:center;padding:32px">
    <div class="card" style="width:min(420px,100%);padding:36px">
      <p class="serif" style="margin:0;color:var(--terracotta);letter-spacing:0.12em;font-size:12px">GOWIKI</p>
      <h1 class="serif" style="margin:8px 0 6px;font-size:36px">把知识摊在桌上</h1>
      <p style="color:var(--muted);margin:0 0 24px">团队协作 Wiki · 纸稿编辑室</p>
      <form @submit.prevent="submit">
        <label>邮箱 *</label>
        <input v-model="email" type="email" placeholder="admin@gowiki.dev" />
        <p v-if="errors.email" class="err">{{ errors.email }}</p>
        <label>密码 *</label>
        <input v-model="password" type="password" placeholder="至少 6 位" />
        <p v-if="errors.password" class="err">{{ errors.password }}</p>
        <label v-if="mode==='register'">昵称 *</label>
        <input v-if="mode==='register'" v-model="name" placeholder="怎么称呼你" />
        <div style="display:flex;gap:10px;margin-top:18px">
          <button class="btn btn-primary" type="submit">{{ mode==='login' ? '进入知识库' : '创建账号' }}</button>
          <button class="btn btn-ghost" type="button" @click="mode = mode==='login'?'register':'login'">
            {{ mode==='login' ? '没有账号' : '已有账号' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '../stores/session'

const session = useSession()
const router = useRouter()
const mode = ref<'login' | 'register'>('login')
const email = ref('admin@gowiki.dev')
const password = ref('admin123')
const name = ref('林昭')
const errors = reactive({ email: '', password: '' })

async function submit() {
  errors.email = email.value.includes('@') ? '' : '邮箱格式不正确'
  errors.password = password.value.length >= 6 ? '' : '密码至少 6 位'
  if (errors.email || errors.password) return
  if (mode.value === 'login') await session.login(email.value, password.value)
  else await session.register(email.value, password.value, name.value)
  router.push('/')
}
</script>

<style scoped>
label { display:block; margin:12px 0 6px; color:var(--muted); font-size:13px }
input {
  width:100%; border:1px solid var(--line); background:#fff; border-radius:12px;
  padding:10px 12px; font-size:14px;
}
.err { color:var(--danger); font-size:12px; margin:4px 0 0 }
</style>
