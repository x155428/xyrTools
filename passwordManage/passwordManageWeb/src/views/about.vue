<template>
  <div class="about-page">
    <header class="about-header">
      <h1>关于密码管理系统</h1>
    </header>

    <main class="about-content">
      <!-- 系统信息 -->
      <section class="about-card">
        <h2>系统信息</h2>
        <div class="system-info">
          <div class="info-item">
            <span class="info-label">当前版本：</span>
            <span class="info-value">v1.1.0</span>
          </div>
          <div class="info-item">
            <span class="info-label">创建日期：</span>
            <span class="info-value">2024年12月</span>
          </div>
          <div class="info-item">
            <span class="info-label">最后更新：</span>
            <span class="info-value">2025年12月18日</span>
          </div>
        </div>

        <h3>项目介绍</h3>
        <p>
          本密码管理系统采用前后端分离架构，前端基于 Vue 3 构建，后端使用 Go 语言开发，
          采用轻量级 SQLite 数据库存储。系统核心功能包括密码的安全存储、生成、强度评估，
          以及常用编码与加密工具集成。
        </p>
        <p>
          系统设计基于日常使用需求，注重安全性与易用性的平衡。所有用户数据均经过严格加密处理，
          确保密码信息的安全性。
        </p>
      </section>

      <!-- 更新历史 -->
      <section class="about-card">
        <h2>更新历史</h2>

        <div class="update-history">
          <!-- 最新版本 -->
          <div class="update-item">
            <div class="update-header">
              <span class="update-version">v1.1.0</span>
              <span class="update-date">2025年12月</span>
            </div>
            <ul class="update-changelog">
              <li>🔒 新增双重验证功能</li>
              <li>📧 新增登录提醒功能</li>
              <li>🛡️ 新增白名单访问功能</li>
              <li>🔄 完成密码更改功能</li>
              <li>🔑 实现主密钥更改功能</li>
              <li>🖼️ 新增头像功能</li>
              <li>🎨 重构整个系统排版界面，优化使用体验</li>
              <li>📤 新增批量导入功能，支持导入Chrome导出的密码记录。存储字段对应表头：name、url、username、password，可自行添加字段：notes、tags、key。其余自动忽略。</li>
            </ul>
          </div>

          <div class="update-item">
            <div class="update-header">
              <span class="update-version">v1.0.2</span>
              <span class="update-date">2025年6月</span>
            </div>
            <ul class="update-changelog">
              <li>📋 新增台账demo（暂不可用）</li>
              <li>🔄 重构字符处理工具为链式处理</li>
            </ul>
          </div>

          <div class="update-item">
            <div class="update-header">
              <span class="update-version">v1.0.1</span>
              <span class="update-date">2025年3月</span>
            </div>
            <ul class="update-changelog">
              <li>🔒 完善密码加密存储</li>
              <li>🎨 优化界面设计</li>
              <li>🔢 新增密码生成功能</li>
              <li>📊 新增密码强度评估功能</li>
            </ul>
          </div>

          <div class="update-item">
            <div class="update-header">
              <span class="update-version">v1.0.0</span>
              <span class="update-date">2024年12月</span>
            </div>
            <ul class="update-changelog">
              <li>✨ 初始版本发布</li>
              <li>🔒 实现基础密码安全存储功能</li>
              <li>🔄 集成常用编码与加密工具、校验工具、文本比较工具</li>
            </ul>
          </div>
        </div>
      </section>
    </main>

    <footer class="about-footer">
      <p>© 2024-2025 密码管理系统 | 保留所有权利</p>
    </footer>
  </div>
</template>

<script setup>
</script>

<style scoped>
  /* 页面整体样式 */
  .about-page {
    max-width: 100%;
    margin: 20px auto;
    padding: var(--space-md);
    width: 100%;
    color: var(--text-main);
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    line-height: 1.7;
    overflow-x: hidden;
    word-wrap: break-word;
    hyphens: auto;
  }

  /* 确保所有元素都支持自适应换行 */
  * {
    box-sizing: border-box;
    word-wrap: break-word;
  }

  /* 防止长单词或URL导致溢出 */
  p,
  h1,
  h2,
  h3,
  div,
  span {
    overflow-wrap: break-word;
    word-break: break-word;
  }

  /* 滚动条美化 */
  .about-page::-webkit-scrollbar {
    width: 6px;
  }

  .about-page::-webkit-scrollbar-track {
    background: var(--bg-muted);
    border-radius: 3px;
  }

  .about-page::-webkit-scrollbar-thumb {
    background: var(--border-color-light);
    border-radius: 3px;
  }

  .about-page::-webkit-scrollbar-thumb:hover {
    background: var(--text-secondary);
  }

  /* 页面头部 */
  .about-header {
    text-align: center;
    margin-bottom: 32px;
    padding: 30px 0 20px;
    border-bottom: 1px solid var(--border-color-light);
  }

  .about-header h1 {
    font-size: 2.5rem;
    font-weight: 700;
    color: var(--color-primary);
    margin-bottom: 12px;
    letter-spacing: -0.5px;
  }

  /* 内容区域 */
  .about-content {
    display: grid;
    grid-template-columns: 1fr;
    gap: 28px;
  }

  /* 卡片样式 */
  .about-card {
    background-color: var(--bg-container);
    border-radius: 16px;
    padding: 30px;
    box-shadow: var(--shadow-sm);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--border-color-light);
  }

  .about-card:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-md);
  }

  /* 标题样式 */
  h2 {
    font-size: 1.6rem;
    color: var(--color-primary);
    border-left: 4px solid var(--color-primary);
    padding-left: 16px;
    margin-bottom: 20px;
    font-weight: 600;
    line-height: 1.3;
  }

  h3 {
    font-size: 1.3rem;
    color: var(--text-main);
    margin: 24px 0 16px 0;
    font-weight: 600;
  }

  /* 段落样式 */
  p {
    margin: 0 0 16px 0;
    color: var(--text-secondary);
    font-size: 1rem;
    text-align: left;
  }

  /* 强调文本 */
  strong {
    color: var(--text-main);
    font-weight: 600;
  }

  /* 系统信息样式 */
  .system-info {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
    margin-bottom: 24px;
    padding: 16px;
    background-color: var(--bg-muted);
    border-radius: 8px;
  }

  .info-item {
    display: flex;
    flex-direction: column;
  }

  .info-label {
    font-size: 0.9rem;
    color: var(--text-muted);
    margin-bottom: 4px;
  }

  .info-value {
    font-size: 1.1rem;
    color: var(--text-main);
    font-weight: 600;
  }

  /* 更新历史样式 */
  .update-history {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .update-item {
    padding: 16px;
    background-color: var(--bg-muted);
    border-radius: 8px;
    transition: all 0.3s ease;
  }

  .update-item:hover {
    background-color: var(--bg-hover);
    transform: translateY(-2px);
  }

  .update-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border-color-light);
  }

  .update-version {
    font-size: 1.1rem;
    font-weight: 700;
    color: var(--color-primary);
  }

  .update-date {
    font-size: 0.9rem;
    color: var(--text-muted);
  }

  /* 更新日志列表样式 */
  .update-changelog {
    list-style-type: none;
    padding-left: 0;
    margin: 0;
  }

  .update-changelog li {
    margin-bottom: 8px;
    color: var(--text-secondary);
    position: relative;
    padding-left: 24px;
    font-size: 0.95rem;
  }

  .update-changelog li::before {
    content: "•";
    position: absolute;
    left: 0;
    top: 0;
    color: var(--color-primary);
    font-size: 1.2rem;
    font-weight: bold;
  }

  /* 通用列表样式 */
  ul {
    list-style-type: none;
    padding-left: 0;
    margin: 0 0 16px 0;
  }

  ul:not(.update-changelog) li {
    margin-bottom: 12px;
    color: var(--text-secondary);
    position: relative;
    padding-left: 24px;
  }

  ul:not(.update-changelog) li::before {
    content: "";
    position: absolute;
    left: 0;
    top: 8px;
    width: 16px;
    height: 16px;
    background-color: var(--color-primary);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 10px;
    font-weight: bold;
  }

  /* 邮件和链接样式 */
  ul li a {
    color: var(--color-primary);
    text-decoration: none;
    transition: all 0.3s ease;
    font-weight: 500;
  }

  ul li a:hover {
    color: var(--color-primary-hover);
    text-decoration: underline;
  }

  /* 页脚样式 */
  .about-footer {
    text-align: center;
    margin-top: var(--space-md);
    padding: var(--space-md);
    font-size: 0.95rem;
    color: var(--text-muted);
    font-style: italic;
    border-top: 1px solid var(--border-color-light);
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .about-page {
      padding: 0 16px;
      margin: 10px auto;
    }

    .about-header {
      padding: 20px 0 16px;
      margin-bottom: 24px;
    }

    .about-header h1 {
      font-size: 2rem;
    }

    .about-card {
      padding: 24px;
    }

    h2 {
      font-size: 1.4rem;
      margin-bottom: 16px;
      padding-left: 12px;
    }

    p {
      font-size: 0.95rem;
      margin-bottom: 14px;
    }
  }

  @media (max-width: 480px) {
    .about-header h1 {
      font-size: 1.8rem;
    }

    .about-card {
      padding: 20px;
    }

    h2 {
      font-size: 1.3rem;
    }

    ul li {
      padding-left: 24px;
    }

    ul li::before {
      width: 14px;
      height: 14px;
    }
  }
</style>