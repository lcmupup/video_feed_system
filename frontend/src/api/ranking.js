import request from './request'

// 获取热门排行榜
export function getRanking(limit = 50) {
  return request.get('/ranking', { params: { limit } })
}
