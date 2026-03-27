<template>
  <div class="account-page">
    <!-- 页面标题 -->
    <div class="header">
      <h1>用户中心</h1>
      <p>管理您的账户信息和安全设置</p>
    </div>

    <!-- 用户信息卡片 -->
    <div class="user-info-card">
      <div class="avatar-section">
        <div class="avatar-container" @click.stop="uploadAvatar">
          <el-avatar :src="authStore.avatar || 'https://empty'" size="120" fit="cover" class="user-avatar">
            <i class="el-icon-user"></i>
          </el-avatar>
          <div class="avatar-upload-overlay">
            <i class="el-icon-camera"></i>
            <span>更换头像</span>
          </div>
        </div>
        <!-- 隐藏的文件上传input -->
        <input ref="fileInput" type="file" accept="image/jpeg, image/jpg, image/png, image/gif, image/webp"
          style="display: none" @change="handleFileSelect">
      </div>

      <div class="user-details">
        <div class="user-main-info">
          <h2>{{ authStore.username }}</h2>
        </div>

        <div class="user-meta-info">
          <div class="meta-item">
            <span class="meta-label">没想好放啥</span>
            <span class="meta-value status-active">xxx</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 设置区域 -->
    <div class="settings-container">
      <!-- 修改密码 -->
      <section class="setting-card">
        <div class="setting-header">
          <div class="setting-icon">
            <i class="el-icon-lock"></i>
          </div>
          <div class="setting-title-section">
            <h3>修改密码</h3>
            <p>定期更换密码以保障账户安全</p>
          </div>
        </div>

        <div class="setting-content">
          <el-form label-width="140px" class="form-layout">
            <el-form-item label="旧密码">
              <el-input v-model="passwords.oldPassword" type="password" placeholder="请输入旧密码"
                :show-password="true"></el-input>
            </el-form-item>
            <el-form-item label="新密码">
              <el-input v-model="passwords.newPassword" type="password" placeholder="请输入新密码"
                :show-password="true"></el-input>
            </el-form-item>
            <el-form-item label="确认密码">
              <el-input v-model="passwords.confirmPassword" type="password" placeholder="请再次输入新密码"
                :show-password="true"></el-input>
            </el-form-item>

            <div class="form-actions">
              <el-button type="primary" @click="updatePassword" class="btn-primary">
                <i class="el-icon-check"></i> 更新密码
              </el-button>
            </div>
          </el-form>
        </div>
      </section>

      <!-- 安全设置 -->
      <section class="setting-card">
        <div class="setting-header">
          <div class="setting-icon">
            <i class="el-icon-shield"></i>
          </div>
          <div class="setting-title-section">
            <h3>安全设置</h3>
            <p>增强您的账户安全性</p>
          </div>
        </div>

        <div class="setting-content">
          <div class="security-option">
            <div class="security-info">
              <h4>双重验证</h4>
              <p>启用双重验证以防止未经授权的访问</p>
            </div>
            <div class="security-action">
              <el-switch v-model="security.twoFactorAuth" active-color="#13ce66" inactive-color="#dcdfe6"
                @change="toggleTwoFactorAuth" />
              <el-button v-if="security.twoFactorAuth" type="success" size="small" @click="openTotpDialog" class="ml-4">
                管理 TOTP
              </el-button>
            </div>
          </div>

          <div class="security-option">
            <div class="security-info">
              <h4>登录提醒</h4>
              <p>登录时接收邮件通知</p>
            </div>
            <div class="security-action">
              <el-switch v-model="security.loginAlert" active-color="#13ce66" inactive-color="#dcdfe6"
                @change="toggleLoginAlert" />
              <el-button v-if="security.loginAlert" type="success" size="small" @click="openLoginAlertDialog"
                class="ml-4">
                管理设置
              </el-button>
            </div>
          </div>

          <div class="security-option">
            <div class="security-info">
              <h4>白名单设置</h4>
              <p>设置受信任的IP地址或网段</p>
            </div>
            <div class="security-action">
              <el-switch v-model="security.ipWhitelist" active-color="#13ce66" inactive-color="#dcdfe6"
                @change="toggleIpWhitelist" />
              <el-button v-if="security.ipWhitelist" type="success" size="small" @click="openWhitelistDialog"
                class="ml-4">
                管理白名单
              </el-button>
            </div>
          </div>

          <div class="security-option" style="border-bottom: none;">
            <div class="security-info">
              <h4>更换主密钥</h4>
              <p>更换您的账户主加密密钥</p>
            </div>
            <div class="security-action full-width">
              <el-input v-model="masterKey.username" type="text" placeholder="请输入用户名" :show-password="false"
                style="margin-bottom: 12px;"></el-input>
              <el-input v-model="masterKey.password" type="password" placeholder="请输入密码" :show-password="true"
                style="margin-bottom: 12px;"></el-input>
              <el-input v-model="masterKey.oldMasterKey" type="password" placeholder="请输入原主密钥" :show-password="true"
                style="margin-bottom: 12px;"></el-input>
              <el-button type="primary" size="small" @click="changeMasterKey">
                更换主密钥
              </el-button>
              <p class="security-note">注意：更换主密钥后，所有加密数据将使用新密钥重新加密</p>
            </div>
          </div>
        </div>
      </section>


    </div>
  </div>

  <!-- TOTP 设置对话框 -->
  <el-dialog v-model="totpDialogVisible" title="设置双重验证" width="500px" :before-close="handleTotpDialogClose">
    <div v-if="!totpSetupSuccess" class="totp-setup-form">
      <el-form label-width="120px" class="form-layout">
        <el-form-item label="密钥">
          <el-input v-model="totpSettings.secretKey" placeholder="请输入32位密钥"
            style="width: calc(100% - 100px); display: inline-block;"></el-input>
          <el-button type="default" size="small" @click="generateSecretKey" style="margin-left: 10px;">生成</el-button>
        </el-form-item>
        <el-form-item label="刷新周期">
          <el-input v-model.number="totpSettings.refreshPeriod" type="number" placeholder="请输入刷新周期(秒)" min="15"
            max="300"></el-input>
        </el-form-item>
        <el-form-item label="应用名">
          <el-input v-model="totpSettings.appName" placeholder="请输入应用名称"></el-input>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="totpSettings.username" placeholder="请输入用户名"></el-input>
        </el-form-item>
        <el-form-item label="位数">
          <el-select v-model="totpSettings.digits" placeholder="请选择验证码位数">
            <el-option label="6位" :value="6"></el-option>
            <el-option label="8位" :value="8"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <div class="form-tips">
        <p><i class="el-icon-info"></i> 提示：请妥善保管您的密钥，丢失后无法恢复</p>
      </div>
    </div>

    <div v-else class="totp-qr-code-section">
      <div class="qr-code-container">
        <img v-if="totpQrCode" :src="totpQrCode" alt="TOTP 二维码" class="qr-code-image">
        <div v-else class="qr-code-loading">
          <i class="el-icon-loading el-icon--loading"></i>
          <span>加载二维码中...</span>
        </div>
      </div>
      <div class="totp-info">
        <h4>双重验证已设置成功</h4>
        <p>请使用认证应用扫描上方二维码，或手动输入以下密钥：</p>
        <div class="secret-key-display">
          <code>{{ totpSettings.secretKey }}</code>
        </div>
        <p class="important-note">重要：请妥善保存此密钥，它将用于账户恢复</p>
      </div>
    </div>

    <template #footer>
      <div v-if="!totpSetupSuccess">
        <el-button @click="handleTotpDialogClose">取消</el-button>
        <el-button type="primary" @click="submitTotpSettings(true)">
          确认设置
        </el-button>
      </div>
      <div v-else>
        <el-button type="primary" @click="totpDialogVisible = false">
          完成
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 登录提醒设置对话框 -->
  <el-dialog v-model="loginAlertDialogVisible" title="登录提醒设置" width="500px" :before-close="handleLoginAlertDialogClose">
    <div class="login-alert-setup-form">
      <el-form label-width="120px" class="form-layout">
        <el-form-item label="发件人邮箱">
          <el-input v-model="loginAlertSettings.from" placeholder="请输入发件人邮箱"></el-input>
        </el-form-item>
        <el-form-item label="发件人昵称">
          <el-input v-model="loginAlertSettings.nickname" placeholder="请输入发件人昵称"></el-input>
        </el-form-item>
        <el-form-item label="授权码">
          <el-input v-model="loginAlertSettings.secret" type="password" placeholder="请输入SMTP授权码"
            :show-password="true"></el-input>
        </el-form-item>
        <el-form-item label="SMTP服务器">
          <el-input v-model="loginAlertSettings.host" placeholder="请输入SMTP服务器地址"></el-input>
        </el-form-item>
        <el-form-item label="SMTP端口">
          <el-input v-model.number="loginAlertSettings.port" type="number" placeholder="请输入SMTP端口号"></el-input>
        </el-form-item>
        <el-form-item label="是否使用SSL">
          <el-switch v-model="loginAlertSettings.isSSL" active-color="#13ce66" inactive-color="#dcdfe6"></el-switch>
        </el-form-item>
        <el-form-item label="接收邮箱">
          <el-input v-model="loginAlertSettings.to" placeholder="请输入接收提醒的邮箱"></el-input>
        </el-form-item>
        <el-form-item label="加密授权码">
          <el-switch v-model="loginAlertSettings.encryptSecret" active-color="#13ce66"
            inactive-color="#dcdfe6"></el-switch>
        </el-form-item>


      </el-form>
      <div class="form-tips">
        <p><i class="el-icon-info"></i> 提示：QQ/163/Gmail等邮箱需要使用授权码而不是登录密码</p>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleLoginAlertDialogClose">取消</el-button>
      <el-button type="primary" @click="submitLoginAlertSettings">
        保存设置
      </el-button>
    </template>
  </el-dialog>

  <!-- 白名单设置对话框 -->
  <el-dialog v-model="whitelistDialogVisible" title="IP白名单设置" width="500px" :before-close="handleWhitelistDialogClose">
    <div class="whitelist-setup-form">
      <!-- 白名单IP/网段设置 -->
      <el-form label-width="120px" class="form-layout">
        <el-form-item label="白名单列表">
          <div class="whitelist-container">
            <div v-for="(item, index) in whitelistSettings.whitelist" :key="index" class="whitelist-item">
              <el-input v-model="whitelistSettings.whitelist[index]"
                placeholder="请输入IP地址或网段（如：192.168.1.1 或 192.168.1.0/24）"></el-input>
              <el-button @click="removeWhitelistItem(index)" type="danger" size="small"
                icon="el-icon-delete"></el-button>
            </div>
            <el-button v-if="whitelistSettings.whitelist.length < 10" @click="addWhitelistItem" type="primary"
              size="small" icon="el-icon-plus">
              添加白名单
            </el-button>
            <div v-else class="whitelist-limit-tip">
              <i class="el-icon-info"></i> 白名单最多支持10条记录
            </div>
          </div>
        </el-form-item>

        <!-- 白名单之外动作选择 -->
        <el-form-item label="白名单之外动作">
          <el-radio-group v-model="whitelistSettings.actionOutsideWhitelist">
            <el-radio label="alert">允许并邮件提醒</el-radio>
            <el-radio label="block">直接拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <div class="form-tips">
        <p><i class="el-icon-info"></i> 白名单说明：白名单内的IP或网段登录时不会发送邮件提醒，最多可添加10条记录</p>
        <p><i class="el-icon-info"></i> 白名单之外动作：设置非白名单IP尝试登录时的处理方式</p>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleWhitelistDialogClose">取消</el-button>
      <el-button type="primary" @click="submitWhitelistSettings">
        保存设置
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { reactive, ref, onMounted } from "vue";
  import keyExchange from '@/js/keyExchange.js'
  import http from '@/js/http.js'
  import { aesGcmEncrypt } from '@/js/aesCryption.js'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAuthStore } from '@/store/auth'

  // 文件输入框引用
  const fileInput = ref(null);
  const authStore = useAuthStore();

  // 组件挂载时初始化用户数据
  onMounted(() => {
    // 从authStore获取用户名
    localUser.name = authStore.username || 'admin';

    // 立即从服务器获取最新用户信息和安全设置
    initUserInfo();
    initSecuritySettings();
    getLoginAlertSettings();
    getWhitelistSettings();
  });

  // 初始化用户信息
  const initUserInfo = async () => {
    try {
      const response = await http.get('/getAvatarAndUser');
      if (response.status === 200 && response.data.data) {
        const data = response.data.data;

        // 更新用户名
        if (data.username) {
          localUser.name = data.username;
          if (!authStore.username) {
            authStore.username = data.username;
            sessionStorage.setItem('username', data.username);
          }
        }

        // 更新头像
        if (data.avatar) {
          const avatarData = data.avatar.startsWith('data:image/')
            ? data.avatar
            : `data:image/png;base64,${data.avatar}`;
          authStore.updateAvatar(avatarData);
        }

        // 其他用户信息已简化，无需更新
      } else if (response.status === 404) {
        ElMessage({
          message: '用户数据404！',
          type: 'warning'
        });
      }
    } catch (error) {
      console.error('初始化用户信息失败:', error);
      // 出错时不影响页面显示，继续使用现有数据
    }
  };

  // 初始化安全设置
  const initSecuritySettings = async () => {
    try {
      // 调用后端API获取安全设置信息
      const response = await http.get('/getSecuritySettings?mod=totp');

      // 统一格式处理：{code, message, data}
      if (response.status === 200 && response.data.code === 200) {
        const data = response.data.data;

        // 更新安全设置
        if (data) {
          // 处理TOTP设置数据
          if (data.totpset && typeof data.totpset === 'object') {
            const totpData = data.totpset;

            // 更新双重验证状态
            if (typeof totpData.enabled === 'boolean') {
              security.twoFactorAuth = totpData.enabled;
            }

            // 更新TOTP设置信息
            if (totpData.secretKey && typeof totpData.secretKey === 'string') {
              totpSettings.secretKey = totpData.secretKey;
            }

            if (totpData.refreshPeriod && typeof totpData.refreshPeriod === 'number') {
              totpSettings.refreshPeriod = totpData.refreshPeriod;
            }

            if (totpData.digits && typeof totpData.digits === 'number') {
              totpSettings.digits = totpData.digits;
            }

            if (totpData.appName && typeof totpData.appName === 'string') {
              totpSettings.appName = totpData.appName;
            }

            if (totpData.username && typeof totpData.username === 'string') {
              totpSettings.username = totpData.username;
            }

            // 记录设置状态
            if (typeof totpData.setupStatus === 'boolean') {
              totpSetupSuccess.value = totpData.setupStatus;
            }
          }

          // 登录提醒状态
          if (typeof data.loginAlert === 'boolean') {
            security.loginAlert = data.loginAlert;
          }

          // 扩展其他安全设置
        }
      } else {
        console.error('获取安全设置失败:', response.data.message || '未知错误');
      }
    } catch (error) {
      console.error('获取安全设置时发生错误:', error);
      // 出错时不影响页面显示，保持默认设置
    }
  };

  // 本地用户数据
  const localUser = reactive({
    name: "admin"
  });

  const passwords = reactive({
    oldPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  // 安全设置
  const security = reactive({
    twoFactorAuth: false,
    loginAlert: false,
    ipWhitelist: false
  });

  // 登录提醒设置相关数据
  const loginAlertSettings = reactive({
    from: '',
    nickname: '',
    secret: '',
    host: '',
    port: 465,
    isSSL: true,
    encryptSecret: false,
    to: ''
  });

  // 白名单设置相关数据
  const whitelistSettings = reactive({
    whitelist: [], // 白名单IP/网段列表，最多10条
    actionOutsideWhitelist: 'alert' // 'alert': 允许并邮件提醒, 'block': 直接拒绝
  });

  // 白名单对话框控制
  const whitelistDialogVisible = ref(false);

  // 登录提醒对话框控制
  const loginAlertDialogVisible = ref(false);

  const masterKey = reactive({
    oldMasterKey: "",
    username: localUser.name,
    password: ""
  });

  // 切换登录提醒状态
  const toggleLoginAlert = async () => {
    if (security.loginAlert) {
      // 当开启登录提醒时，打开设置对话框
      openLoginAlertDialog();
    } else {
      // 关闭登录提醒，发送关闭请求
      try {
        await submitLoginAlertSettings(false);
      } catch (error) {
        console.error("关闭登录提醒失败:", error);
        ElMessage({
          message: "关闭登录提醒失败，请稍后重试",
          type: 'error',
          grouping: true,
        });
        // 恢复状态
        security.loginAlert = true;
      }
    }
  };

  // 切换白名单状态
  const toggleIpWhitelist = async () => {
    if (security.ipWhitelist) {
      // 当开启白名单时，打开设置对话框
      openWhitelistDialog();
    } else {
      // 关闭白名单，发送关闭请求
      try {
        await submitWhitelistSettings(false);
      } catch (error) {
        console.error("关闭白名单失败:", error);
        ElMessage({
          message: "关闭白名单失败，请稍后重试",
          type: 'error',
          grouping: true,
        });
        // 恢复状态
        security.ipWhitelist = true;
      }
    }
  };

  // 打开登录提醒设置对话框
  const openLoginAlertDialog = () => {
    // 从服务器获取当前设置（如果有）
    //getLoginAlertSettings();
    loginAlertDialogVisible.value = true;
  };

  // 打开白名单设置对话框
  const openWhitelistDialog = () => {
    // 从服务器获取当前设置（如果有）
    //getWhitelistSettings();
    whitelistDialogVisible.value = true;
  };

  // 获取登录提醒设置
  const getLoginAlertSettings = async () => {
    try {
      const response = await http.get('/getSecuritySettings?mod=mailAlert');
      if (response.status === 200 && response.data.code === 200) {
        const data = response.data.data;
        if (data) {
          // 更新登录提醒设置
          security.loginAlert = data.enabled;
          loginAlertSettings.from = data.from;
          loginAlertSettings.nickname = data.nickname;
          loginAlertSettings.secret = data.secret;
          loginAlertSettings.host = data.host;
          loginAlertSettings.port = data.port;
          loginAlertSettings.isSSL = data.isSSL;
          loginAlertSettings.to = data.to;
          loginAlertSettings.encryptSecret = data.encryptSecret;
        }
      }
    } catch (error) {
      console.error("获取登录提醒设置失败:", error);
    }
  };

  // 获取白名单设置
  const getWhitelistSettings = async () => {
    try {
      const response = await http.get('/getSecuritySettings?mod=ipWhitelist');
      if (response.status === 200 && response.data.code === 200) {
        const data = response.data.data;
        if (data) {
          // 更新白名单设置
          security.ipWhitelist = data.enabled;
          whitelistSettings.whitelist = data.whitelist || [];
          whitelistSettings.actionOutsideWhitelist = data.actionOutsideWhitelist || 'alert';
        }
      }
    } catch (error) {
      console.error("获取白名单设置失败:", error);
    }
  };

  // 提交登录提醒设置
  const submitLoginAlertSettings = async (isEnable = true) => {
    try {
      // 获取通信密钥
      const tmpAesKey = await keyExchange.getAesKey();


      const alertData = {
        from: isEnable ? loginAlertSettings.from : '',
        nickname: isEnable ? loginAlertSettings.nickname : '',
        secret: isEnable ? loginAlertSettings.secret : '',
        host: isEnable ? loginAlertSettings.host : '',
        port: isEnable ? loginAlertSettings.port : 465,
        isSSL: !!isEnable && !!loginAlertSettings.isSSL,
        to: isEnable ? loginAlertSettings.to : '',
        encryptSecret: !!isEnable && !!loginAlertSettings.encryptSecret,
        enabled: !!isEnable
      };

      // 加密整个数据对象
      const encryptedData = await keyExchange.encryptData(JSON.stringify(alertData), tmpAesKey);

      // 发送数据到后端API
      const response = await http.post('/setLoginAlertSettings', encryptedData, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // 处理响应
      const data = response.data;
      if (data.code === 200 && (data.message === 'Success' || data.message === '登录提醒设置成功')) {
        ElMessage({
          message: isEnable ? "登录提醒设置成功" : "登录提醒已关闭",
          type: 'success',
          grouping: true,
        });
        loginAlertDialogVisible.value = false;
        if (!isEnable) {
          loginAlertDialogVisible.value = false;
        }
      } else {
        ElMessage({
          message: data.message || '操作失败',
          type: 'error',
          grouping: true,
        });
        // 如果操作失败，恢复登录提醒开关状态
        security.loginAlert = isEnable;
      }
    } catch (error) {
      console.error("设置登录提醒失败:", error);
      ElMessage({
        message: "设置登录提醒失败，请稍后重试",
        type: 'error',
        grouping: true,
      });
      // 如果设置失败，恢复状态
      security.loginAlert = false;
    }
  };

  // 处理登录提醒对话框关闭
  const handleLoginAlertDialogClose = () => {
    // 如果用户关闭对话框且登录提醒是开启状态，则要求用户确认
    if (security.loginAlert) {
      ElMessageBox.confirm(
        '您确定要关闭设置窗口吗？未保存的设置将丢失。',
        '确认关闭',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(() => {
        loginAlertDialogVisible.value = false;
        getLoginAlertSettings();
      }).catch(() => {
        // 用户取消关闭，什么都不做
      });
    } else {
      loginAlertDialogVisible.value = false;
    }
  };

  // 添加白名单项
  const addWhitelistItem = () => {
    if (whitelistSettings.whitelist.length < 10) {
      whitelistSettings.whitelist.push('');
    }
  };

  // 移除白名单项
  const removeWhitelistItem = (index) => {
    whitelistSettings.whitelist.splice(index, 1);
  };

  // 提交白名单设置
  const submitWhitelistSettings = async (isEnable = true) => {
    try {
      // 验证必填字段
      if (isEnable) {
        // 过滤掉白名单中的空值
        const filteredWhitelist = whitelistSettings.whitelist.filter(item => item && item.trim() !== '');
        if (filteredWhitelist.length === 0) {
          ElMessage({
            message: "请至少添加一个白名单IP或网段",
            type: 'warning',
            grouping: true,
          });
          return;
        }
      }
      // 获取通信密钥
      const tmpAesKey = await keyExchange.getAesKey();
      // 组装数据
      const whitelistData = {
        enabled: !!isEnable,
        whitelist: whitelistSettings.whitelist.filter(item => item && item.trim() !== ''),
        actionOutsideWhitelist: whitelistSettings.actionOutsideWhitelist
      };

      // 加密数据
      const encryptedData = await keyExchange.encryptData(JSON.stringify(whitelistData), tmpAesKey);

      // 发送数据到后端API
      const response = await http.post('/setWhitelistSettings', encryptedData, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // 处理响应
      const data = response.data;
      if (data.code === 200 && (data.message === 'Success' || data.message === '白名单设置成功')) {
        ElMessage({
          message: isEnable ? "白名单设置成功" : "白名单已关闭",
          type: 'success',
          grouping: true,
        });
        whitelistDialogVisible.value = false;
      } else {
        ElMessage({
          message: data.message || '操作失败',
          type: 'error',
          grouping: true,
        });
        // 如果操作失败，恢复白名单开关状态
        security.ipWhitelist = isEnable;
      }
    } catch (error) {
      console.error("设置白名单失败:", error);
      ElMessage({
        message: "设置白名单失败，请稍后重试",
        type: 'error',
        grouping: true,
      });
      // 如果设置失败，恢复状态
      security.ipWhitelist = false;
    }
  };

  // 处理白名单对话框关闭
  const handleWhitelistDialogClose = () => {
    // 如果用户关闭对话框且白名单是开启状态，则要求用户确认
    if (security.ipWhitelist) {
      ElMessageBox.confirm(
        '您确定要关闭设置窗口吗？未保存的设置将丢失。',
        '确认关闭',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(() => {
        whitelistDialogVisible.value = false;
        getWhitelistSettings();
      }).catch(() => {
        // 用户取消关闭，什么都不做
      });
    } else {
      whitelistDialogVisible.value = false;
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
  }

  const changeMasterKey = async () => {
    if (!masterKey.username || !masterKey.password || !masterKey.oldMasterKey) {
      ElMessage({
        message: "请填写用户名、密码和原主密钥",
        type: 'warning',
        grouping: true,
      });
      return;
    }

    try {
      // 获取通信密钥
      const tmpAesKey = await keyExchange.getAesKey();
      // 组装数据，包含用户名、密码和原主密钥用于身份核验
      const changeMasterKeyData = {
        username: masterKey.username,
        password: masterKey.password,
        oldMasterKey: masterKey.oldMasterKey
      };
      // 加密数据
      const encryptedData = await keyExchange.encryptData(JSON.stringify(changeMasterKeyData), tmpAesKey);
      // 发送数据到后端专门的API
      const response = await http.post(`/changeMasterKey`, encryptedData, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // 处理后端返回的响应，参考注册实现的处理方式
      const data = response.data;
      // 获取data字段值
      const result = data.data;

      // 解码和解密返回的数据
      if (result && result.iv && result.encryptedPEM) {
        const iv = Uint8Array.from(atob(result.iv), c => c.charCodeAt(0));
        const encryptedPEM = Uint8Array.from(atob(result.encryptedPEM), c => c.charCodeAt(0));

        try {
          // 解密数据
          const decryptedPEM = await keyExchange.decryptAES(encryptedPEM, tmpAesKey, iv);
          console.log('解密后的 PEM 数据:', decryptedPEM);

          // 自动下载 PEM 文件
          downloadPEM(decryptedPEM);
        } catch (decryptError) {
          console.error('解密数据失败:', decryptError);
        }
      }

      // 处理主密钥更换状态
      if (data.code === 200 && data.message === 'Success') {
        ElMessage({
          message: "主密钥更换成功！",
          type: 'success',
          grouping: true,
        });
        // 清空输入框
        masterKey.oldMasterKey = "";
        masterKey.password = "";
      } else {
        ElMessage({
          message: data.message || '更换主密钥失败',
          type: 'error',
          grouping: true,
        });
      }
    } catch (error) {
      console.error("更换主密钥时发生错误:", error);
      ElMessage({
        message: "更换主密钥失败，请稍后重试",
        type: 'error',
        grouping: true,
      });
    }
  };

  // 触发文件选择框
  const uploadAvatar = () => {
    // 使用Vue的ref来触发文件选择
    if (fileInput.value) {
      fileInput.value.click();
    } else {
      // 备选方案，确保能正常工作
      const input = document.querySelector('input[type="file"]');
      if (input) {
        input.click();
      } else {
        console.error('未找到文件上传输入框');
      }
    }
  };

  // 处理文件选择
  const handleFileSelect = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    // 验证文件类型
    const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    if (!validTypes.includes(file.type)) {
      ElMessage({
        message: '请上传有效的图片文件(JPG、PNG、GIF、WebP)',
        type: 'error',
        grouping: true,
      });
      // 清空文件输入，允许重新选择
      event.target.value = '';
      return;
    }

    // 验证文件大小（限制为5MB）
    const maxSize = 5 * 1024 * 1024; // 5MB
    if (file.size > maxSize) {
      ElMessage({
        message: '图片大小不能超过5MB',
        type: 'error',
        grouping: true,
      });
      // 清空文件输入，允许重新选择
      event.target.value = '';
      return;
    }

    try {
      // 读取文件二进制内容
      const arrayBuffer = await readFileAsArrayBuffer(file);

      // 将ArrayBuffer转换为Base64字符串，用于JSON传输
      const base64Data = arrayBufferToBase64(arrayBuffer);

      // 构造请求数据对象，包含filename、size和data
      const requestData = {
        filename: file.name,
        size: file.size,
        data: base64Data
      };

      // 调用后端updateAvatar接口，以JSON格式发送数据
      const response = await http.post('/updateAvatar', requestData, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // 处理响应
      if (response.status === 200 && response.data.code === 200) {
        // 根据服务端返回格式，直接显示成功消息
        ElMessage({
          message: response.data.message || '头像更新成功',
          type: 'success',
          grouping: true,
        });

        // 调用接口获取最新的头像数据，以便同步更新Manager.vue中的头像显示
        try {
          const avatarResponse = await http.get('/getAvatarAndUser');
          if (avatarResponse.status === 200 && avatarResponse.data.data && avatarResponse.data.data.avatar) {
            // 处理头像数据，转换为Data URL格式（与Manager.vue中的处理保持一致）
            const avatarData = avatarResponse.data.data.avatar.startsWith('data:image/')
              ? avatarResponse.data.data.avatar
              : `data:image/png;base64,${avatarResponse.data.data.avatar}`;

            // 使用store更新头像，实现跨组件同步
            authStore.updateAvatar(avatarData);
          }
        } catch (avatarError) {
          console.error('获取最新头像失败:', avatarError);
        }

        // 清空文件输入
        event.target.value = '';
      } else {
        ElMessage({
          message: response.data.message || '头像更新失败',
          type: 'error',
          grouping: true,
        });
      }
    } catch (error) {
      console.error('头像上传失败:', error);
      ElMessage({
        message: '头像上传失败，请稍后重试',
        type: 'error',
        grouping: true,
      });
    }
  };

  // 将ArrayBuffer转换为Base64字符串
  const arrayBufferToBase64 = (buffer) => {
    let binary = '';
    const bytes = new Uint8Array(buffer);
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
  };

  // 将文件读取为ArrayBuffer
  const readFileAsArrayBuffer = (file) => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        resolve(reader.result);
      };
      reader.onerror = () => {
        reject(new Error('文件读取失败'));
      };
      reader.readAsArrayBuffer(file);
    });
  };




  const updatePassword = async () => {
    if (passwords.newPassword !== passwords.confirmPassword) {
      alert("两次密码输入不一致");
      return;
    }
    // 获取通信密钥
    const tmpAesKey = await keyExchange.getAesKey();
    // 组装数据
    const updatePasswordData = {
      username: localUser.name,
      oldPassword: passwords.oldPassword,
      newPassword: passwords.newPassword,
    };
    // 加密数据
    const encryptedData = await keyExchange.encryptData(JSON.stringify(updatePasswordData), tmpAesKey);
    // 发送数据
    const response = await http.post(`/updatePassword`, encryptedData, {
      headers: {
        'Content-Type': 'application/json'
      }
    });
    // 处理成功响应
    const data = response.data;
    if (data.message === 'Success') {
      ElMessage({
        message: "密码更新成功！",
        type: 'success',
        grouping: true,
      });
    } else {
      ElMessage({
        message: data.data || '未知错误',
        type: 'error',
        grouping: true,
      });
    }
  };

  // TOTP设置相关数据
  const totpSettings = reactive({
    secretKey: '',
    refreshPeriod: 30,
    appName: 'PasswordManager',
    username: localUser.name,
    digits: 6
  });

  // TOTP对话框控制
  const totpDialogVisible = ref(false);
  const totpQrCode = ref('');
  const totpSetupSuccess = ref(false);

  // 打开TOTP设置对话框
  const openTotpDialog = () => {
    totpSetupSuccess.value = false;
    totpQrCode.value = '';
    totpSettings.username = localUser.name;
    totpDialogVisible.value = true;
  };

  // 切换双重验证状态
  const toggleTwoFactorAuth = async () => {
    if (security.twoFactorAuth) {
      // 打开时，发送请求获取totp设置状态
      try {
        // 打开时，发送POST请求获取totp设置状态，并附带布尔类型的启用标识
        const response = await http.post('/enableTotp', {
          enabled: true
        });
        if (response.status === 200 && response.data.code === 200) {
          // 处理响应数据，更新totpSettings状态
          totpSettings.secretKey = response.data.data.secretKey;
          totpSettings.refreshPeriod = response.data.data.refreshPeriod;
          totpSettings.appName = response.data.data.appName;
          totpSettings.username = response.data.data.username;
          totpSettings.digits = response.data.data.digits;
          totpSetupSuccess.value = response.data.data.enabled;
          if (!response.data.data.setupStatus) {
            ElMessage({
              message: "TOTP未设置！",
              type: 'warning',
              grouping: true,
            });
            // totp未设置，打开对话框
            totpDialogVisible.value = true;
            totpSetupSuccess.value = false;

          } else {
            ElMessage({
              message: "TOTP已设置！",
              type: 'success',
              grouping: true,
            });
          }
        } else {
          security.twoFactorAuth = false;
          ElMessage({
            message: response.data.message || '获取TOTP状态失败',
            type: 'error',
            grouping: true,
          });
        }
      } catch (error) {
        console.error("获取TOTP状态失败:", error);
        ElMessage({
          message: "获取TOTP状态失败，请稍后重试",
          type: 'error',
          grouping: true,
        });
      }
    } else {
      // 关闭双重验证，发送关闭请求
      try {
        // 调用submitTotpSettings函数并传入isEnable=false
        await submitTotpSettings(false);
        // 不再需要重复显示成功消息，submitTotpSettings内部已处理
      } catch (error) {
        console.error("关闭双重验证失败:", error);
        ElMessage({
          message: "关闭双重验证失败，请稍后重试",
          type: 'error',
          grouping: true,
        });
        // 恢复状态
        security.twoFactorAuth = true;
      }
    }
  };

  // 提交TOTP设置
  // 添加类型检查确保isEnable是布尔值
  const submitTotpSettings = async (isEnable = true) => {
    // 确保isEnable是布尔值，如果传入的是事件对象则使用默认值true
    if (typeof isEnable !== 'boolean') {
      isEnable = true;
    }
    try {
      // 获取通信密钥
      const tmpAesKey = await keyExchange.getAesKey();
      // 组装数据
      const totpData = {
        secretKey: totpSettings.secretKey,
        refreshPeriod: totpSettings.refreshPeriod,
        appName: totpSettings.appName,
        username: totpSettings.username,
        digits: totpSettings.digits,
        enabled: isEnable
      };
      // 打印提交的数据
      console.log('提交的TOTP数据:', totpData);
      // 加密数据
      const encryptedData = await keyExchange.encryptData(JSON.stringify(totpData), tmpAesKey);
      // 发送数据到后端API
      const response = await http.post('/setTotp', encryptedData, {
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // 处理响应
      const data = response.data;
      if (isEnable && data.code === 200) {
        if (data.message === 'Success' || data.message === 'TOTP设置成功') {
          // 开启双重验证时，处理返回的base64编码二进制图像数据
          if (data.data && data.data.qrCode) {
            // 确保base64数据包含正确的data URL前缀
            if (data.data.qrCode.startsWith('data:image/')) {
              // 如果已经包含前缀，则直接使用
              totpQrCode.value = data.data.qrCode;
            } else {
              // 否则添加png格式的data URL前缀
              totpQrCode.value = `data:image/png;base64,${data.data.qrCode}`;
            }
            totpSetupSuccess.value = true;
            ElMessage({
              message: "双重验证设置成功",
              type: 'success',
              grouping: true,
            });
          } else {
            ElMessage({
              message: '设置双重验证失败：未返回二维码数据',
              type: 'error',
              grouping: true,
            });
            security.twoFactorAuth = false;
          }
        } else {
          // 关闭双重验证时，直接成功
          ElMessage({
            message: data.message || '数据缺失！',
            type: 'error',
            grouping: true,
          });
          totpDialogVisible.value = false;
        }
      } else {
        ElMessage({
          message: data.message || '操作失败',
          type: 'error',
          grouping: true,
        });
        // 如果操作失败，恢复双重验证开关状态
        security.twoFactorAuth = isEnable;
      }
    } catch (error) {
      console.error("设置双重验证失败:", error);
      ElMessage({
        message: "设置双重验证失败，请稍后重试",
        type: 'error',
        grouping: true,
      });
      // 如果设置失败，关闭双重验证开关
      //security.twoFactorAuth = false;
    }
  };

  // 生成随机密钥
  const generateSecretKey = () => {
    // 生成32位随机密钥（Base32格式）
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
    let key = '';
    for (let i = 0; i < 32; i++) {
      key += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    totpSettings.secretKey = key;
  };

  // 处理TOTP对话框关闭
  const handleTotpDialogClose = async () => {
    // 重置表单状态
    totpDialogVisible.value = false;
  };


</script>

<style scoped>
  /* 页面整体样式 */
  .account-page {
    max-width: 1000px;
    margin: 30px auto;
    padding: 0 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    max-height: calc(100vh - 100px);
    overflow-y: auto;
  }

  /* 滚动条样式美化 */
  .account-page::-webkit-scrollbar {
    width: 6px;
  }

  .account-page::-webkit-scrollbar-track {
    background: var(--bg-muted);
    border-radius: 3px;
  }

  .account-page::-webkit-scrollbar-thumb {
    background: var(--border-color-light);
    border-radius: 3px;
  }

  .account-page::-webkit-scrollbar-thumb:hover {
    background: var(--text-secondary);
  }

  /* 标题样式 */
  .header {
    text-align: center;
    margin-bottom: 40px;
    padding-top: 20px;
  }

  .header h1 {
    font-size: 32px;
    font-weight: 600;
    color: var(--text-main);
    margin-bottom: 8px;
    letter-spacing: -0.5px;
  }

  .header p {
    font-size: 16px;
    color: var(--text-secondary);
    margin: 0;
  }

  /* 用户信息卡片 */
  .user-info-card {
    display: flex;
    align-items: center;
    background-color: var(--bg-container);
    border-radius: 16px;
    padding: 32px;
    margin-bottom: 32px;
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border-color-light);
    transition: all 0.3s ease;
  }

  .user-info-card:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }

  /* TOTP 设置对话框样式 */
  .totp-setup-form {
    padding: 10px 0;
  }

  .form-tips {
    margin-top: 15px;
    padding: 10px;
    background-color: var(--bg-muted);
    border-radius: 6px;
  }

  .form-tips p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 14px;
  }

  .totp-qr-code-section {
    text-align: center;
    padding: 20px 0;
  }

  .qr-code-container {
    margin: 20px auto;
    padding: 15px;
    background-color: white;
    border-radius: 8px;
    display: inline-block;
  }

  .qr-code-image {
    max-width: 200px;
    max-height: 200px;
  }

  .qr-code-loading {
    padding: 40px;
    color: var(--text-secondary);
  }

  .totp-info {
    margin-top: 20px;
    text-align: left;
  }

  .totp-info h4 {
    margin-bottom: 10px;
    color: var(--text-main);
  }

  .totp-info p {
    margin: 10px 0;
    color: var(--text-secondary);
  }

  .secret-key-display {
    background-color: var(--bg-muted);
    padding: 10px;
    border-radius: 6px;
    word-break: break-all;
  }

  .secret-key-display code {
    font-family: 'Courier New', Courier, monospace;
    color: var(--color-primary);
  }

  .important-note {
    color: var(--color-warning) !important;
    font-weight: 500;
  }

  .avatar-section {
    flex: 0 0 auto;
    margin-right: 30px;
  }

  .avatar-container {
    position: relative;
    cursor: pointer;
  }

  .user-avatar {
    border: 4px solid rgba(255, 255, 255, 0.8);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }

  .user-avatar:hover {
    transform: scale(1.05);
  }

  .avatar-upload-overlay {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    background: rgba(0, 0, 0, 0.6);
    color: white;
    text-align: center;
    padding: 8px 0;
    border-radius: 0 0 12px 12px;
    opacity: 0;
    transition: opacity 0.3s ease;
    font-size: 12px;
  }

  .avatar-container:hover .avatar-upload-overlay {
    opacity: 1;
  }

  .user-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  .user-main-info {
    margin-bottom: 16px;
  }

  .user-main-info h2 {
    font-size: 24px;
    font-weight: 600;
    color: var(--text-main);
    margin: 0 0 4px 0;
  }

  .user-email {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0;
  }

  .user-meta-info {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
    margin-bottom: 20px;
  }

  .meta-item {
    display: flex;
    align-items: center;
  }

  .meta-label {
    font-size: 13px;
    color: var(--text-secondary);
    margin-right: 8px;
  }

  .meta-value {
    font-size: 13px;
    color: var(--text-main);
    font-weight: 500;
  }

  .status-active {
    color: var(--color-success);
  }

  .user-actions {
    align-self: flex-start;
  }

  /* 设置容器 */
  .settings-container {
    display: grid;
    grid-template-columns: 1fr;
    gap: 24px;
  }

  /* 设置卡片 */
  .setting-card {
    background-color: var(--bg-container);
    border-radius: 16px;
    overflow: hidden;
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border-color-light);
    transition: all 0.3s ease;
  }

  .setting-card:hover {
    box-shadow: var(--shadow-md);
    transform: translateY(-2px);
  }

  .setting-header {
    display: flex;
    align-items: center;
    padding: 24px 30px;
    background-color: var(--bg-container);
    border-bottom: 1px solid var(--border-color-light);
  }

  .setting-icon {
    width: 40px;
    height: 40px;
    background-color: rgba(22, 93, 255, 0.1);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 16px;
    color: var(--primary-color);
    font-size: 18px;
  }

  .setting-title-section h3 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-main);
    margin: 0 0 4px 0;
  }

  .setting-title-section p {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0;
  }

  .setting-content {
    padding: 24px 30px;
  }

  /* 表单布局 */
  .form-layout {
    max-width: 600px;
  }

  .el-form-item {
    margin-bottom: 24px;
  }

  .el-form-item__label {
    color: var(--text-main) !important;
    font-weight: 500;
    font-size: 14px;
  }

  .el-input__wrapper {
    border-radius: 10px !important;
    background-color: var(--bg-muted) !important;
    height: 40px !important;
    transition: all 0.3s ease;
  }

  .el-input__wrapper:focus-within {
    box-shadow: 0 0 0 3px rgba(22, 93, 255, 0.2) !important;
    border-color: var(--primary-color) !important;
  }

  .form-actions {
    display: flex;
    justify-content: flex-start;
    margin-top: 16px;
    padding-top: 20px;
    border-top: 1px solid var(--border-light);
  }

  /* 安全选项 */
  .security-option {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    border-bottom: 1px solid var(--border-light);
  }

  .security-option:last-child {
    border-bottom: none;
  }

  .security-action.full-width {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .security-note {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 8px;
    padding: 8px 12px;
    background-color: rgba(255, 170, 0, 0.1);
    border-radius: 6px;
    border-left: 3px solid #ffaa00;
    width: 100%;
  }

  .security-info h4 {
    font-size: 15px;
    font-weight: 500;
    color: var(--text-main);
    margin: 0 0 4px 0;
  }

  .security-info p {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 0;
  }

  .security-action {
    display: flex;
    align-items: center;
  }

  .ml-4 {
    margin-left: 16px;
  }

  /* 按钮样式 */
  .el-button {
    border-radius: 10px !important;
    transition: all 0.3s ease;
    font-size: 14px;
    font-weight: 500;
    padding: 8px 20px;
  }

  .btn-primary {
    padding: 10px 32px;
    font-size: 15px;
  }

  .el-button--primary {
    background-color: var(--color-primary) !important;
    border-color: var(--color-primary) !important;
  }

  .el-button--primary:hover {
    background-color: var(--color-primary-hover) !important;
    border-color: var(--color-primary-hover) !important;
    transform: translateY(-1px);
    box-shadow: 0 6px 16px rgba(22, 93, 255, 0.3);
  }

  .el-button--success {
    background-color: var(--color-success) !important;
    border-color: var(--color-success) !important;
  }

  .el-button--success:hover {
    background-color: var(--color-success-hover) !important;
    border-color: var(--color-success-hover) !important;
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(19, 206, 102, 0.3);
  }

  /* 开关样式 */
  .el-switch {
    --el-switch-on-color: var(--color-success);
    --el-switch-off-color: var(--border-color-light);
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .account-page {
      margin: 15px auto;
      padding: 0 15px;
    }

    .header {
      margin-bottom: 25px;
    }

    .header h1 {
      font-size: 26px;
    }

    .user-info-card {
      flex-direction: column;
      text-align: center;
      padding: 20px;
      margin-bottom: 25px;
    }

    .avatar-section {
      margin-right: 0;
      margin-bottom: 20px;
    }

    .user-details {
      align-items: center;
    }

    .user-actions {
      align-self: center;
    }

    .user-meta-info {
      grid-template-columns: 1fr;
      gap: 8px;
      text-align: left;
      width: 100%;
    }

    .settings-container {
      gap: 16px;
    }

    .setting-card {
      border-radius: 12px;
    }

    .setting-header {
      padding: 20px;
    }

    .setting-content {
      padding: 20px;
    }

    .form-layout {
      max-width: 100%;
    }

    .security-option {
      flex-direction: column;
      align-items: flex-start;
      padding: 16px 0;
    }

    .security-action {
      margin-top: 12px;
      align-self: stretch;
      justify-content: space-between;
    }

    .form-actions {
      justify-content: center;
    }
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

  .setting-card {
    animation: fadeIn 0.4s ease-out;
  }

  .setting-card:nth-child(2) {
    animation-delay: 0.1s;
  }

  .setting-card:nth-child(3) {
    animation-delay: 0.2s;
  }

  /* 白名单样式 */
  .whitelist-container {
    max-width: 400px;
  }

  .whitelist-item {
    display: flex;
    align-items: center;
    margin-bottom: 10px;
  }

  .whitelist-item .el-input {
    flex: 1;
    margin-right: 10px;
  }

  .whitelist-limit-tip {
    color: #909399;
    font-size: 12px;
    margin-top: 10px;
  }
</style>