import request from './request'

// 发送消息
export function sendMessage(toUserId, content) {
  return request.post('/chat/send', { to_user_id: toUserId, content })
}

// 获取聊天记录
export function getHistory(peerId, page = 1, pageSize = 20) {
  return request.get(`/chat/history/${peerId}`, { params: { page, page_size: pageSize } })
}

// 获取会话列表
export function getConversations() {
  return request.get('/chat/conversations')
}
