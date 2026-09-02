<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo">
          <svg viewBox="0 0 100 100" width="60" height="60">
            <defs>
              <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
              </linearGradient>
            </defs>
            <circle cx="50" cy="50" r="45" fill="url(#grad)"/>
            <text x="50" y="60" font-family="Arial" font-size="28" font-weight="bold" fill="white" text-anchor="middle">G</text>
          </svg>
        </div>
        <h1>GoAWD</h1>
        <p>EDR for CTF AWD</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
        <el-form-item prop="token">
          <el-input
            v-model="form.token"
            placeholder="请输入 Token"
            prefix-icon="Key"
            size="large"
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-button
          type="primary"
          size="large"
          :loading="loading"
          class="login-btn"
          @click="handleLogin"
        >
          登录
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ping } from '@/api/apis'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({ token: '' })

const rules = {
  token: [{ required: true, message: '请输入 Token', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return

    loading.value = true
    try {
      await ping()
      sessionStorage.setItem('token', form.token)
      ElMessage.success('登录成功')
      router.push('/main')
    } catch (error: any) {
      const status = error.response?.status
      if (status === 401 || status === 403) {
        ElMessage.error('Token 无效')
      } else {
        ElMessage.error('网络异常')
      }
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  document.querySelector<HTMLInputElement>('input')?.focus()
})
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo {
  margin-bottom: 16px;
}

.login-header h1 {
  margin: 0;
  font-size: 28px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.login-header p {
  margin: 8px 0 0;
  color: #909399;
}

.login-btn {
  width: 100%;
  margin-top: 16px;
}
</style>
