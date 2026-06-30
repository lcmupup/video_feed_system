import request from './request'

// 获取 Feed 流
export function getFeed(page = 1, pageSize = 10) {
  return request.get('/feed', { params: { page, page_size: pageSize } })
}

// 上传视频
export function uploadVideo(formData) {
  return request.post('/video/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

// 获取视频播放地址
export function getVideoPlayUrl(id) {
  return `/api/video/play/${id}`
}

// 搜索视频
export function searchVideos(keyword, page = 1, pageSize = 10) {
  return request.get('/video/search', { params: { keyword, page, page_size: pageSize } })
}
