<template>
  <div class="user-page">
    <div class="user-header">
      <el-avatar :size="72" :src="user.avatar" />
      <div class="user-basic">
        <h2>{{ user.nickname || user.username }}</h2>
        <p class="user-bio">{{ user.bio || '这个人很懒，什么都没写' }}</p>
        <div class="user-stats">
          <span>关注 {{ followings.length }}</span>
          <span>粉丝 {{ followers.length }}</span>
        </div>
        <el-button
          v-if="userStore.isLoggedIn && userStore.userId !== userId"
          :type="isFollowing ? 'default' : 'primary'"
          @click="toggleFollow"
          :loading="followLoading"
        >
          {{ isFollowing ? '已关注' : '+ 关注' }}
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="关注列表" name="followings">
        <div v-if="followings.length === 0" class="empty">暂无关注</div>
        <div v-for="u in followings" :key="u.id" class="user-item" @click="$router.push(`/user/${u.id}`)">
          <el-avatar :size="40" :src="u.avatar" />
          <span>{{ u.nickname || u.username }}</span>
        </div>
      </el-tab-pane>
      <el-tab-pane label="粉丝列表" name="followers">
        <div v-if="followers.length === 0" class="empty">暂无粉丝</div>
        <div v-for="u in followers" :key="u.id" class="user-item" @click="$router.push(`/user/${u.id}`)">
          <el-avatar :size="40" :src="u.avatar" />
          <span>{{ u.nickname || u.username }}</span>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { followUser, unfollowUser, getFollowers, getFollowings } from '../api/follow'
import { useUserStore } from '../store/user'

const route = useRoute()
const userStore = useUserStore()
const userId = ref(Number(route.params.id))
const user = ref({})
const isFollowing = ref(false)
const followLoading = ref(false)
const activeTab = ref('followings')
const followers = ref([])
const followings = ref([])

async function loadData() {
  try {
    const [followerRes, followingRes] = await Promise.all([
      getFollowers(userId.value),
      getFollowings(userId.value),
    ])
    followers.value = followerRes.data || []
    followings.value = followingRes.data || []

    // 从列表中找到当前用户信息
    if (followers.value.length > 0) {
      const found = followers.value.find(u => u.id === userId.value)
      if (found) user.value = found
    }
    // 检测是否已关注
    if (userStore.isLoggedIn) {
      isFollowing.value = followings.value.some(u => u.id === userStore.userId) ||
                          followers.value.some(u => u.id === userStore.userId && isFollowing.value)
    }
  } catch {}
}

async function toggleFollow() {
  followLoading.value = true
  try {
    if (isFollowing.value) {
      await unfollowUser(userId.value)
      isFollowing.value = false
      ElMessage.success('已取消关注')
    } else {
      await followUser(userId.value)
      isFollowing.value = true
      ElMessage.success('关注成功')
    }
  } catch {} finally {
    followLoading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.user-page { max-width: 600px; margin: 0 auto; }
.user-header { display: flex; align-items: flex-start; gap: 20px; padding: 20px 0; border-bottom: 1px solid #eee; margin-bottom: 16px; }
.user-basic h2 { margin-bottom: 6px; }
.user-bio { color: #999; margin-bottom: 8px; }
.user-stats { display: flex; gap: 20px; margin-bottom: 12px; font-size: 13px; color: #666; }
.user-item { display: flex; align-items: center; gap: 12px; padding: 10px; cursor: pointer; border-radius: 8px; }
.user-item:hover { background: #f5f5f5; }
.empty { text-align: center; padding: 40px; color: #999; }
</style>
