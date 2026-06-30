<template>
  <div class="ranking-page">
    <h2 class="page-title">🏆 热门排行榜</h2>
    <div class="ranking-list" v-loading="loading">
      <div v-for="(item, index) in list" :key="item.video_id" class="ranking-item" @click="$router.push('/')">
        <div class="rank-badge" :class="'rank-' + (index + 1)">
          {{ index + 1 }}
        </div>
        <div class="rank-info">
          <h3>{{ item.title }}</h3>
          <div class="rank-meta">
            <span>👍 {{ item.like_count }} 赞</span>
            <span>👤 {{ item.author?.nickname || item.author?.username || '未知' }}</span>
          </div>
        </div>
        <div class="rank-count">
          <el-tag type="danger" size="large">{{ item.like_count }} 赞</el-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getRanking } from '../api/ranking'

const list = ref([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getRanking(50)
    list.value = res.data || []
  } catch {} finally {
    loading.value = false
  }
})
</script>

<style scoped>
.ranking-page { max-width: 700px; margin: 0 auto; }
.ranking-item { display: flex; align-items: center; padding: 14px 16px; background: #fff; border-radius: 8px; margin-bottom: 8px; cursor: pointer; box-shadow: 0 1px 4px rgba(0,0,0,0.06); transition: transform 0.15s; }
.ranking-item:hover { transform: translateX(4px); }
.rank-badge { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 16px; color: #fff; flex-shrink: 0; background: #999; }
.rank-1 { background: #f56c6c; font-size: 20px; }
.rank-2 { background: #e6a23c; font-size: 18px; }
.rank-3 { background: #f0ad4e; font-size: 17px; }
.rank-info { flex: 1; margin-left: 12px; }
.rank-info h3 { font-size: 15px; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 400px; }
.rank-meta { font-size: 12px; color: #999; display: flex; gap: 16px; }
.rank-count { flex-shrink: 0; }
</style>
