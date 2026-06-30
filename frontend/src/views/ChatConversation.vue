<template>
  <div class="chat-conv-page">
    <div class="chat-header">
      <el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon> 返回</el-button>
      <h3>聊天</h3>
    </div>

    <div class="message-list" ref="msgListRef">
      <div v-for="msg in messages" :key="msg.id" class="msg-item" :class="{ 'msg-mine': msg.from_user_id === userStore.userId }">
        <div class="msg-bubble">
          <p>{{ msg.content }}</p>
          <span class="msg-time">{{ msg.created_at }}</span>
        </div>
      </div>
      <div v-if="messages.length === 0 && !loading" class="empty-msg">暂无消息，发送第一条吧</div>
    </div>

    <div class="input-area">
      <el-input v-model="inputText" placeholder="输入消息..." @keyup.enter="send" maxlength="500" show-word-limit>
        <template #append>
          <el-button type="primary" @click="send" :loading="sending">发送</el-button>
        </template>
      </el-input>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getHistory, sendMessage } from '../api/chat'
import { useUserStore } from '../store/user'

const route = useRoute()
const userStore = useUserStore()
const peerId = ref(Number(route.params.userId))
const messages = ref([])
const inputText = ref('')
const loading = ref(false)
const sending = ref(false)
const msgListRef = ref(null)

let ws = null

function scrollToBottom() {
  nextTick(() => {
    if (msgListRef.value) {
      msgListRef.value.scrollTop = msgListRef.value.scrollHeight
    }
  })
}

async function loadHistory() {
  loading.value = true
  try {
    const res = await getHistory(peerId.value)
    messages.value = (res.data || []).reverse()
    scrollToBottom()
  } catch {} finally {
    loading.value = false
  }
}

async function send() {
  const text = inputText.value.trim()
  if (!text) return
  sending.value = true
  try {
    await sendMessage(peerId.value, text)
    messages.value.push({
      id: Date.now(),
      from_user_id: userStore.userId,
      to_user_id: peerId.value,
      content: text,
      created_at: new Date().toLocaleString(),
      is_read: false,
    })
    inputText.value = ''
    scrollToBottom()
  } catch {} finally {
    sending.value = false
  }
}

function connectWS() {
  const token = localStorage.getItem('token')
  if (!token) return
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${location.host}/api/ws/chat?token=${token}`)
  ws.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data)
      if (data.type === 'new_message' && data.from === peerId.value) {
        messages.value.push({
          id: data.msg_id,
          from_user_id: data.from,
          to_user_id: userStore.userId,
          content: data.content,
          created_at: data.time,
          is_read: true,
        })
        scrollToBottom()
      }
    } catch {}
  }
}

onMounted(() => {
  loadHistory()
  connectWS()
})

onUnmounted(() => {
  if (ws) ws.close()
})
</script>

<style scoped>
.chat-conv-page { max-width: 600px; margin: 0 auto; height: calc(100vh - 120px); display: flex; flex-direction: column; }
.chat-header { display: flex; align-items: center; gap: 8px; padding: 8px 0; border-bottom: 1px solid #eee; margin-bottom: 12px; }
.message-list { flex: 1; overflow-y: auto; padding: 8px 0; }
.msg-item { display: flex; margin-bottom: 12px; }
.msg-mine { justify-content: flex-end; }
.msg-bubble { max-width: 70%; padding: 10px 14px; border-radius: 12px; background: #f0f0f0; }
.msg-mine .msg-bubble { background: #409eff; color: #fff; }
.msg-bubble p { margin-bottom: 4px; word-break: break-word; }
.msg-time { font-size: 11px; color: #999; }
.msg-mine .msg-time { color: rgba(255,255,255,0.7); }
.input-area { padding: 12px 0; border-top: 1px solid #eee; }
.empty-msg { text-align: center; padding: 40px; color: #999; }
</style>
