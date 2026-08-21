<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal" style="width:min(640px,94vw)">
      <h3 class="serif">检索结果</h3>
      <div v-if="!hits.length" style="color:var(--muted)">没有匹配的文档。</div>
      <button v-for="h in hits" :key="h.id" class="hit" @click="$emit('open', h.id)">
        <strong>{{ h.title }}</strong>
        <p v-html="h.fragment"></p>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ hits: any[] }>()
defineEmits<{ close: []; open: [id: string] }>()
</script>

<style scoped>
.hit {
  display:block; width:100%; text-align:left; border:1px solid var(--line);
  background:var(--paper); border-radius:12px; padding:10px 12px; margin:8px 0; cursor:pointer;
}
.hit p { margin:6px 0 0; color:var(--muted); font-size:13px; }
</style>
