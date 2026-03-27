<template>
  <div class="settings-page">
    <!-- 页面标题 -->
    <div class="header">
      <h1>系统设置</h1>
      <p>管理您的密码管理器系统设置，优化使用体验。</p>
    </div>

    <!-- 设置区域 -->
    <div class="settings-sections">
      <!-- 系统名称设置 -->
      <section class="setting-section">
        <h2>基本配置</h2>
        <el-form label-width="120px">
          <el-form-item label="系统名称">
            <el-input v-model="systemNameTmp" placeholder="请输入系统名称" size="medium"></el-input>
          </el-form-item>
        </el-form>
        <!-- 系统名称保存按钮 -->
        <div class="saveBtn">
          <el-button type="primary" size="medium" @click="saveSystemName">保存系统名称</el-button>
        </div>
      </section>

      <!-- 安全设置 -->
      <section class="setting-section">
        <h2>安全设置</h2>
        <el-form label-width="120px">
          <el-form-item label="会话超时">
            <el-input-number v-model="timeout" :min="1" :max="21600" size="medium"></el-input-number>
            <span class="unit">秒</span>
          </el-form-item>
        </el-form>
        <!-- 安全设置保存按钮 -->
        <div class="saveBtn">
          <el-button type="primary" size="medium" @click="saveSecuritySettings">保存安全设置</el-button>
        </div>
      </section>

      <!-- 密钥设置 -->
      <section class="setting-section">
        <h2>密钥文件配置
          <el-tooltip content="上传注册时生成的密钥文件（可自动解析）或手动输入。同步到服务器可使用完整功能，同步到前端只有前端加解密可用。" placement="top">
            <span style="margin-left: 10px; color: red;">
              ?
            </span>
          </el-tooltip>
        </h2>
        <el-tabs type="card" @tab-click="handleTabChange">
          <!-- 文件上传标签 -->
          <el-tab-pane label="文件上传" name="upload">
            <el-form>
              <el-form-item>
                <div class="file-upload-area">
                  <el-upload drag action="''" :before-upload="handleBeforeUpload" :limit="1" :on-exceed="handleExceed">
                    <div class="upload-container">
                      <p class="upload-text">点击或拖拽上传密钥文件</p>
                    </div>
                  </el-upload>
                </div>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- 手动输入标签 -->
          <el-tab-pane label="手动输入" name="manual">
            <el-form label-width="120px">
              <el-form-item label="输入密钥:">
                <el-input v-model="aesKey" placeholder="密钥Hex:" size="medium" />
              </el-form-item>
              <el-form-item label="密钥Base64:">
                <el-input v-model="aesKeyBase64" placeholder="密钥Base64:" size="medium" />
              </el-form-item>

            </el-form>
          </el-tab-pane>
        </el-tabs>

        <!-- 保存按钮 -->
        <div class="saveBtn">
          <el-button type="primary" size="medium" @click="saveKeyToWeb">仅配置到前端</el-button>
          <el-button type="primary" size="medium" @click="saveKeyToCS">同时配置到服务端</el-button>
          <el-button type="danger" size="medium" @click="clearKey">清除密钥</el-button>
        </div>
      </section>

      <!-- 天气 API 设置 -->
      <section class="setting-section">
        <h2>天气 API 设置</h2>
        <el-form label-width="120px">
          <el-form-item label="天气 API Key">
            <el-input v-model="weatherApiKey" placeholder="请输入天气 API Key" size="medium"></el-input>
          </el-form-item>
        </el-form>
        <!-- 天气 API 保存按钮 -->
        <div class="saveBtn">
          <el-button type="primary" size="medium" @click="saveWeatherApiKey">保存天气 API Key</el-button>
          <el-button type="danger" size="medium" @click="clearWeatherApiKey">清除天气 API Key</el-button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
  import { reactive, inject, onMounted, ref, watch } from "vue";
  import { UploadFilled } from '@element-plus/icons-vue'
  import { getServerAddress } from '@/js/getServerAddress.js';
  import http from '@/js/http.js';
  import { isAxiosError } from 'axios';

  const systemName = inject('systemName');// 系统名称配置
  const timeout = inject('timeout');// 会话超时时间（分钟）
  const weatherApiKey = inject('weatherApiKey');// 天气 API Key

  const aesKey = ref('');// 手动输入的密钥
  const fileList = reactive([]);// 上传文件列表
  const aesKeyBase64 = ref('');// 手动输入的密钥Base64
  const systemNameTmp = ref('');// 系统名称配置临时变量
  const timeoutTmp = ref('');// 会话超时时间（分钟）临时变量
  const weatherApiKeyTmp = ref('');// 天气 API Key临时变量

  // 保存天气 API Key（仅保存到本地，不发送到服务器）
  const saveWeatherApiKey = () => {
    // 保存到本地存储
    localStorage.setItem('weatherApiKey', weatherApiKey.value);
    ElMessage({
      message: '天气 API Key 已保存到本地',
      type: 'success',
      grouping: true,
    })

    // 触发自定义事件，通知天气组件刷新数据
    const event = new CustomEvent('weatherApiKeyChanged', { detail: { newKey: weatherApiKey.value } });
    window.dispatchEvent(event);
  };

  // 清除天气 API Key
  const clearWeatherApiKey = () => {
    // 清除本地存储中的天气 API Key
    localStorage.removeItem('weatherApiKey');
    // 清空输入框
    weatherApiKey.value = '';
    ElMessage({
      message: '天气 API Key 已清除',
      type: 'success',
      grouping: true,
    })

    // 触发自定义事件，通知天气组件刷新数据
    const event = new CustomEvent('weatherApiKeyChanged', { detail: { newKey: null } });
    window.dispatchEvent(event);
  };


  watch(systemName, val => {
    if (val !== undefined && val !== null) {
      systemNameTmp.value = val
    }
  }, { immediate: true })

  watch(timeout, val => {
    if (val !== undefined && val !== null) {
      timeoutTmp.value = val
    }
  }, { immediate: true })

  watch(weatherApiKey, val => {
    if (val !== undefined && val !== null) {
      weatherApiKeyTmp.value = val
    }
  }, { immediate: true })
  onMounted(() => {
    systemNameTmp.value = systemName.value;
    timeoutTmp.value = timeout.value;
    weatherApiKeyTmp.value = weatherApiKey.value;
    // 获取密钥
    const aesKeyFromLocalStorage = localStorage.getItem('aesKey');
    if (aesKeyFromLocalStorage) {
      aesKey.value = aesKeyFromLocalStorage;
    }
    // 获取密钥base64
    const aesKeyBase64FromLocalStorage = localStorage.getItem('aesKeyBase64');
    if (aesKeyBase64FromLocalStorage) {
      aesKeyBase64.value = aesKeyBase64FromLocalStorage;
    }
  });

  const handleExceed = (files, fileList) => {
    // 如果文件超出了限制数量，移除旧的文件并添加新文件
    fileList.splice(0, fileList.length, files[files.length - 1]); // 保留最新的文件，删除旧文件
  };

  // 保存系统名称到服务器的逻辑
  const saveSystemName = async () => {
    const data = {
      data: { sysName: systemNameTmp.value }, // 设置的系统名称
      setType: 'systemName',   // 设置类型
    };
    // 使用 Axios 发送 POST 请求
    const response = await http.post("/setting", data, {
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (response.data.message == 'Success') {
      systemName.value = response.data.data;
      ElMessage({
        message: '系统名称保存成功',
        type: 'success',
        grouping: true,
      })
    } else {
      ElMessage({
        message: '系统名称保存失败',
        type: 'error',
        grouping: true,
      })
    }
  };

  const handleBeforeUpload = async (file) => {
    try {
      const reader = new FileReader();

      // 读取文件内容为文本
      reader.onload = (event) => {
        const fileContent = event.target.result;
        console.log("文件内容读取成功:", fileContent);

        // 使用正则表达式提取 AES 密钥
        const aesKeyMatch = fileContent.match(/-----BEGIN AES KEY-----([\s\S]*?)-----END AES KEY-----/);
        if (aesKeyMatch) {
          aesKeyBase64.value = aesKeyMatch[1].trim(); // 提取并去掉多余的空白字符
          const aesKeyDecoded = atob(aesKeyBase64.value);
          const decodedBuffer = new Uint8Array(aesKeyDecoded.length);
          for (let i = 0; i < aesKeyDecoded.length; i++) {
            decodedBuffer[i] = aesKeyDecoded.charCodeAt(i);
          }
          const aesKeyHex = Array.from(decodedBuffer).map(byte => byte.toString(16).padStart(2, '0')).join('');
          //console.log("aes密钥：", aesKeyHex);

          aesKey.value = aesKeyHex; // 更新到密钥输入框
          ElMessage({
            message: "密钥文件上传成功，AES 密钥已提取！",
            type: 'success',
            grouping: true,
          })
        } else {
          ElMessage({
            message: "未找到有效的 AES 密钥，请检查文件格式。",
            type: 'error',
            grouping: true,
          })
        }
      };

      // 错误处理
      reader.onerror = (event) => {
        console.error("文件读取失败:", event.target.error);
        alert("文件读取失败，请检查文件格式或文件内容！");
      };

      // 开始读取文件内容
      reader.readAsText(file);

      // 返回 false 防止自动上传
      return false;
    } catch (error) {
      console.error("处理文件时发生错误:", error);
      alert("处理文件失败，请稍后重试！");
      return false;
    }
  };

  // 安全设置
  const saveSecuritySettings = async () => {
    try {
      // 发送请求
      const response = await http.post("/setting", {
        data: { timeout: timeoutTmp.value, },
        setType: 'security',
      }, {
        headers: {
          'Content-Type': 'application/json'
        }
      });
      if (response.data.message == 'Success') {
        timeout.value = response.data.data;
        ElMessage({
          message: '安全设置已成功保存！',
          type: 'success',
          grouping: true,
        })
      } else {
        ElMessage({
          message: '安全设置保存失败',
          type: 'error',
          grouping: true,
        })
      }
    } catch (error) {
      ElMessage({
        message: '安全设置保存失败:', error,
        type: 'error',
        grouping: true,
      })
    }
  };

  // 保存密钥到web本地存储
  const saveKeyToWeb = () => {
    try {
      if (aesKey.value && aesKeyBase64.value) {
        localStorage.setItem('aesKey', aesKey.value);
        localStorage.setItem('aesKeyBase64', aesKeyBase64.value);
        ElMessage({
          message: '密钥已成功保存到前端！',
          type: 'success',
          grouping: true,
        })
      } else if (aesKeyBase64.value) {
        // 根据 aesKeyBase64 计算 aesKey
        const aesKeyDecoded = atob(aesKeyBase64.value);
        const decodedBuffer = new Uint8Array(aesKeyDecoded.length);
        for (let i = 0; i < aesKeyDecoded.length; i++) {
          decodedBuffer[i] = aesKeyDecoded.charCodeAt(i);
        }
        const aesKeyHex = Array.from(decodedBuffer).map(byte => byte.toString(16).padStart(2, '0')).join('');
        aesKey.value = aesKeyHex;
        localStorage.setItem('aesKey', aesKey.value);
        localStorage.setItem('aesKeyBase64', aesKeyBase64.value);
        ElMessage({
          message: '密钥已成功保存到前端！',
          type: 'success',
          grouping: true,
        })
      } else if (aesKey.value) {
        // 根据 aesKey 计算 aesKeyBase64
        const aesKeyBytes = aesKey.value.match(/.{1,2}/g).map(byte => parseInt(byte, 16));
        const aesKeyDecoded = String.fromCharCode(...aesKeyBytes);
        aesKeyBase64.value = btoa(aesKeyDecoded);
        localStorage.setItem('aesKey', aesKey.value);
        localStorage.setItem('aesKeyBase64', aesKeyBase64.value);
        ElMessage({
          message: '密钥已成功保存到前端！',
          type: 'success',
          grouping: true,
        })
      } else {
        ElMessage({
          message: '请先上传密钥文件或手动输入密钥！',
          type: 'error',
          grouping: true,
        })
        return;
      }
    } catch (error) {
      ElMessage({
        message: '保存密钥到前端时出错:', error,
        type: 'error',
        grouping: true,
      })
    }
  };

  // 保存密钥到web本地存储和服务端
  const saveKeyToCS = async () => {
    // 本地保存密钥
    saveKeyToWeb();

    // 准备请求数据
    const payload = {
      aesKey: aesKey.value,
      aesKeyBase64: aesKeyBase64.value
    };

    // 发送加密请求
    const response = await http.post("/saveKeyToCS", payload, {
      headers: {
        'Content-Type': 'application/json',
      }
    });

    // 处理2xx状态码）
    if (response.data.message === 'Success') {
      ElMessage({
        message: '密钥已成功同步到服务端！',
        type: 'success',
        grouping: true,
      })
    } else {
      ElMessage({
        message: '同步到服务端失败: ' + response.data.message,
        type: 'error',
        grouping: true,
      })
    }
  };

  // 清除密钥
  const clearKey = async () => {
    // 清除本地保存的密钥信息
    localStorage.removeItem('aesKey');
    localStorage.removeItem('aesKeyBase64');
    ElMessage({
      message: "本地密钥已清除！",
      type: 'success',
      grouping: true,
    })
    aesKey.value = '';
    aesKeyBase64.value = '';

    // 发送请求到服务器清除密钥
    const response = await http.get("/clearKey", {
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (response.data.message === 'Success') {
      ElMessage({
        message: response.data.data,
        type: 'success',
        grouping: true,
      })
    } else {
      ElMessage({
        message: `清除失败: ${response.data.message}`,
        type: 'error',
        grouping: true,
      })
    }
  };
</script>

<style scoped>
  /* 页面整体样式 */
  .settings-page {
    max-width: 1500px;
    margin: 10px auto;
    padding: 20px;
    background: var(--bg-container);
    border-radius: 12px;
    box-shadow: var(--shadow-sm);
    font-family: inherit;
    height: 100%;
    width: 100%;
    transition: all 0.3s ease;
  }

  /* 标题样式 */
  .header {
    text-align: center;
    margin-bottom: 30px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border-light);
  }

  .header h1 {
    font-size: 28px;
    color: var(--text-main);
    margin-bottom: 10px;
    font-weight: 600;
  }

  .header p {
    font-size: 16px;
    color: var(--text-secondary);
  }

  /* 设置区域样式 */
  .settings-sections {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(600px, 1fr));
    gap: 25px;
    max-height: calc(100vh - 220px);
    overflow-y: auto;
    padding-right: 10px;
  }

  /* 每个设置区域 - 分块卡片样式 */
  .setting-section {
    padding: 25px;
    background: var(--bg-container);
    border-radius: 12px;
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border-color-light);
    transition: all 0.3s ease;
  }

  .setting-section:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }

  .setting-section h2 {
    font-size: 20px;
    color: var(--text-main);
    margin-bottom: 20px;
    border-left: 4px solid var(--color-primary);
    padding-left: 15px;
    font-weight: 600;
  }

  /* 表单项单位样式 */
  .unit {
    margin-left: 10px;
    font-size: 14px;
    color: var(--text-secondary);
  }

  /* 按钮样式 */
  .saveBtn {
    text-align: center;
    margin-top: 25px;
  }

  .el-button {
    margin-right: 10px;
    transition: all 0.3s ease;
  }

  .el-button:hover {
    transform: translateY(-1px);
  }

  /* 文件上传区域样式 */
  .file-upload-area {
    width: 100%;
    align-items: center;
    padding: 5px;
    border: 2px dashed var(--color-primary);
    border-radius: 10px;
    background-color: var(--bg-container);
    cursor: pointer;
    text-align: center;
    margin-top: 20px;
    transition: all 0.3s ease;
  }

  .file-upload-area:hover {
    border-color: var(--color-primary-hover);
    background-color: var(--bg-muted);
  }

  /* Element Plus 组件样式覆盖 */
  .el-input__wrapper {
    border-color: var(--border-color-light);
    transition: border-color 0.3s ease;
  }

  .el-input__wrapper:focus-within {
    border-color: var(--color-primary) !important;
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
  }

  .el-input-number__decrease:hover:not(.is-disabled):not(.is-without-controls) {
    border-right-color: var(--primary-color);
  }

  .el-tabs__nav {
    padding-left: 10px;
  }

  .el-tabs__item {
    color: var(--text-secondary);
    font-size: 14px;
    transition: all 0.3s ease;
  }

  .el-tabs__item.is-active {
    color: var(--color-primary);
  }

  .el-tabs__active-bar {
    background-color: var(--color-primary);
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .settings-page {
      margin: 0;
      border-radius: 0;
      padding: 15px;
    }

    .header h1 {
      font-size: 24px;
    }

    .settings-sections {
      grid-template-columns: 1fr;
      padding-right: 0;
    }

    .setting-section {
      padding: 15px;
      margin-right: 0;
    }

    .setting-section h2 {
      font-size: 18px;
    }

    .el-form label {
      font-size: 14px;
    }

    .el-button {
      width: 100%;
      margin-right: 0;
      margin-bottom: 10px;
    }
  }

  @media (max-width: 480px) {
    .header h1 {
      font-size: 20px;
    }

    .header p {
      font-size: 14px;
    }

    .setting-section h2 {
      font-size: 16px;
    }
  }
</style>