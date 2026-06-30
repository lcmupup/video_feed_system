import request from './request'

// 注册
export function register(data) {
  return request.post('/user/register', data)
}

// 登录
export function login(data) {
  return request.post('/user/login', data)
}

// 获取个人信息
export function getProfile() {
  return request.get('/user/profile')
}

// 更新个人信息
export function updateProfile(data) {
  return request.put('/user/profile', data)
}

// 上传头像
export function uploadAvatar(file) {
  const formData = new FormData()
  formData.append('avatar', file)
  return request.post('/user/avatar', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
