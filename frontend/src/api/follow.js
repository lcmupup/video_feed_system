import request from './request'

// 关注用户
export function followUser(userId) {
  return request.post(`/user/${userId}/follow`)
}

// 取消关注
export function unfollowUser(userId) {
  return request.delete(`/user/${userId}/follow`)
}

// 获取粉丝列表
export function getFollowers(userId, page = 1, pageSize = 10) {
  return request.get(`/user/${userId}/followers`, { params: { page, page_size: pageSize } })
}

// 获取关注列表
export function getFollowings(userId, page = 1, pageSize = 10) {
  return request.get(`/user/${userId}/followings`, { params: { page, page_size: pageSize } })
}
