<template>
  <div class="common-layout">
    <!--垂直容器布局-->
    <el-container class="app-container">
      <!--顶部头部，包含系统名称、水平菜单、跑马灯和用户信息-->
      <el-header class="app-header">
        <!--顶部第一行：系统名称、水平菜单和用户信息-->
        <div class="header-main">
          <!--左侧：系统名称-->
          <div class="header-section">
            <div class="system-name">
              <svg t="1730969397262" class="icon" viewBox="0 0 1024 1024" version="1.1"
                xmlns="http://www.w3.org/2000/svg" p-id="6748" width="20" height="20">
                <path
                  d="M634.571245 772.507235h-67.882244c-238.098795 0-298.937344-164.231235-298.937344-304.192744l5.839333-193.427899 194.084824-10.948749c155.143774-4.343004 295.616224 71.67781 293.316987 296.638107v67.444294c0 53.539383 32.846247 96.896429 86.349134 96.896429h160.581652c0-79.086464 0.547437-136.238934 0.547437-211.675815C1008.653503 229.814242 786.247915 0 512.018248 0S15.419488 229.814242 15.419488 513.350346s222.332597 510.649654 496.59876 510.649654h229.522275V858.746881c0-59.086749-47.882529-86.239646-106.969278-86.239646z"
                  p-id="6749"></path>
                <path
                  d="M392.786371 379.885095m-42.700121 0a42.700121 42.700121 0 1 0 85.400242 0 42.700121 42.700121 0 1 0-85.400242 0Z"
                  p-id="6750"></path>
                <path
                  d="M774.532754 765.719011h66.750873a166.347993 166.347993 0 0 1 166.347993 166.347993v90.071709h-66.86036A166.347993 166.347993 0 0 1 774.532754 855.790719v-90.071708z"
                  p-id="6751"></path>
              </svg>
              <span class="system-name-text">{{systemName}}</span>
            </div>
          </div>

          <!-- 分隔线 -->
          <div class="header-divider"></div>

          <!--中间：水平菜单-->
          <div class="header-section">
            <div class="header-menu">
              <!--菜单具体项：水平菜单-->
              <el-menu class="el-menu-horizontal-demo" :unique-opened="true" :router="true" mode="horizontal"
                :ellipsis="false" style="border-bottom: none;">
                <el-sub-menu index="1">
                  <template #title>
                    <span>概览</span>
                  </template>
                  <el-menu-item-group>
                    <el-menu-item index="/dashboard">
                      密码记录概览
                    </el-menu-item>
                    <el-menu-item index="/introduce">
                      系统介绍
                    </el-menu-item>
                    <el-menu-item index="/about">
                      关于
                    </el-menu-item>
                  </el-menu-item-group>
                </el-sub-menu>
                <el-sub-menu index="2">
                  <template #title>
                    <span>管理</span>
                  </template>
                  <el-menu-item-group>
                    <el-menu-item index="/passwordManage">
                      密码管理
                    </el-menu-item>
                    <el-menu-item index="/shareManage">
                      共享管理
                    </el-menu-item>
                    <el-menu-item index="/ledgerManage">
                      台账管理
                    </el-menu-item>
                  </el-menu-item-group>
                </el-sub-menu>
                <el-sub-menu index="3">
                  <template #title>
                    <span>工具</span>
                  </template>
                  <el-menu-item-group>
                    <el-menu-item index="/tools">
                      常用工具
                    </el-menu-item>
                  </el-menu-item-group>
                </el-sub-menu>
                <el-sub-menu index="4">
                  <template #title>
                    <span>系统管理</span>
                  </template>
                  <el-menu-item-group>
                    <el-menu-item index="/account">
                      用户中心
                    </el-menu-item>
                    <el-menu-item index="/set">
                      配置
                    </el-menu-item>
                  </el-menu-item-group>
                </el-sub-menu>
              </el-menu>
            </div>
          </div>

          <!-- 分隔线 -->
          <div class="header-divider"></div>

          <!--右侧：用户信息-->
          <div class="header-section">
            <div class="user-profile-content">
              <el-avatar :src="authStore.avatar || 'https://empty'" class="user-avatar" :fit="'cover'" />
              <span class="username">{{ authStore.username }}</span>
              <span @click="layoutElHeaderLogout" class="logout-link">
                退出
              </span>
            </div>
          </div>
        </div>

        <!--顶部第二行：跑马灯-->
        <div class="header-marquee">
          <el-carousel height="40px" arrow="never" indicator-position="none" :interval="6000" trigger="click"
            direction="vertical">
            <el-carousel-item>
              <!--诗词 -->
              <div class="marquee-item marquee-quote">
                <div class="quote-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M9 11H7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2h-2"></path>
                    <path d="M9 19h7a2 2 0 0 1 2-2v-4a2 2 0 0 1-2-2H9a2 2 0 0 1-2 2v4a2 2 0 0 1 2 2z"></path>
                  </svg>
                </div>
                <span class="quote-text">{{tmpRandomWord}}</span>
              </div>
            </el-carousel-item>
            <el-carousel-item>
              <!--天气 -->
              <div class="marquee-item marquee-weather">
                <WeatherForecast />
              </div>
            </el-carousel-item>
            <el-carousel-item>
              <!--提示 -->
              <div class="marquee-item marquee-tip">
                <div class="tip-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z">
                    </path>
                    <line x1="12" y1="9" x2="12" y2="13"></line>
                    <line x1="12" y1="17" x2="12.01" y2="17"></line>
                  </svg>
                </div>
                <span class="tip-text">定期更换密码，保障账号安全</span>
              </div>
            </el-carousel-item>
          </el-carousel>
        </div>
      </el-header>

      <!--主内容区域-->
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
  import WeatherForecast from './WeatherForecast.vue';
  import { ref, onBeforeMount, onMounted, onBeforeUnmount, provide } from 'vue';
  import { useRouter } from 'vue-router';
  import { checkSession } from '@/js/checkSession.js';
  import { useAuthStore } from '@/store/auth.js';
  import http from '@/js/http.js'

  const router = useRouter();
  const authStore = useAuthStore();
  let intervalId = null;
  // 这里未来可能会接入api，现在先用固定的
  const tmpRandomWord = ref('金风玉露一相逢，便胜却人间无数。');
  /* 原有的代码保留作为注释、固定模式获取
  const words = [
    "金风玉露一相逢，便胜却人间无数。", "石门江水流，空忆南飞雁。", "世情薄，人情恶，雨送黄昏花易落。", "未经审视的生活不值得过。", "我们必须在迷雾中前行，而非停滞不前。", "世界上只有两种人，过得幸福的人和不幸福的人。", "一个人所拥有的时间，是他能赋予自己最大的自由。", "在一个谎言遍布的世界里，讲真话就是革命。", "生活就像骑单车，要保持平衡就得不断前行。", "逻辑是人生的智慧，它让一切变得简单。"
  ];

  const randomWord = () => {
    const randomIndex = Math.floor(Math.random() * words.length);
    tmpRandomWord.value = words[randomIndex];
  };
  */

  // 调用API获取词的新实现，包含作者信息
  async function randomWord() {
    try {
      const response = await fetch('https://v1.hitokoto.cn/?c=i&encode=json');
      const data = await response.json();

      // 解析数据，获取词内容和作者信息
      const content = data.hitokoto || '';
      const author = data.from_who || data.from || '未知';
      const source = data.from || '';

      // 拼接美化后的内容
      if (content) {
        if (author && author !== source) {
          // 有作者且与来源不同时显示作者和来源
          tmpRandomWord.value = `${content} —— ${author}《${source}》`;
        } else if (author) {
          // 只有作者时显示作者
          tmpRandomWord.value = `${content} —— ${author}`;
        } else {
          // 只有内容时直接显示
          tmpRandomWord.value = content;
        }
      } else {
        tmpRandomWord.value = '金风玉露一相逢，便胜却人间无数！';
      }
    } catch (error) {
      console.error('获取API数据失败:', error);
      tmpRandomWord.value = '金风玉露一相逢，便胜却人间无数！';
    }
  }

  const systemName = ref('小鱼密码管理系统');  // 系统名称
  const timeout = ref(15);
  const weatherApiKey = ref('');
  provide('systemName', systemName);   // 共享到子组件
  provide('timeout', timeout); // 共享到子组件
  provide('weatherApiKey', weatherApiKey); // 共享到子组件

  const layoutisCollapse = ref(false)
  const layoutElHeaderLogout = async () => {
    try {
      // 清除本地保存的密钥信息
      localStorage.removeItem('aesKey');
      localStorage.removeItem('aesKeyBase64');
      // 清除本地保存的天气API密钥
      localStorage.removeItem('weatherApiKey');

      // 发送请求到服务器清除
      const response = await http.get("/logout", {
        headers: {
          'Content-Type': 'application/json',
        }
      });
      // 响应处理
      const authStore = useAuthStore();
      if (response.status == 200) {
        // 清除pinia

        authStore.logout();
        ElMessage({
          message: "安全退出！",
          type: 'success',
          grouping: true,
        })
        router.push('/');
      } else {
        ElMessage({
          message: "出错，服务端数据未完全清除！",
          type: 'error',
          grouping: true,
        })
        router.push('/');
      }
    } catch (error) {
      ElMessage({
        message: "出错，数据未完全清除！",
        type: 'error',
        grouping: true,
      })
      router.push('/');
    }
  }

  // 从服务端获取头像和用户名信息
  const getAvatarAndUsername = async () => {
    try {
      const response = await http.get("/getAvatarAndUser");
      if (response.status == 200) {
        // 更新头像信息
        if (response.data.data && response.data.data.avatar) {
          // 将 Base64 编码数据转换为 Data URL 格式
          const avatarData = response.data.data.avatar.startsWith('data:image/')
            ? response.data.data.avatar
            : `data:image/png;base64,${response.data.data.avatar}`;
          authStore.updateAvatar(avatarData);
        }
        // 更新用户名信息
        if (response.data.data && response.data.data.username) {
          authStore.username = response.data.data.username;
          sessionStorage.setItem('username', response.data.data.username);
        }
      } else {
        // 尝试从sessionStorage获取备用数据
        const storedUsername = sessionStorage.getItem('username');
        if (storedUsername && !authStore.username) {
          authStore.username = storedUsername;
        }
      }
    } catch (error) {
      console.error("获取用户信息失败:", error);
      // 错误时尝试从sessionStorage获取
      const storedUsername = sessionStorage.getItem('username');
      if (storedUsername && !authStore.username) {
        authStore.username = storedUsername;
      }
    }
  };

  // 在组件挂载后添加监听器
  onMounted(async () => {
    await sysCfgInit();
    await getAvatarAndUsername(); // 获取头像和用户名信息
    randomWord(); // 初始化时设置一个随机词
    intervalId = setInterval(randomWord, Math.floor(Math.random() * (30000)) + 60000); // 每60-90秒更新一次，避免API调用频率过高
  });

  // 在组件卸载前移除监听器，清除定时器
  onBeforeUnmount(() => {
    if (intervalId) {
      clearInterval(intervalId); // 清除定时器，防止内存泄漏
    }
  });


  // 从服务端获取系统配置
  const sysCfgInit = async () => {
    try {
      const response = await http.get("/getTmpCfg");
      if (response.status == 200) {
        systemName.value = response.data.data.systemName;
        timeout.value = response.data.data.timeout;
        weatherApiKey.value = response.data.data.weatherApiKey;// 暂时天气apikey没保存到后端，必定是空字符串
        // 如果为空，尝试从本地获取
        if (!weatherApiKey.value) {
          const storedKey = localStorage.getItem('weatherApiKey');
          if (storedKey) {
            weatherApiKey.value = storedKey;
          }
        }
        // 如果服务器返回了用户头像信息
        if (response.data.data.userAvatar) {
          authStore.updateAvatar(response.data.data.userAvatar);
        }
      } else {
        ElMessage({
          message: "服务器异常，部分前端设置初始化出错！",
          type: 'error',
          grouping: true,
        })
        router.push('/');
      }
    } catch (error) {
      console.error("获取系统配置失败:", error);
      // 错误处理保持不变
      ElMessage({
        message: "服务器异常，前端初始化错误！",
        type: 'error',
        grouping: true,
      })
      router.push('/');
    }
  }

  // 更新用户头像，不再保存到本地
  const saveUserAvatar = (avatarUrl) => {
    authStore.updateAvatar(avatarUrl);
  }
</script>
<style scoped>
  /* 使用base.css中定义的CSS变量 */

  /* 容器样式 */
  .common-layout {
    width: 100%;
    height: 100vh;
    display: flex;
    flex-direction: column;
  }

  .app-container {
    background-color: var(--bg-page);
    color: var(--text-main);
    box-sizing: border-box;
    display: flex;
    flex: 1;
    flex-basis: auto;
    min-width: 0;
    height: 100%;
    width: 100%;
  }

  /* 头部布局样式 */
  .header-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0;
    box-sizing: border-box;
    background-color: var(--bg-container);
    white-space: nowrap;
  }

  .header-section {
    display: flex;
    align-items: center;
    height: 100%;
  }

  /* 左侧系统名称区域 */
  .header-section:first-child {
    padding: 0 var(--space-md) 0 var(--space-sm);
    justify-content: flex-start;
  }

  /* 中间菜单区域 */
  .header-section:nth-child(3) {
    flex: 1;
    padding: 0 var(--space-md);
    justify-content: center;
  }

  /* 右侧用户信息区域 */
  .header-section:last-child {
    padding: 0 var(--space-sm) 0 var(--space-md);
    justify-content: flex-end;
  }

  /* 分隔线样式 */
  .header-divider {
    width: 1px;
    height: 32px;
    background-color: var(--border-color);
    margin: 0;
  }

  /* 头部样式 */
  .app-header {
    background-color: var(--bg-container);
    box-shadow: none;
    border-bottom: 1px solid var(--border-color-light);
    z-index: 10;
    width: 100%;
    height: auto;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    color: var(--text-main);
  }

  .header-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0 var(--space-sm);
    box-sizing: border-box;
    background-color: var(--bg-container);
    white-space: nowrap;
  }

  /* 系统名字样式*/
  .system-name {
    max-width: 200px;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-bold);
    color: var(--text-main);
    margin: 0;
    padding: 0 var(--space-md) 0 0;
    flex-shrink: 1;
    flex-grow: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* 菜单样式 */
  .header-menu {
    flex: 1;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    margin: 0 var(--space-lg);
    flex-grow: 1;
    flex-shrink: 1;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* 水平菜单样式 */
  .el-menu-horizontal-demo {
    background-color: var(--bg-container);
    border-bottom: none;
    justify-content: flex-start;
  }

  /* 用户信息样式*/
  .user-profile-content {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    flex-shrink: 0;
    gap: 8px;
  }

  .user-avatar {
    margin-right: 4px;
  }

  .username {
    color: var(--text-main);
  }

  .logout-link {
    color: var(--text-main);
    cursor: pointer;
  }

  /* 跑马灯样式*/
  .header-marquee {
    height: 40px;
    width: 100%;
    margin: 0 var(--space-lg);
    padding: 0;
    background-color: var(--bg-container);
  }

  /* 跑马灯容器样式 */
  .el-carousel {
    height: 40px !important;
    line-height: 40px;
    color: var(--text-main);
  }

  .el-carousel-item {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }

  .marquee-item {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 100%;
    border-radius: var(--radius-md);
    transition: all 0.3s ease;
    position: relative;
    overflow: hidden;
    cursor: pointer;
    background: var(--color-primary-light);
    white-space: nowrap;
  }

  /* 引用内容样式 */
  .marquee-quote {
    padding: 0 var(--space-lg);
    gap: var(--space-sm);
    background: var(--color-primary-light);
    border-radius: var(--radius-md);
    white-space: nowrap;
    color: var(--text-main);
  }

  .quote-icon {
    color: var(--color-primary);
    flex-shrink: 0;
    animation: none;
  }

  .quote-text {
    color: var(--text-main);
    font-family: 'Microsoft YaHei', Arial, sans-serif;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    text-align: center;
    line-height: 40px;
    text-shadow: none;
    max-width: 90%;
    animation: none;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* 天气内容样式 */
  .marquee-weather {
    background: var(--color-primary-light);
    padding: 0 var(--space-lg);
    border-radius: var(--radius-md);
    white-space: nowrap;
    color: var(--text-main);
  }

  /* 提示内容样式 */
  .marquee-tip {
    background: var(--color-primary-light);
    padding: 0 var(--space-lg);
    gap: var(--space-sm);
    border-radius: var(--radius-md);
    white-space: nowrap;
    color: var(--text-main);
  }

  .tip-icon {
    color: var(--color-primary);
    flex-shrink: 0;
    animation: none;
  }

  .tip-text {
    color: var(--text-main);
    font-family: 'Microsoft YaHei', Arial, sans-serif;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    text-align: center;
    line-height: 40px;
    text-shadow: none;
    white-space: nowrap;
  }

  /* 装饰元素 */
  .marquee-decoration,
  .marquee-glow,
  .weather-decoration {
    display: none;
  }

  .app-main {
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    width: 100%;
    padding: 0;
    margin: 0;
    box-sizing: border-box;
    background-color: var(--bg-container);
    white-space: nowrap;
  }
</style>