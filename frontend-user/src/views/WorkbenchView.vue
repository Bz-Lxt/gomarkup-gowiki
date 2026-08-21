<template>
  <div class="wb">
    <section>
      <h2 class="serif">最近浏览</h2>
      <div class="grid">
        <article v-for="it in data.recents" :key="it.documentId" class="card item" @click="go(it.documentId)">
          <h3>{{ it.title }}</h3>
          <time>{{ it.at }}</time>
        </article>
        <p v-if="!data.recents?.length" class="empty">还没有阅读足迹。</p>
      </div>
    </section>
    <section>
      <h2 class="serif">我的收藏</h2>
      <div class="grid">
        <article v-for="it in data.favorites" :key="it.documentId" class="card item" @click="go(it.documentId)">
          <h3>{{ it.title }}</h3>
          <time>{{ it.at }}</time>
        </article>
        <p v-if="!data.favorites?.length" class="empty">点亮编辑页的星标即可收藏。</p>
      </div>
    </section>
    <section>
      <h2 class="serif">团队动态</h2>
      <ul class="feed">
        <li v-for="a in data.activities" :key="a.id">
          <b>{{ a.summary }}</b>
          <span>{{ a.at }}</span>
        </li>
        <li v-if="!data.activities?.length" class="empty">空间还很安静。</li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/wiki'

const router = useRouter()
const data = ref<any>({ recents: [], favorites: [], activities: [] })
onMounted(async () => {
  const res = await api.workbench()
  data.value = res.data
})
function go(id: string) { router.push(`/doc/${id}`) }
</script>

<style scoped>
.wb { padding: 22px 24px 48px; width:100%; }
.grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(220px,1fr)); gap:12px; }
.item { padding:16px; cursor:pointer; }
.item h3 { margin:0 0 8px; font-size:18px; }
.item time, .empty { color:var(--muted); font-size:13px; }
.feed { list-style:none; padding:0; }
.feed li { display:flex; justify-content:space-between; padding:10px 0; border-bottom:1px solid var(--line); }
</style>
