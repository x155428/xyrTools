<template>
  <div class="password-dashboard">
    <!-- 标题 -->
    <div class="dashboard-header">
      <h2>数据概览</h2>
    </div>

    <!-- 内容 -->
    <div class="dashboard-content">
      <!-- 数据卡片 -->
      <div class="stat-card" v-for="(item, index) in stats" :key="index">
        <div class="stat-icon" :style="{ backgroundColor: item.color }">
          <el-icon :size="32">{{ item.icon }}</el-icon>
        </div>
        <div class="stat-info">
          <h3>{{ item.title }}</h3>
          <p>{{ item.value }}</p>
        </div>
      </div>
    </div>

    <!-- 密码强度分布 -->
    <div class="strength-chart">
      <h3>密码强度分布</h3>
      <el-progress v-for="(item, index) in strength" :key="index" :percentage="item.percentage" :color="item.color"
        :stroke-width="16" :text-inside="true">
        <template #default>
          <span>{{ item.label }}</span>
        </template>
      </el-progress>
    </div>
    <!-- 更新统计按钮 -->
    <div class="update-button">
      <el-button type="primary" @click="updateStats">更新统计</el-button>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue';
  import http from '@/js/http.js';
  import { ElIcon, ElProgress } from 'element-plus';
  import { getServerAddress } from '@/js/getServerAddress.js'
  import isAxiosError from 'axios'

  // 初始化统计数据为零值，固定颜色和图标
  const stats = ref([
    { title: '记录总数', value: '0', color: '#6c8ae4', icon: '总计' },
    { title: '网站记录数量', value: '0', color: '#f5a623', icon: 'Web' },
    { title: '弱口令', value: '0', color: '#e74c3c', icon: '危' },
  ]);

  // 初始化密码强度分布为零值，固定颜色
  const strength = ref([
    { label: '非常弱', percentage: 0, color: '#ef4444' },
    { label: '弱', percentage: 0, color: '#f97316' },
    { label: '一般', percentage: 0, color: '#f59e0b' },
    { label: '强', percentage: 0, color: '#22c55e' },
    { label: '非常强', percentage: 0, color: '#10b981' },
  ]);

  // 更新统计数据
  const updateStats = async () => {

    // 使用 Axios 发送 GET 请求
    const response = await http.get("/updateStats");

    // 获取响应数据
    const data = response.data.data;

    // 判断响应数据是否为空（保留原始验证逻辑）
    if (!data || !data.stats || !data.strength) {
      ElMessage({
        message: '响应数据格式错误',
        type: 'error',
        grouping: true,
      });
      throw new Error('空的返回数据！');
    }

    // 更新统计数据，补充颜色和图标（与原始逻辑相同）
    stats.value = data.stats.map(item => ({
      ...item,
      color: getStatColor(item.title),
      icon: getStatIcon(item.title)
    }));

    // 更新密码强度分布，补充颜色
    strength.value = data.strength.map(item => ({
      ...item,
      color: getStrengthColor(item.label)
    }));
    ElMessage({
      message: "统计数据更新成功",
      type: 'success',
      grouping: true,
    });
  };

  // 根据统计项标题获取颜色
  const getStatColor = (title) => {
    switch (title) {
      case '记录总数':
        return '#6c8ae4';
      case '网站记录数量':
        return '#f5a623';
      case '弱口令':
        return '#e74c3c';
      default:
        return '#000000';
    }
  };

  // 根据统计项标题获取图标
  const getStatIcon = (title) => {
    switch (title) {
      case '记录总数':
        return '总计';
      case '网站记录数量':
        return 'Web';
      case '弱口令':
        return '危';
      default:
        return '';
    }
  };

  // 根据密码强度标签获取颜色
  const getStrengthColor = (label) => {
    switch (label) {
      case '非常弱':
        return '#ef4444';
      case '弱':
        return '#f97316';
      case '一般':
        return '#f59e0b';
      case '强':
        return '#22c55e';
      case '非常强':
        return '#10b981';
      default:
        return '#000000';
    }
  };

  // 在页面加载时调用更新统计数据的函数
  onMounted(() => {
    updateStats();
  });
</script>

<style scoped>
  /* 整体面板样式 */
  .password-dashboard {
    max-width: 1200px;
    margin: 20px auto;
    padding: 20px;
    background: var(--bg-container);
    border-radius: 12px;
    box-shadow: var(--shadow-md);
    animation: fadeIn 0.5s ease;
  }

  /* 标题 */
  .dashboard-header {
    margin-bottom: 30px;
    text-align: center;
    padding-bottom: 15px;
    border-bottom: 1px solid var(--border-color-light);
  }

  .dashboard-header h2 {
    font-size: 28px;
    color: var(--text-main);
    margin: 0;
    font-weight: 600;
  }

  /* 数据内容区域 */
  .dashboard-content {
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
    justify-content: space-between;
  }

  /* 数据卡片样式 */
  .stat-card {
    flex: 1 1 calc(33.33% - 20px);
    min-width: 200px;
    padding: 24px;
    background: var(--bg-muted);
    border-radius: 12px;
    box-shadow: var(--shadow-sm);
    display: flex;
    align-items: center;
    gap: 16px;
    transition: all 0.3s ease;
    cursor: pointer;
    border: 1px solid var(--border-color-light);
    position: relative;
    overflow: hidden;
  }

  /* 卡片悬停效果 */
  .stat-card:hover {
    transform: translateY(-5px);
    box-shadow: var(--shadow-md);
    border-color: var(--color-primary);
  }

  /* 卡片装饰元素 */
  .stat-card::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 4px;
    height: 100%;
    background: var(--color-primary);
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .stat-card:hover::after {
    opacity: 1;
  }

  /* 图标样式 */
  .stat-icon {
    width: 64px;
    height: 64px;
    border-radius: 12px;
    display: flex;
    justify-content: center;
    align-items: center;
    color: white;
    font-weight: bold;
    font-size: 24px;
    flex-shrink: 0;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }

  .stat-card:hover .stat-icon {
    transform: scale(1.1);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);
  }

  /* 文字内容 */
  .stat-info h3 {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0 0 8px 0;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .stat-info p {
    font-size: 28px;
    font-weight: bold;
    margin: 0;
    color: var(--text-main);
    line-height: 1.2;
  }

  /* 密码强度分布 */
  .strength-chart {
    margin-top: 40px;
    padding: 24px;
    background: var(--bg-muted);
    border-radius: 12px;
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border-color-light);
  }

  .strength-chart h3 {
    font-size: 20px;
    color: var(--text-main);
    margin-bottom: 24px;
    font-weight: 600;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--border-color-light);
  }

  .el-progress {
    margin-bottom: 16px;
    transition: all 0.3s ease;
  }

  /* 进度条标签 */
  .el-progress__text {
    color: var(--text-main) !important;
    font-weight: 600 !important;
    font-size: 14px !important;
  }

  /* 调整进度条的背景色 */
  .el-progress__bar {
    background-color: var(--bg-muted) !important;
  }

  /* 调整进度条的填充颜色，使其更鲜明 */
  .el-progress__inner {
    transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1) !important;
    border-radius: 4px !important;
  }

  /* 更新按钮样式 */
  .update-button {
    margin-top: 30px;
    text-align: center;
  }

  /* 按钮悬停效果 */
  .el-button--primary:hover {
    background-color: var(--color-primary-hover) !important;
    border-color: var(--color-primary-hover) !important;
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }

  .el-button--primary {
    background-color: var(--color-primary) !important;
    border-color: var(--color-primary) !important;
    padding: 10px 24px;
    font-size: 16px;
    font-weight: 500;
    transition: all 0.3s ease;
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

  /* 响应式设计 */
  @media (max-width: 1024px) {
    .stat-card {
      flex: 1 1 calc(50% - 20px);
    }
  }

  @media (max-width: 768px) {
    .password-dashboard {
      padding: 16px;
      margin: 10px;
    }

    .stat-card {
      flex: 1 1 100%;
      min-width: auto;
    }

    .stat-info p {
      font-size: 24px;
    }

    .dashboard-header h2 {
      font-size: 24px;
    }
  }

  @media (max-width: 480px) {
    .stat-card {
      padding: 16px;
    }

    .stat-icon {
      width: 50px;
      height: 50px;
      font-size: 20px;
    }

    .stat-info h3 {
      font-size: 12px;
    }

    .stat-info p {
      font-size: 20px;
    }
  }

  @media (max-width: 768px) {
    .password-dashboard {
      padding: 16px;
      margin: 10px;
    }

    .stat-card {
      flex: 1 1 100%;
      min-width: auto;
    }

    .stat-info p {
      font-size: 24px;
    }

    .dashboard-header h2 {
      font-size: 24px;
    }
  }
</style>