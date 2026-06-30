<template>
  <div class="search-page">
    <h2 class="page-title">搜索结果：{{ keyword }}</h2>
    <div class="video-grid" v-loading="loading">
      <div v-for="video in videos" :key="video.id" class="video-item" @click="playVideo(video)">
        <div class="video-cover">
          <video v-if="playingId === video.id" :src="getPlayUrl(video.id)" controls autoplay class="video-player" @click.stop />
          <div v-else class="cover-placeholder"><el-icon :size="40"><VideoPlay /></el-icon></div>
        </div>
        <div class="video-info">
          <h3 v-html="video.highlight_title || video.title"></h3>
          <p v-html="video.highlight_desc || video.description || '暂无描述'"></p>
          <span class="meta">{{ video.created_at }}</span>
        </div>
      </div>
    </div>
    <div v-if="!loading && videos.length === 0" class="empty"><p>未找到相关视频</p></div>
    <div class="pagination" v-if="total > pageSize">
      <el-pagination background layout="prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="doSearch" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { searchVideos, getVideoPlayUrl } from '../api/video'

const route = useRoute()
const keyword = ref(route.query.keyword || '')
const videos = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const playingId = ref(null)

function getPlayUrl(id) { return getVideoPlayUrl(id) }
function playVideo(v) { playingId.value = playingId.value === v.id ? null : v.id }

async function doSearch() {
  if (!keyword.value) return
  loading.value = true
  try {
    const res = await searchVideos(keyword.value, page.value, pageSize.value)
    videos.value = res.data || []
    total.value = res.total || 0
  } catch {} finally {
    loading.value = false
  }
}

onMounted(doSearch)
</script>

<style scoped>
.search-page { max-width: 900px; margin: 0 auto; }
.video-grid { display: flex; flex-direction: column; gap: 12px; }
.video-item { display: flex; background: #fff; border-radius: 8px; overflow: hidden; cursor: pointer; box-shadow: 0 1px 4px rgba(0,0,0,0.08); }
.video-cover { width: 280px; min-height: 160px; background: #000; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.cover-placeholder { color: #fff; }
.video-player { width: 100%; height: 100%; object-fit: cover; }
.video-info { padding: 12px 16px; flex: 1; }
.video-info h3 { font-size: 15px; margin-bottom: 6px; }
.video-info p { color: #999; font-size: 13px; margin-bottom: 6px; }
.meta { font-size: 12px; color: #bbb; }
.empty { text-align: center; padding: 60px; color: #999; }
.pagination { display: flex; justify-content: center; margin-top: 24px; }
</style>
