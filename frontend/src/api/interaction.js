import request from './request'

// 点赞
export function likeVideo(videoId) {
  return request.post(`/video/${videoId}/like`)
}

// 取消点赞
export function unlikeVideo(videoId) {
  return request.delete(`/video/${videoId}/like`)
}

// 获取点赞状态
export function getLikeStatus(videoId) {
  return request.get(`/video/${videoId}/like`)
}

// 收藏
export function favoriteVideo(videoId) {
  return request.post(`/video/${videoId}/favorite`)
}

// 取消收藏
export function unfavoriteVideo(videoId) {
  return request.delete(`/video/${videoId}/favorite`)
}

// 获取收藏状态
export function getFavoriteStatus(videoId) {
  return request.get(`/video/${videoId}/favorite`)
}

// 获取收藏列表
export function getUserFavorites() {
  return request.get('/user/favorites')
}
