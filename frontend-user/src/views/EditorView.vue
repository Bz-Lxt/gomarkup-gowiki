<template>
  <div class="editor-page" v-if="doc">
    <div class="canvas">
      <div class="meta">
        <input class="title" v-model="doc.title" @change="saveTitle" />
        <div class="tools">
          <select v-model="mode" @change="switchMode">
            <option value="markdown">Markdown 协同</option>
            <option value="rich">富文本 · 段落锁</option>
          </select>
          <button class="btn btn-ghost" @click="toggleFav">{{ favorite ? '已收藏' : '收藏' }}</button>
          <button class="btn btn-primary" @click="askSave = true">保存版本</button>
        </div>
        <div class="faces">
          <span v-for="u in users" :key="u.userId" class="face" :style="{ background: u.color }">{{ u.name?.[0] }}</span>
        </div>
      </div>

      <textarea
        v-if="mode==='markdown'"
        ref="ta"
        class="md"
        :value="text"
        @input="onMdInput"
      />

      <div v-else class="rich">
        <editor-content v-if="editor" :editor="editor" />
      </div>
    </div>

    <aside class="rail">
      <h3 class="serif">版本</h3>
      <button v-for="v in versions" :key="v.id" class="ver" @click="showDiff(v.id)">
        <b>{{ v.label || v.layer }}</b>
        <small>{{ formatTime(v.createdAt) }}</small>
      </button>
      <div v-if="diff" class="diff-box">
        <DiffView :line="diff.line || []" :chars="diff.char || []" />
        <button class="btn btn-primary" @click="doRollback">回滚到此版本</button>
      </div>
    </aside>

    <div v-if="askSave" class="modal-mask" @click.self="askSave=false">
      <div class="modal">
        <h3 class="serif">保存命名版本</h3>
        <input v-model="label" placeholder="例如：发布前终稿" />
        <p v-if="labelErr" style="color:var(--danger)">{{ labelErr }}</p>
        <div style="display:flex;gap:8px;justify-content:flex-end;margin-top:12px">
          <button class="btn btn-ghost" @click="askSave=false">取消</button>
          <button class="btn btn-primary" @click="saveVer">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Editor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import fastDiff from 'fast-diff'
import { ElMessage } from 'element-plus'
import { api, type Document, type Version } from '../api/wiki'
import { connectCollab } from '../lib/collab'
import { LocalRGA } from '../lib/rga-client'
import DiffView from '../components/DiffView.vue'

const lowlight = createLowlight(common)
const route = useRoute()
const doc = ref<Document | null>(null)
const favorite = ref(false)
const mode = ref<'markdown' | 'rich'>('markdown')
const text = ref('')
const users = ref<any[]>([])
const versions = ref<Version[]>([])
const diff = ref<any>(null)
const askSave = ref(false)
const label = ref('')
const labelErr = ref('')
const ta = ref<HTMLTextAreaElement | null>(null)
const applying = ref(false)
const rga = new LocalRGA()
let sock: ReturnType<typeof connectCollab> | null = null
const editor = ref<Editor | null>(null)
let diffTarget = ''

function formatTime(s: string) {
  if (!s) return ''
  return s.replace('T', ' ').slice(0, 19)
}

async function load() {
  const id = String(route.params.id)
  const res = await api.getDoc(id)
  doc.value = res.data.document
  favorite.value = res.data.favorite
  mode.value = (doc.value.editorMode as any) || 'markdown'
  text.value = doc.value.contentMd || ''
  const vs = await api.versions(id)
  versions.value = vs.data || []
  connect(id)
  setupEditor()
}

function setupEditor() {
  editor.value?.destroy()
  editor.value = new Editor({
    extensions: [
      StarterKit.configure({ codeBlock: false }),
      Image,
      Placeholder.configure({ placeholder: '从这里写下一页纸稿…' }),
      CodeBlockLowlight.configure({ lowlight }),
    ],
    content: tryJSON(doc.value?.contentJson) || doc.value?.contentMd || '',
    editorProps: { attributes: { class: 'tiptap' } },
    onUpdate: async ({ editor: ed }) => {
      if (!doc.value || mode.value !== 'rich') return
      await api.updateDoc(doc.value.id, { contentMd: ed.getText(), contentJson: JSON.stringify(ed.getJSON()) })
    },
  })
}

function tryJSON(raw?: string) {
  if (!raw) return null
  try { return JSON.parse(raw) } catch { return null }
}

function connect(id: string) {
  sock?.close()
  const token = localStorage.getItem('gowiki.token') || ''
  sock = connectCollab(id, token, {
    onSnapshot() {},
    onOp() {},
    onPresence() {},
    onLock() {},
    onError() {},
  })
  sock.raw.onmessage = (ev) => {
    const msg = JSON.parse(ev.data)
    if (msg.type === 'snapshot') {
      rga.load(msg.siteId, msg.clock || 0, msg.atoms || [])
      if (!rga.text() && msg.text) {
        // fallback empty snapshot with seed text already on server
      }
      text.value = rga.text() || msg.text || ''
      users.value = msg.users || []
    } else if (msg.type === 'op' && msg.op) {
      applying.value = true
      rga.apply(msg.op)
      text.value = rga.text()
      nextTick(() => { applying.value = false })
    } else if (msg.type === 'presence') {
      users.value = msg.users || []
    } else if (msg.type === 'lock' && msg.message) {
      ElMessage.warning(msg.message)
    } else if (msg.type === 'error') {
      ElMessage.error(msg.message)
    }
  }
}

function onMdInput(e: Event) {
  if (applying.value) return
  const next = (e.target as HTMLTextAreaElement).value
  const changes = fastDiff(text.value, next)
  let idx = 0
  const ops = []
  for (const [kind, chunk] of changes) {
    if (kind === 0) idx += Array.from(chunk).length
    else if (kind === -1) {
      const del = rga.localDelete(idx, Array.from(chunk).length)
      ops.push(...del)
    } else {
      const ins = rga.localInsert(idx, chunk)
      ops.push(...ins)
      idx += Array.from(chunk).length
    }
  }
  text.value = next
  for (const op of ops) sock?.send({ type: 'op', op })
}

async function saveTitle() {
  if (!doc.value) return
  await api.updateDoc(doc.value.id, { title: doc.value.title })
}
async function switchMode() {
  if (!doc.value) return
  await api.updateDoc(doc.value.id, { editorMode: mode.value })
}
async function toggleFav() {
  if (!doc.value) return
  const res = await api.favorite(doc.value.id)
  favorite.value = res.data.favorite
}
async function saveVer() {
  labelErr.value = label.value.trim() ? '' : '请填写版本说明'
  if (labelErr.value || !doc.value) return
  if (mode.value === 'markdown') {
    await api.updateDoc(doc.value.id, { contentMd: text.value })
  }
  await api.saveVersion(doc.value.id, label.value.trim())
  askSave.value = false
  label.value = ''
  const vs = await api.versions(doc.value.id)
  versions.value = vs.data
}
async function showDiff(id: string) {
  diffTarget = id
  const res = await api.diff(id, 'current')
  diff.value = res.data
}
async function doRollback() {
  if (!diffTarget) return
  await api.rollback(diffTarget)
  ElMessage.success('已按该版本创建新快照')
  await load()
  diff.value = null
}

watch(() => route.params.id, load, { immediate: true })
onBeforeUnmount(() => {
  sock?.close()
  editor.value?.destroy()
})
</script>

<style scoped>
.editor-page { display:flex; width:100%; min-height:calc(100vh - 64px); }
.canvas { flex:1; padding:20px 28px 48px; min-width:0; }
.meta { display:flex; gap:12px; align-items:center; flex-wrap:wrap; }
.title {
  flex:1; min-width:220px; border:0; background:transparent;
  font-family:"Fraunces","Noto Serif SC",serif; font-size:34px; font-weight:650;
}
.tools { display:flex; gap:8px; align-items:center; }
.faces { display:flex; gap:6px; }
.face {
  width:28px; height:28px; border-radius:50%; color:#fff; display:grid; place-items:center; font-size:12px;
}
.md {
  width:100%; min-height:62vh; margin-top:16px; border:1px solid var(--line);
  border-radius:18px; padding:22px; background:var(--card);
  font-size:16px; line-height:1.75; resize:vertical;
}
.rich { margin-top:16px; background:var(--card); border:1px solid var(--line); border-radius:18px; padding:22px; min-height:62vh; }
.rail { width:320px; border-left:1px solid var(--line); padding:16px; background:rgba(231,221,208,0.45); }
.ver {
  display:block; width:100%; text-align:left; margin:8px 0; padding:10px;
  border:1px solid var(--line); border-radius:12px; background:var(--card); cursor:pointer;
}
.ver small { display:block; color:var(--muted); }
.diff-box { margin-top:12px; }
input { width:100%; border:1px solid var(--line); border-radius:10px; padding:8px 10px; }
@media (max-width: 768px) { .rail { display:none; } }
</style>
