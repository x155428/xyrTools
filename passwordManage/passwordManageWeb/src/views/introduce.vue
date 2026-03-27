<template>
  <div class="system-intro">
    <header class="intro-header">
      <h1>🔒 密码管理系统</h1>
      <p>安全与便捷的密码管理。</p>
    </header>

    <main class="intro-content">
      <section class="overview-card">
        <h2>系统概述</h2>
        <p>
          本系统通过加密技术和配置，帮助用户高效、安全地管理个人密码。适用于个人密码集中管理，兼顾数据移动与安全需求。
        </p>
      </section>

      <section class="features-card">
        <h2>核心特色</h2>
        <ul>
          <li>
            <strong>字段级加密：</strong>自由选择记录字段加密，确保重要数据的安全性。
          </li>
          <li>
            <strong>密钥文件管理：</strong>支持密钥文件备份与恢复，数据迁移更轻松。
          </li>
          <li>
            <strong>密码工具：</strong>内置密码生成器与强度评估器，助您生成高安全性密码。
          </li>
          <li>
            <strong>内置小工具：</strong>包含常用加密、解密、编码工具，便捷处理日常需求，纯前端。
          </li>
          <li>
            <strong>SQLite3 数据库：</strong>轻量级存储，便于数据移动与本地加密存储。
          </li>
          <li>
            <strong>前后端分离架构：</strong>安全高效，中间通信加密。
          </li>
        </ul>
      </section>

      <section class="use-case-card">
        <h2>适用场景</h2>
        <p>
          适合管理社交媒体账号、银行账户、ssh登录密钥、工作系统等多种平台密码，支持个人隐私数据的集中管理。
        </p>
      </section>
    </main>

    <footer class="intro-footer">
      <p>✨ 让密码管理变得简单安全。</p>
    </footer>
  </div>
</template>

<script setup>
// 静态展示组件，可按需求扩展交互功能。
</script>

<style scoped>
.system-intro {
  padding: 1.5rem;
  font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
  color: #333;
  line-height: 1.6;
  background: linear-gradient(120deg, #f5f7fa, #c3cfe2);
  border-radius: 8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.intro-header {
  text-align: center;
  margin-bottom: 2rem;
}

.intro-header h1 {
  font-size: 2.2rem;
  font-weight: bold;
  color: #0078d7;
}

.intro-header p {
  font-size: 1.2rem;
  color: #555;
  margin-top: 0.5rem;
}

.intro-content {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.overview-card,
.features-card,
.use-case-card {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s, box-shadow 0.3s;
}

.overview-card:hover,
.features-card:hover,
.use-case-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 6px 10px rgba(0, 0, 0, 0.15);
}

h2 {
  font-size: 1.5rem;
  color: #0078d7;
  border-left: 4px solid #0078d7;
  padding-left: 0.5rem;
  margin-bottom: 1rem;
}

ul {
  list-style-type: none;
  padding-left: 0;
}

ul li {
  margin-bottom: 0.8rem;
}

ul li::before {
  content: "✔️";
  margin-right: 0.5rem;
  color: #28a745;
}

p {
  margin: 0.5rem 0;
}

.intro-footer {
  text-align: center;
  margin-top: 2rem;
  font-size: 1rem;
  color: #777;
  font-style: italic;
}
</style>
