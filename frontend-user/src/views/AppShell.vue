<template>
  <div class="shell">
    <aside class="side hide-tablet" v-if="showSide">
      <div class="brand" @click="$router.push('/')">
        <span class="serif">GoWiki</span>
        <small>知识库</small>
      </div>
      <select v-model="spaceId" @change="loadTree">
        <option v-for="s in spaces" :key="s.id" :value="s.id">{{ s.name }}</option>
      </select>
      <div class="side-actions">
        <button class="btn btn-primary" @click="createRoot">新文档</button>
        <button class="btn btn-ghost" @click="openSearch">检索</button>
      </div>
      <DocTree :nodes="tree" :current="currentId" @open="open" @move="onMove" @create="createChild" @remove="onRemove" />
    </aside>
    <main class="main">
      <header class="top">
        <button class="btn btn-ghost" @click="$router.push('/')">工作台</button>
        <input v-model="q" class="search" placeholder="搜索文档内容…" @keydown.enter="doSearch" />
        <button class="btn btn-ghost" @click="logout">退出</button>
      </header>
      <router-view />
    </main>
    <SearchPalette v-if="palette" :hits="hits" @close="palette=false" @open="open" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, type Document, type Space } from '../api/wiki'
import { useSession } from '../stores/session'
import DocTree from '../components/DocTree.vue'
import SearchPalette from '../components/SearchPalette.vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const session = useSession()
const spaces = ref<Space[]>([])
const spaceId = ref('')
const tree = ref<Document[]>([])
const q = ref('')
const palette = ref(false)
const hits = ref<any[]>([])
const showSide = ref(true)

const currentId = ref(String(route.params.id || ''))

async function loadSpaces() {
  const res = await api.spaces()
  spaces.value = res.data || []
  if (!spaceId.value && spaces.value[0]) spaceId.value = spaces.value[0].id
  await loadTree()
}
async function loadTree() {
  if (!spaceId.value) return
  const res = await api.tree(spaceId.value)
  tree.value = res.data || []
}
function open(id: string) {
  currentId.value = id
  palette.value = false
  router.push(`/doc/${id}`)
}
async function createRoot() {
  const res = await api.createDoc({ spaceId: spaceId.value, title: '未命名文档' })
  await loadTree()
  open(res.data.id)
}
async function createChild(parentId: string) {
  const res = await api.createDoc({ spaceId: spaceId.value, parentId, title: '子文档' })
  await loadTree()
  open(res.data.id)
}
async function onMove(payload: { id: string; parentId: string | null; sortOrder: number }) {
  try {
    await api.moveDoc(payload.id, payload.parentId, payload.sortOrder)
    await loadTree()
  } catch (e: any) {
    const code = e?.response?.data?.code
    if (code === 'TREE_CYCLE') ElMessage.error('不能将节点拖入自身子树')
  }
}
async function onRemove(id: string) {
  await api.deleteDoc(id)
  await loadTree()
  if (currentId.value === id) router.push('/')
}
async function doSearch() {
  if (!q.value.trim()) return
  const res = await api.search(q.value.trim())
  hits.value = res.data.hits || []
  palette.value = true
}
function openSearch() {
  palette.value = true
}
function logout() {
  session.logout()
  router.push('/login')
}

onMounted(loadSpaces)
</script>

<style scoped>
.shell { display:flex; min-height:100vh; width:100%; }
.side { width:280px; background:var(--sidebar); border-right:1px solid var(--line); padding:18px 14px; }
.brand { cursor:pointer; margin-bottom:14px; }
.brand span { font-size:28px; display:block; }
.brand small { color:var(--muted); letter-spacing:0.16em; }
.side-actions { display:flex; gap:8px; margin:12px 0; }
.main { flex:1; min-width:0; }
.top { display:flex; gap:10px; align-items:center; padding:14px 18px; border-bottom:1px solid var(--line); }
.search { flex:1; border:1px solid var(--line); border-radius:999px; padding:8px 14px; background:var(--card); }
select { width:100%; border:1px solid var(--line); border-radius:12px; padding:8px 12px; background:var(--card); }
</style>
