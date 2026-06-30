<template>
  <div class="profile-page">
    <h2 class="page-title">个人中心</h2>
    <div class="profile-content">
      <!-- 头像区 -->
      <el-card class="profile-card">
        <div class="avatar-section">
          <el-avatar :size="80" :src="userStore.avatar" />
          <el-upload
            :show-file-list="false"
            :before-upload="handleAvatarUpload"
            accept=".jpg,.jpeg,.png,.gif"
          >
            <el-button size="small" type="primary" text>更换头像</el-button>
          </el-upload>
        </div>
      </el-card>

      <!-- 信息编辑 -->
      <el-card class="profile-card">
        <template #header><span>编辑资料</span></template>
        <el-form :model="form" label-width="80px">
          <el-form-item label="用户名">
            <el-input v-model="userStore.user.username" disabled />
          </el-form-item>
          <el-form-item label="昵称">
            <el-input v-model="form.nickname" placeholder="请输入昵称" />
          </el-form-item>
          <el-form-item label="简介">
            <el-input v-model="form.bio" type="textarea" :rows="3" placeholder="介绍一下自己" />
          </el-form-item>
          <el-form-item label="性别">
            <el-radio-group v-model="form.gender">
              <el-radio :value="0">保密</el-radio>
              <el-radio :value="1">男</el-radio>
              <el-radio :value="2">女</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="生日">
            <el-input v-model="form.birthday" placeholder="如：2000-01-01" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="saveProfile">保存</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 收藏列表 -->
      <el-card class="profile-card">
        <template #header><span>我的收藏</span></template>
        <div v-if="favorites.length === 0" class="empty-text">暂无收藏</div>
        <div v-for="v in favorites" :key="v.id" class="fav-item" @click="$router.push('/')">
          <span>{{ v.title }}</span>
          <span class="fav-time">{{ v.created_at }}</span>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getProfile, updateProfile, uploadAvatar } from '../api/user'
import { getUserFavorites } from '../api/interaction'
import { useUserStore } from '../store/user'

const userStore = useUserStore()
const saving = ref(false)
const favorites = ref([])

const form = reactive({
  nickname: '',
  bio: '',
  gender: 0,
  birthday: '',
})

onMounted(async () => {
  try {
    const res = await getProfile()
    const u = res.data
    form.nickname = u.nickname || ''
    form.bio = u.bio || ''
    form.gender = u.gender || 0
    form.birthday = u.birthday || ''
    userStore.updateUser(u)
  } catch {}

  try {
    const res = await getUserFavorites()
    favorites.value = res.data || []
  } catch {}
})

async function saveProfile() {
  saving.value = true
  try {
    await updateProfile(form)
    userStore.updateUser(form)
    ElMessage.success('保存成功')
  } catch {} finally {
    saving.value = false
  }
}

async function handleAvatarUpload(file) {
  try {
    await uploadAvatar(file)
    // 重新拉取个人信息以获取新头像URL
    const res = await getProfile()
    userStore.updateUser(res.data)
    ElMessage.success('头像更新成功')
  } catch {}
  return false // 阻止默认上传
}
</script>

<style scoped>
.profile-page { max-width: 600px; margin: 0 auto; }
.profile-card { margin-bottom: 16px; }
.avatar-section { display: flex; align-items: center; gap: 16px; }
.fav-item { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #f5f5f5; cursor: pointer; }
.fav-item:hover { color: #409eff; }
.fav-time { color: #bbb; font-size: 12px; }
.empty-text { color: #999; text-align: center; padding: 20px; }
</style>
