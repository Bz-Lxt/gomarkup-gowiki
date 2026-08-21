<template>
  <div class="diff">
    <div class="col">
      <h4>行级</h4>
      <pre><span v-for="(s,i) in line" :key="'l'+i" :class="s.kind">{{ prefix(s.kind) }}{{ s.text }}\n</span></pre>
    </div>
    <div class="col">
      <h4>字符级</h4>
      <p class="char"><span v-for="(s,i) in chars" :key="'c'+i" :class="s.kind">{{ s.text }}</span></p>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ line: { kind: string; text: string }[]; chars: { kind: string; text: string }[] }>()
function prefix(k: string) {
  if (k === 'insert') return '+ '
  if (k === 'delete') return '- '
  return '  '
}
</script>

<style scoped>
.diff { display:grid; grid-template-columns:1fr 1fr; gap:12px; }
pre, .char { background:#fff; border:1px solid var(--line); border-radius:12px; padding:10px; white-space:pre-wrap; font-family:"IBM Plex Mono", monospace; font-size:12px; }
.insert { background:#d7ebd8; }
.delete { background:#f4d6d0; text-decoration:line-through; }
@media (max-width: 768px) { .diff { grid-template-columns:1fr; } }
</style>
