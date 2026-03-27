<template>
  <div class="container">
    <h2>登录</h2>
    <!-- 服务器配置部分 -->
    <div class="input-group">
      <input type="text" v-model="savedServerAddress" placeholder="服务器地址，例如：example.com" required />
      <input type="text" v-model="savedServerPort" placeholder="服务器端口，默认8443" required />
      <button type="button" class="submit" @click="saveServerConfig">保存服务器配置</button>
    </div>

    <div style="margin: 20px 0;">
      <el-segmented v-model="isPasswordLogin" :options="[
          { label: '账号密码登录', value: true },
          { label: '密钥登录', value: false }
        ]" @change="handleSegmentedChange" style="width: 100%; max-width: 400px;" />
    </div>
    <form @submit.prevent>
      <!-- 账号密码登录表单 -->
      <div v-if="isPasswordLogin" class="input-group">
        <input type="text" v-model="username" @keydown.enter="loginWithPassword" placeholder="账号名" required />
        <input type="password" v-model="password" @keydown.enter="loginWithPassword" placeholder="密码" required />
        <button type="button" class="submit" @click="loginWithPassword">账号密码登录</button>
      </div>
      <!-- 密钥登录表单 -->
      <div v-else class="input-group">
        <input type="text" v-model="username" placeholder="账号名" required />
        <input type="file" ref="privateKeyFile" accept=".key" />
        <button type="button" class="submit" @click="loginWithPrivateKey">上传密钥登录</button>
      </div>
    </form>
    <div class="button-container">
      <router-link to="/register">
        <button class="switch"><span>没有账号？</span>去注册</button>
      </router-link>
      <router-link to="/">
        <button class="switch"><span>忘记密码？</span>找回密码</button>
      </router-link>
    </div>
  </div>

  <!-- TOTP验证对话框 -->
  <el-dialog v-model="totpDialogVisible" title="TOTP验证" width="30%" :before-close="handleTotpDialogClose">
    <el-form label-width="80px">
      <el-form-item label="用户名">
        <el-input v-model="totpUsername" disabled></el-input>
      </el-form-item>
      <el-form-item label="TOTP码">
        <el-input v-model="totpCode" placeholder="请输入6位或8位数字验证码" maxlength="8" show-word-limit></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleTotpDialogClose">取消</el-button>
      <el-button type="primary" @click="verifyTotpCode">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import keyExchange from '@/js/keyExchange.js'
  import { checkSession } from '@/js/checkSession.js';
  import { ElMessage } from 'element-plus';
  import { getServerAddress } from '@/js/getServerAddress.js';
  import { useAuthStore } from '@/store/auth';
  import http from '@/js/http.js'
  //import { isAxiosError } from 'axios';

  // 响应式数据
  const isPasswordLogin = ref(true); // 控制登录方式切换
  const username = ref('');
  const password = ref('');
  const router = useRouter();
  const savedServerAddress = ref('');
  const savedServerPort = ref('');
  // TOTP验证相关
  const totpDialogVisible = ref(false);
  const totpUsername = ref('');
  const totpCode = ref('');

  // 切换登录方式（适配el-segmented组件）
  const handleSegmentedChange = (value) => {
    // 这里可以添加额外的逻辑，如果需要
    console.log('登录方式已切换为:', value ? '账号密码登录' : '密钥登录');
  };

  // 保留原方法以保持兼容性
  const toggleLoginMethod = (method) => {
    isPasswordLogin.value = (method === 'password');
  };

  // 账号密码登录
  const loginWithPassword = async () => {
    // 判空处理
    if (!username.value || !password.value) {
      ElMessage({
        message: '账号和密码不能为空',
        type: 'error',
        grouping: true,
      })
      return;
    }
    // 交换临时通信密钥
    const tmpAesKey = await keyExchange.getAesKey();

    // 构建登录数据
    const loginData = {
      username: username.value,
      password: password.value,
    };
    const loginJsonDataStr = JSON.stringify(loginData);

    // 加密登录信息
    const encryptedResult = await keyExchange.encryptData(loginJsonDataStr, tmpAesKey);

    // 组合加密数据和iv
    const loginPayload = {
      iv: encryptedResult.iv,   // 初始化向量 (IV)
      encryptedData: encryptedResult.encryptedData // 加密数据
    };

    const authStore = useAuthStore();

    // 发送登录信息到服务器
    const response = await http.post(`/login`, loginPayload, {
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    // 处理成功响应
    const data = response.data;

    if (data.message === 'Success') {
      ElMessage({
        message: '登录成功！',
        type: 'success',
        grouping: true,
      })
      authStore.login(username.value, ["user.isLogin"]);
      router.push('/set'); // 跳转到设置页面
    } else if (data.message === 'TOTP') {
      // 打开TOTP验证对话框
      totpUsername.value = username.value;
      totpCode.value = '';
      totpDialogVisible.value = true;
    } else {
      authStore.logout();
      ElMessage({
        message: data.message || '未知错误',
        type: 'error',
        grouping: true,
      })
    }
  };

  // 密钥文件登录
  const loginWithPrivateKey = () => {
    const file = $refs.privateKeyFile.files[0];
    if (file) {
      ElMessage({
        message: "使用密钥文件登录，待开发",
        type: 'error',
        grouping: true,
      })
    } else {
      ElMessage({
        message: "请选择密钥文件",
        type: 'error',
        grouping: true,
      })
    }
  };

  // 页面加载时检查会话
  onMounted(async () => {
    const localSavedServerAddress = localStorage.getItem('savedServerAddress');
    const localSavedServerPort = localStorage.getItem('savedServerPort');
    if (localSavedServerAddress && localSavedServerPort) {
      savedServerAddress.value = localSavedServerAddress;
      savedServerPort.value = localSavedServerPort;
    } else {
      savedServerAddress.value = window.location.hostname;
      savedServerPort.value = 8443;
    }
    const isLogin = await checkSession();
    if (isLogin) {
      router.push('/set');
    } else {
      ElMessage({
        message: '会话过期，请重新登录',
        type: 'info',
        grouping: true,
      })
    };
  });

  const saveServerConfig = () => {
    // 保存服务器地址和端口到 localStorage
    localStorage.setItem('savedServerAddress', savedServerAddress.value);
    localStorage.setItem('savedServerPort', savedServerPort.value);
    ElMessage({
      message: '服务器配置已保存',
      type: 'success',
      grouping: true,
    })
  };

  // 处理TOTP对话框关闭
  const handleTotpDialogClose = () => {
    totpDialogVisible.value = false;
    // 重置TOTP相关数据
    totpUsername.value = '';
    totpCode.value = '';
  };

  // 验证TOTP码
  const verifyTotpCode = async () => {
    // 判空处理
    if (!totpCode.value || (totpCode.value.length !== 6 && totpCode.value.length !== 8) || !/^\d+$/.test(totpCode.value)) {
      ElMessage({
        message: '请输入6位或8位数字TOTP验证码',
        type: 'error',
        grouping: true,
      })
      return;
    }

    // 交换临时通信密钥
    const tmpAesKey = await keyExchange.getAesKey();

    // 构建TOTP验证数据
    const totpData = {
      username: totpUsername.value,
      code: totpCode.value
    };
    const totpJsonDataStr = JSON.stringify(totpData);

    // 加密TOTP验证信息
    const encryptedResult = await keyExchange.encryptData(totpJsonDataStr, tmpAesKey);

    // 组合加密数据和iv
    const totpPayload = {
      iv: encryptedResult.iv,
      encryptedData: encryptedResult.encryptedData
    };

    const authStore = useAuthStore();

    try {
      // 发送TOTP验证信息到服务器
      const response = await http.post(`/verifyTotp`, totpPayload, {
        withCredentials: true,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // 处理响应
      const data = response.data;
      if (data.message === 'Success') {
        ElMessage({
          message: "TOTP验证成功！",
          type: 'success',
          grouping: true,
        })
        totpDialogVisible.value = false;
        authStore.login(totpUsername.value, ["user.isLogin"]);
        router.push('/set'); // 跳转到设置页面
      } else {
        ElMessage({
          message: data.message || 'TOTP验证失败',
          type: 'error',
          grouping: true,
        })
      }
    } catch (error) {
      ElMessage({
        message: 'TOTP验证异常：' + error,
        type: 'error',
        grouping: true,
      })
    }
  };

</script>

<style scoped>
  .container {
    background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    border: 1px solid var(--border-color-light);
    padding: 32px;
    width: 400px;
    max-width: 90vw;
    position: relative;
    transition: all 0.3s ease;
  }

  .container:hover {
    box-shadow: 0 6px 24px rgba(0, 0, 0, 0.2);
  }

  h2 {
    text-align: center;
    margin-bottom: 24px;
    color: var(--text-main);
    font-family: 'Inter', sans-serif;
    font-size: 24px;
    font-weight: 600;
  }

  .toggle-buttons {
    display: flex;
    justify-content: center;
    margin-bottom: 20px;
    gap: 8px;
  }

  .toggle-buttons button {
    padding: 10px 20px;
    border: 1px solid var(--border-color-light);
    border-radius: 6px;
    cursor: pointer;
    background-color: var(--bg-muted);
    color: var(--text-secondary);
    transition: all 0.3s ease;
    font-weight: 500;
  }

  .toggle-buttons button.active {
    background-color: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
  }

  input[type="text"],
  input[type="file"],
  input[type="password"] {
    width: 100%;
    padding: 12px 16px;
    margin: 12px 0;
    border: 1px solid var(--border-color-light);
    border-radius: 6px;
    box-sizing: border-box;
    font-size: 16px;
    background-color: var(--bg-container);
    color: var(--text-main);
    transition: all 0.3s ease;
  }

  input:focus {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(22, 93, 255, 0.1);
    outline: none;
  }

  .input-group {
    margin-bottom: 20px;
  }

  button.submit {
    width: 100%;
    padding: 12px;
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
    margin-top: 8px;
  }

  button.submit:hover {
    background-color: var(--color-primary-hover);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }

  .button-container {
    display: flex;
    justify-content: space-between;
    margin-top: 20px;
  }

  .switch {
    background: none;
    border: none;
    color: var(--color-primary);
    cursor: pointer;
    font-size: 14px;
    text-decoration: none;
    transition: color 0.3s ease;
  }

  .switch:hover {
    color: var(--color-primary-hover);
    transform: none;
    box-shadow: none;
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .container {
      width: 90vw;
      padding: 24px;
    }

    .button-container {
      flex-direction: column;
      gap: 10px;
    }

    .switch {
      width: 100%;
      text-align: center;
    }
  }

  button.submit {
    padding: 10px;
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    width: 100%;
    margin-top: 10px;
  }

  button.submit:hover {
    background-color: var(--color-primary-hover);
  }

  .input-group input {
    display: block;
    width: 100%;
    padding: 8px;
    margin: 10px 0;
  }

  .button-container {
    margin-top: 20px;
    display: flex;
    justify-content: center;
    gap: 20px;
    flex-wrap: wrap;
  }

  button.switch {
    background: none;
    color: var(--color-primary);
    border: none;
    padding: 8px 15px;
    cursor: pointer;
    text-decoration: underline;
    margin-top: 5px;
    font-size: 14px;
  }

  .switch:hover {
    color: var(--color-primary-hover);
    transform: none;
    box-shadow: none;
  }

  /* 确保el-segmented组件与输入框样式协调 */
  :deep(.el-segmented) {
    margin: 0 auto;
  }

  /* 优化表单元素间距 */
  .input-group {
    margin-top: 20px;
  }
</style>