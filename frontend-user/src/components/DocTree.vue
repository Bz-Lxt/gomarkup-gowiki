<template>
  <div class="tree">
    <TreeNode
      v-for="n in roots"
      :key="n.id"
      :node="n"
      :all="nodes"
      :current="current"
      :depth="0"
      @open="$emit('open', $event)"
      @move="$emit('move', $event)"
      @create="$emit('create', $event)"
      @remove="$emit('remove', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Document } from '../api/wiki'
import TreeNode from './TreeNode.vue'

const props = defineProps<{ nodes: Document[]; current: string }>()
defineEmits<{
  open: [id: string]
  move: [payload: { id: string; parentId: string | null; sortOrder: number }]
  create: [parentId: string]
  remove: [id: string]
}>()

const roots = computed(() => props.nodes.filter((n) => !n.parentId).sort((a, b) => a.sortOrder - b.sortOrder))
</script>

<style scoped>
.tree { margin-top: 8px; }
</style>
