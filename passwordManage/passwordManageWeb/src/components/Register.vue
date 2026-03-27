<template>
  <div class="container">
    <h2>注册</h2>
    <form id="form-register">
      <input type="text" v-model="username" placeholder="账号名（必填）" required />
      <input type="text" v-model="password" placeholder="密码（必填）" required />

      <!-- 使用v-model绑定到customInfoFields数组中的每个元素 -->
      <div v-for="(customInfo, index) in customInfoFields" :key="index" class="custom-info-container">
        <input type="text" v-model="customInfoFields[index]" class="input-register-custom-info"
          :placeholder="'自定义信息 ' + (index + 1)" required title="自定义信息用作恢复密码时认证，不设置则无法找回，请设置合适个数，丢失任意一个无法找回！" />
        <button v-if="index > 0" type="button" class="btn-delete-custom-info" @click="removeCustomInfo(index)">
          删除
        </button>
      </div>

      <!-- 增加自定义信息按钮 -->
      <button type="button" class="btn-add-custom-info" @click="addCustomInfoField">增加自定义信息</button>
      <button type="button" class="btn-submit" @click="register">注册</button>
    </form>

    <!-- 切换至登录 -->
    <router-link to="/">
      <button class="btn-switch"><span>已有账号？</span>去登录</button>
    </router-link>
  </div>
</template>

<script setup>
  import { ref } from 'vue';
  import { isAxiosError } from 'axios';
  import http from '@/js/http.js';
  import { ElMessage } from 'element-plus';
  import { useRouter } from 'vue-router';
  import keyExchange from '@/js/keyExchange.js';
  import { getServerAddress } from '@/js/getServerAddress.js';

  const router = useRouter();

  // 响应式变量
  const username = ref('');
  const password = ref('');
  const customInfoFields = ref(['']); // 初始自定义信息字段

  // 添加自定义信息字段
  function addCustomInfoField() {
    customInfoFields.value.push(''); // 增加一个空字段
  }

  // 移除指定的自定义信息字段
  function removeCustomInfo(index) {
    customInfoFields.value.splice(index, 1); // 删除指定的自定义信息字段
  }

  // 注册函数
  const register = async () => {
    // 从 localStorage 获取服务器地址
    const serverAddress = getServerAddress();
    // Elmessage显示服务器地址
    ElMessage({
      message: '服务器地址' + serverAddress,
      type: 'info',
      grouping: true,
    })

    // 1. 获取输入框中的数据并组装 JSON 数据
    const registerData = {
      username: username.value,
      password: password.value,
      customInfo: customInfoFields.value.filter(info => info !== ''), // 过滤空的自定义信息
    };
    const jsonData = JSON.stringify(registerData);

    // 获取临时 AES 密钥
    const tmpAesKey = await keyExchange.getAesKey();

    // 加密注册信息
    const encryptedResult = await keyExchange.encryptData(jsonData, tmpAesKey);

    // 准备请求负载
    const payload = {
      iv: encryptedResult.iv,
      encryptedData: encryptedResult.encryptedData
    };

    // 发送 POST 请求
    const response = await http.post("/register", payload, {
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json'  // 显式声明确保兼容性
      }
    });

    // 响应数据
    const data = response.data;
    // 获取data字段值
    const result = data.data;
    // 解码和解密返回的数据
    const iv = Uint8Array.from(atob(result.iv), c => c.charCodeAt(0));
    const encryptedPEM = Uint8Array.from(atob(result.encryptedPEM), c => c.charCodeAt(0));
    const decryptedPEM = await keyExchange.decryptAES(encryptedPEM, tmpAesKey, iv);
    //console.log('解密后的 PEM 数据:', decryptedPEM);

    // 自动下载 PEM 文件
    downloadPEM(decryptedPEM);

    // 处理注册状态
    if (data.code === 200 && data.message === 'Success') {
      ElMessage({
        message: "注册成功!",
        type: 'success',
        grouping: true,
      })
      router.push('/');  // 路由跳转到登录页面
    } else {
      ElMessage({
        message: "注册失败: " + (data.message || "未知错误"),
        type: 'error',
        grouping: true,
      })
    }

  };

  // 自动下载 PEM 文件
  function downloadPEM(decryptedPEM) {
    const blob = new Blob([decryptedPEM], { type: 'application/x-pem-file' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'keys.pem';
    link.click();
    window.URL.revokeObjectURL(url);
  }
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

  input[type="text"],
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

  .btn-submit,
  .btn-add-custom-info {
    padding: 12px 16px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    width: 100%;
    margin-top: 12px;
    color: white;
    font-size: 16px;
    font-weight: 500;
    transition: all 0.3s ease;
  }

  .btn-submit {
    background-color: var(--color-success);
  }

  .btn-submit:hover {
    background-color: var(--color-success-hover);
    box-shadow: 0 4px 12px rgba(0, 180, 42, 0.3);
  }

  .btn-add-custom-info {
    background-color: var(--color-primary);
    margin-top: 16px;
  }

  .btn-add-custom-info:hover {
    background-color: var(--color-primary-hover);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }

  .btn-delete-custom-info {
    background-color: var(--color-danger);
    color: white;
    border: none;
    border-radius: 6px;
    padding: 6px 12px;
    margin-left: 10px;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.3s ease;
  }

  .btn-delete-custom-info:hover {
    background-color: var(--color-danger-hover);
    box-shadow: 0 2px 8px rgba(245, 63, 63, 0.3);
  }

  .btn-switch {
    background: none;
    border: none;
    color: var(--color-primary);
    cursor: pointer;
    font-size: 14px;
    text-decoration: none;
    transition: color 0.3s ease;
    width: 100%;
    text-align: center;
    margin-top: 20px;
    padding: 8px;
  }

  .btn-switch:hover {
    color: var(--color-primary-hover);
    transform: none;
    box-shadow: none;
  }

  .custom-info-container {
    display: flex;
    align-items: center;
    margin-top: 12px;
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .container {
      width: 90vw;
      padding: 24px;
    }
  }

  /* 表单标题提示样式 */
  .input-register-custom-info[title]:hover::after {
    content: attr(title);
    position: absolute;
    background-color: rgba(0, 0, 0, 0.7);
    color: white;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    z-index: 1000;
    white-space: nowrap;
    transform: translateY(-100%);
    margin-top: -5px;
  }

  /* 按钮禁用状态 */
  button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
  }

  /* 动画效果 */
  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }

    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .container {
    animation: fadeIn 0.5s ease-out;
  }

  .custom-info-container {
    animation: fadeIn 0.3s ease-out;
  }

  .input-register-custom-info {
    flex-grow: 1;
    padding: 10px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    margin-right: 8px;
    width: calc(100% - 60px);
    display: inline-block;
  }

  .btn-delete-custom-info {
    background-color: var(--color-danger);
    padding: 10px;
    border: none;
    border-radius: 4px;
    color: white;
    cursor: pointer;
    width: 60px;
  }

  .btn-delete-custom-info:hover {
    background-color: var(--color-danger-hover);
  }

  .custom-info-container .btn-delete-custom-info {
    background-color: var(--color-danger);
    color: white;
    border: none;
    padding: 6px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }

  .custom-info-container .btn-delete-custom-info:hover {
    background-color: var(--color-danger-hover);
  }
</style>