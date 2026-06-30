<template>
  <div class="upload-page">
    <h2 class="page-title">上传视频</h2>
    <el-card class="upload-card">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="视频标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入视频标题" />
        </el-form-item>
        <el-form-item label="视频描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入视频描述（选填）" />
        </el-form-item>
        <el-form-item label="视频文件" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="handleRemove"
            accept=".mp4,.avi,.mov,.wmv,.flv,.mkv"
            drag
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽视频文件到此处或<em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">支持 MP4/AVI/MOV/WMV/FLV/MKV，最大 100MB</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="uploading" @click="handleUpload" size="large">
            <el-icon><Upload /></el-icon> 上传
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadVideo } from '../api/video'

const formRef = ref(null)
const uploadRef = ref(null)
const uploading = ref(false)
const selectedFile = ref(null)

const form = reactive({ title: '', description: '' })
const rules = {
  title: [{ required: true, message: '请输入视频标题', trigger: 'blur' }],
}

function handleFileChange(file) {
  selectedFile.value = file.raw
}

function handleRemove() {
  selectedFile.value = null
}

async function handleUpload() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  if (!selectedFile.value) {
    ElMessage.warning('请选择视频文件')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('title', form.title)
    formData.append('description', form.description)
    formData.append('file', selectedFile.value)

    await uploadVideo(formData)
    ElMessage.success('文件已接收，正在处理中，稍后可在首页查看')
    form.title = ''
    form.description = ''
    selectedFile.value = null
    uploadRef.value?.clearFiles()
  } catch {} finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.upload-page {
  max-width: 600px;
  margin: 0 auto;
}
.upload-card {
  padding: 8px;
}
</style>
