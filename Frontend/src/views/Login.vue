<template>
  <div class="login-container">
    <div class="login-bg">
      <div class="particles"></div>
    </div>
    <div class="login-card">
      <div class="login-header">
        <div class="logo">
          <svg viewBox="0 0 120 120" class="logo-svg">
            <defs>
              <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
              </linearGradient>
            </defs>
            <circle cx="60" cy="60" r="55" fill="url(#grad1)" />
            <text x="60" y="72" font-size="36" fill="white" text-anchor="middle" font-weight="bold">G</text>
          </svg>
        </div>
        <h1 class="title">GoAWD</h1>
        <p class="subtitle">轻量级 EDR for CTF AWD</p>
      </div>
      <el-form :model="ruleForm2" :rules="rules2" ref="ruleForm2" label-position="left" label-width="0px" class="login-form">
        <el-form-item prop="accessToken">
          <el-input 
            type="text" 
            v-model="ruleForm2.accessToken" 
            auto-complete="off" 
            placeholder="请输入 Access Token"
            prefix-icon="fa fa-key">
          </el-input>
        </el-form-item>
        <el-form-item style="width:100%;">
          <el-button 
            type="primary" 
            class="login-btn"
            @click.native.prevent="handleLogin" 
            :loading="logining">
            <span v-if="!logining">登 录</span>
            <span v-else>登录中...</span>
          </el-button>
        </el-form-item>
      </el-form>
      <div class="login-footer">
        <p>GoAWD - Lightweight EDR for CTF AWD</p>
      </div>
    </div>
  </div>
</template>

<script>
  import { HomeApi } from "../api/index.js";
  import Axios from 'axios';
  export default {
    data() {
      return {
        logining: false,
        ruleForm2: {
          accessToken: '',
        },
        rules2: {
          accessToken: [
            { required: true, message: '请输入 Access Token', trigger: 'blur' },
          ],
        },
        checked: true
      };
    },
    methods: {
      handleLogin() {
        this.$refs.ruleForm2.validate((valid) => {
          if (valid) {
            this.logining = true;
            sessionStorage.clear();
            Axios.defaults.headers['Token'] = this.ruleForm2.accessToken;
            HomeApi.ping()
              .then(res => {
                sessionStorage.setItem('accessToken', this.ruleForm2.accessToken);
                this.$router.push({
                  path:'/main'
                })
              }).catch(err => {
                Axios.defaults.headers['Token'] = "";
                this.logining = false;
                this.$message.error('登录失败，请检查 Token 是否正确');
              });
          }
        });
      }
    }
  }
</script>

<style lang="scss" scoped>
.login-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  overflow: hidden;
}

.login-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  
  .particles {
    position: absolute;
    width: 100%;
    height: 100%;
    background-image: 
      radial-gradient(ellipse at center, rgba(102, 126, 234, 0.15) 0%, transparent 70%);
  }
}

.login-card {
  position: relative;
  width: 400px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  z-index: 10;
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
  
  .logo {
    margin-bottom: 20px;
    
    .logo-svg {
      width: 80px;
      height: 80px;
      filter: drop-shadow(0 4px 8px rgba(102, 126, 234, 0.3));
    }
  }
  
  .title {
    font-size: 32px;
    font-weight: 700;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    margin-bottom: 8px;
  }
  
  .subtitle {
    font-size: 14px;
    color: #909399;
  }
}

.login-form {
  .el-form-item {
    margin-bottom: 24px;
  }
  
  .el-input {
    .el-input__inner {
      height: 48px;
      border-radius: 8px;
      border: 2px solid #e4e7ed;
      padding-left: 44px;
      transition: all 0.3s ease;
      
      &:focus {
        border-color: #667eea;
        box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
      }
    }
    
    .el-input__prefix {
      left: 14px;
      color: #c0c4cc;
    }
  }
  
  .login-btn {
    width: 100%;
    height: 48px;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 600;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    transition: all 0.3s ease;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
    }
    
    &:active {
      transform: translateY(0);
    }
  }
}

.login-footer {
  text-align: center;
  margin-top: 24px;
  
  p {
    font-size: 12px;
    color: #909399;
  }
}
</style>
