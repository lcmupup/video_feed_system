<template>
  <div class="chat-page">
    <h2 class="page-title">消息</h2>
    <div class="conversation-list" v-loading="loading">
      <div v-for="c in conversations" :key="c.user_id" class="conversation-item" @click="$router.push(`/chat/${c.user_id}`)">
        <el-avatar :size="48" :src="c.avatar" />
        <div class="conv-info">
          <div class="conv-top">
            <span class="conv-name">{{ c.nickname || c.username }}</span>
            <span class="conv-time">{{ c.last_time }}</span>
          </div>
          <div class="conv-bottom">
            <span class="conv-msg">{{ c.last_message }}</span>
            <el-badge v-if="c.unread_count > 0" :value="c.unread_count" type="danger" />
          </div>
        </div>
      </div>
      <div v-if="!loading && conversations.length === 0" class="empty">
        <p>暂无会话</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getConversations } from '../api/chat'

const conversations = ref([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getConversations()
    conversations.value = res.data || []
  } catch {} finally {
    loading.value = false
  }
})
</script>

<style scoped>
.chat-page { max-width: 600px; margin: 0 auto; }
.conversation-item { display: flex; align-items: center; padding: 14px 16px; background: #fff; border-radius: 8px; margin-bottom: 8px; cursor: pointer; gap: 12px; box-shadow: 0 1px 4px rgba(0,0,0,0.06); }
.conversation-item:hover { background: #f5f7fa; }
.conv-info { flex: 1; min-width: 0; }
.conv-top { display: flex; justify-content: space-between; margin-bottom: 4px; }
.conv-name { font-weight: 600; }
.conv-time { font-size: 12px; color: #bbb; }
.conv-bottom { display: flex; justify-content: space-between; align-items: center; }
.conv-msg { color: #999; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 400px; }
.empty { text-align: center; padding: 60px; color: #999; }
</style>
