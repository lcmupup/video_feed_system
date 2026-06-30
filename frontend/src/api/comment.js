import request from './request'

// 获取视频评论
export function getComments(videoId, page = 1, pageSize = 10) {
  return request.get(`/video/${videoId}/comments`, { params: { page, page_size: pageSize } })
}

// 发表评论
export function createComment(videoId, content, parentId = null) {
  return request.post(`/video/${videoId}/comment`, { content, parent_id: parentId })
}

// 删除评论
export function deleteComment(commentId) {
  return request.delete(`/comment/${commentId}`)
}
