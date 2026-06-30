<template>
  <div class="feed-page">
    <h2 class="page-title">推荐视频</h2>
    <div class="video-grid" v-loading="loading">
      <div v-for="video in videos" :key="video.id" class="video-card" @click="playVideo(video)">
        <div class="video-cover">
          <video
            v-if="playingId === video.id"
            :src="getPlayUrl(video.id)"
            controls
            autoplay
            class="video-player"
            @click.stop
          />
          <div v-else class="cover-placeholder">
            <el-icon :size="48"><VideoPlay /></el-icon>
          </div>
        </div>
        <div class="video-info">
          <h3 class="video-title">{{ video.title }}</h3>
          <p class="video-desc">{{ video.description || '暂无描述' }}</p>
          <div class="video-meta">
            <span>📁 {{ formatSize(video.file_size) }}</span>
            <span>👁 {{ video.view_count || 0 }}</span>
          </div>
          <div class="video-actions" @click.stop>
            <el-button
              :type="video.is_liked ? 'danger' : 'default'"
              size="small"
              text
              @click="toggleLike(video)"
            >
              <el-icon><StarFilled v-if="video.is_liked" /><Star v-else /></el-icon>
              {{ video.like_count || 0 }}
            </el-button>
            <el-button
              :type="video.is_favored ? 'warning' : 'default'"
              size="small"
              text
              @click="toggleFavorite(video)"
            >
              <el-icon><Collection v-if="video.is_favored" /><CollectionTag v-else /></el-icon>
              收藏
            </el-button>
            <el-button size="small" text @click="$router.push(`/user/${video.user_id || video.author?.id}`)">
              <el-icon><User /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </div>
    <div v-if="!loading && videos.length === 0" class="empty">
      <p>暂无视频</p>
    </div>
    <div class="pagination" v-if="total > pageSize">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="total"
        :page-size="pageSize"
        v-model:current-page="page"
        @current-change="loadFeed"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getFeed, getVideoPlayUrl } from '../api/video'
import { likeVideo, unlikeVideo, favoriteVideo, unfavoriteVideo, getLikeStatus, getFavoriteStatus } from '../api/interaction'
import { useUserStore } from '../store/user'
import { ElMessage } from 'element-plus'

const userStore = useUserStore()
const videos = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const playingId = ref(null)

function getPlayUrl(id) {
  return getVideoPlayUrl(id)
}

function playVideo(video) {
  if (playingId.value === video.id) {
    playingId.value = null
  } else {
    playingId.value = video.id
  }
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
  return size.toFixed(1) + ' ' + units[i]
}

async function toggleLike(video) {
  if (!userStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  try {
    if (video.is_liked) {
      await unlikeVideo(video.id)
      video.like_count = Math.max(0, (video.like_count || 0) - 1)
    } else {
      await likeVideo(video.id)
      video.like_count = (video.like_count || 0) + 1
    }
    video.is_liked = !video.is_liked
  } catch {}
}

async function toggleFavorite(video) {
  if (!userStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  try {
    if (video.is_favored) {
      await unfavoriteVideo(video.id)
    } else {
      await favoriteVideo(video.id)
    }
    video.is_favored = !video.is_favored
  } catch {}
}

async function loadFeed() {
  loading.value = true
  try {
    const res = await getFeed(page.value, pageSize.value)
    videos.value = (res.data || []).map(v => ({ ...v, is_liked: false, is_favored: false }))
    total.value = res.total || 0

    // 如果已登录，批量查询点赞/收藏状态
    if (userStore.isLoggedIn && videos.value.length > 0) {
      for (const video of videos.value) {
        try {
          const [likeRes, favRes] = await Promise.all([
            getLikeStatus(video.id),
            getFavoriteStatus(video.id),
          ])
          video.is_liked = likeRes.data?.is_liked || false
          video.like_count = likeRes.data?.like_count || video.like_count || 0
          video.is_favored = favRes.data?.is_favorited || false
        } catch {}
      }
    }
  } catch {} finally {
    loading.value = false
  }
}

onMounted(loadFeed)
</script>

<style scoped>
.feed-page {
  max-width: 900px;
  margin: 0 auto;
}
.video-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.video-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  display: flex;
}
.video-cover {
  width: 320px;
  min-height: 180px;
  background: #000;
  flex-shrink: 0;
}
.cover-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #fff;
  min-height: 180px;
}
.video-player {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.video-info {
  padding: 12px 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
}
.video-title {
  font-size: 16px;
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.video-desc {
  color: #999;
  font-size: 13px;
  margin-bottom: 8px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.video-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
}
.video-actions {
  display: flex;
  gap: 4px;
}
.empty {
  text-align: center;
  padding: 60px;
  color: #999;
}
.pagination {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>
