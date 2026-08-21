<template>
  <div>
    <div
      class="row"
      :class="{ active: current === node.id, over: over }"
      :style="{ paddingLeft: 10 + depth * 14 + 'px' }"
      draggable="true"
      @click="$emit('open', node.id)"
      @dragstart="onStart"
      @dragover.prevent="over = true"
      @dragleave="over = false"
      @drop.prevent="onDrop"
    >
      <span class="dot">{{ children.length ? '▾' : '·' }}</span>
      <span class="title">{{ node.title }}</span>
      <button class="mini" @click.stop="$emit('create', node.id)">+</button>
      <button class="mini" @click.stop="confirmRemove">×</button>
    </div>
    <TreeNode
      v-for="c in children"
      :key="c.id"
      :node="c"
      :all="all"
      :current="current"
      :depth="depth + 1"
      @open="$emit('open', $event)"
      @move="$emit('move', $event)"
      @create="$emit('create', $event)"
      @remove="$emit('remove', $event)"
    />
    <div v-if="ask" class="modal-mask" @click.self="ask=false">
      <div class="modal">
        <h3 class="serif">移入回收站？</h3>
        <p>文档「{{ node.title }}」将可从回收站恢复。</p>
        <div style="display:flex;gap:8px;justify-content:flex-end">
          <button class="btn btn-ghost" @click="ask=false">取消</button>
          <button class="btn btn-primary" @click="doRemove">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Document } from '../api/wiki'

const props = defineProps<{ node: Document; all: Document[]; current: string; depth: number }>()
const emit = defineEmits<{
  open: [id: string]
  move: [payload: { id: string; parentId: string | null; sortOrder: number }]
  create: [parentId: string]
  remove: [id: string]
}>()

const over = ref(false)
const ask = ref(false)
const children = computed(() =>
  props.all.filter((n) => n.parentId === props.node.id).sort((a, b) => a.sortOrder - b.sortOrder),
)

function onStart(e: DragEvent) {
  e.dataTransfer?.setData('text/plain', props.node.id)
}
function onDrop(e: DragEvent) {
  over.value = false
  const id = e.dataTransfer?.getData('text/plain')
  if (!id || id === props.node.id) return
  emit('move', { id, parentId: props.node.id, sortOrder: children.value.length })
}
function confirmRemove() { ask.value = true }
function doRemove() {
  ask.value = false
  emit('remove', props.node.id)
}
</script>

<style scoped>
.row {
  display:flex; align-items:center; gap:6px;
  padding:7px 8px; border-radius:10px; cursor:pointer; user-select:none;
}
.row:hover { background: rgba(196,92,38,0.08); }
.row.active { background: #fff; outline: 1px solid var(--terracotta); }
.row.over { outline: 1px dashed var(--terracotta); transform: scale(0.98); }
.title { flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.dot { width:14px; color:var(--muted); }
.mini {
  border:0; background:transparent; color:var(--muted); cursor:pointer; border-radius:6px;
}
</style>
