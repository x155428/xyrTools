<template>
  <div class="json-tool">
    <div class="container">
      <header class="header">
        <h1 class="title">JSON 工具箱</h1>
        <p class="subtitle">格式化、压缩、生成和导出JSON数据</p>
      </header>

      <!-- 主要功能标签页 -->
      <div class="main-panel">
        <div class="tab-header">
          <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id"
            :class="['tab-button', { active: activeTab === tab.id }]">
            {{ tab.name }}
          </button>
        </div>

        <div class="tab-content">
          <!-- JSON编辑器标签页 -->
          <div v-if="activeTab === 'editor'" class="editor-tab">
            <div class="editor-grid">
              <!-- 输入区域 -->
              <div class="input-section">
                <div class="section-header">
                  <h3 class="section-title">JSON 输入</h3>
                  <div class="button-group">
                    <input type="file" accept=".json" @change="handleFileUpload" ref="fileInput" style="display: none">
                    <button @click="$refs.fileInput.click()" class="btn btn-success">
                      上传文件
                    </button>
                    <button @click="clearInput" class="btn btn-danger">
                      清空
                    </button>
                  </div>
                </div>

                <textarea v-model="inputJson" @input="validateAndFormat" placeholder="在此输入或粘贴JSON数据..."
                  class="json-textarea"></textarea>

                <div v-if="validationError" class="error-message">
                  <p>{{ validationError }}</p>
                </div>
              </div>

              <!-- 输出区域 -->
              <div class="output-section">
                <div class="section-header">
                  <h3 class="section-title">格式化输出</h3>
                  <div class="button-group">
                    <button @click="formatType = 'formatted'"
                      :class="['btn', formatType === 'formatted' ? 'btn-primary active' : 'btn-secondary']">
                      格式化
                    </button>
                    <button @click="formatType = 'compressed'"
                      :class="['btn', formatType === 'compressed' ? 'btn-primary active' : 'btn-secondary']">
                      压缩
                    </button>
                    <button @click="copyToClipboard(outputJson)" class="btn btn-success">
                      复制
                    </button>
                    <button @click="downloadJson(outputJson, 'formatted.json')" class="btn btn-purple">
                      下载
                    </button>
                  </div>
                </div>

                <textarea v-model="outputJson" readonly class="json-textarea readonly"></textarea>
              </div>
            </div>
          </div>

          <!-- 批量生成标签页 -->
          <div v-if="activeTab === 'generator'" class="generator-tab">
            <div class="generator-grid">
              <!-- 模板定义 -->
              <div class="template-section">
                <h3 class="section-title">JSON 模板</h3>
                <p class="help-text">
                  使用 {{variable}} 语法定义变量，支持的变量类型：id, name, age, email, boolean, number, string
                </p>

                <textarea v-model="template"
                  placeholder='例如：{"id": {{id}}, "name": {{name}}, "age": {{age}}, "email": {{email}}}'
                  class="json-textarea template-textarea"></textarea>

                <div class="generator-controls">
                  <label class="control-label">生成数量:</label>
                  <input v-model.number="generateCount" type="number" min="1" max="1000" class="number-input">
                  <button @click="generateJsonData" class="btn btn-success">
                    生成数据
                  </button>
                </div>
              </div>

              <!-- 生成结果 -->
              <div class="result-section">
                <div class="section-header">
                  <h3 class="section-title">生成结果</h3>
                  <div class="button-group">
                    <button @click="copyToClipboard(JSON.stringify(generatedData, null, 2))" class="btn btn-success"
                      :disabled="!generatedData.length">
                      复制全部
                    </button>
                    <button @click="downloadGeneratedData" class="btn btn-purple" :disabled="!generatedData.length">
                      批量下载
                    </button>
                  </div>
                </div>

                <div class="result-container">
                  <div v-if="generatedData.length === 0" class="empty-state">
                    暂无生成的数据
                  </div>
                  <div v-else class="result-list">
                    <div v-for="(item, index) in generatedData" :key="index" class="result-item">
                      <div class="result-item-header">
                        <span class="item-label">项目 {{ index + 1 }}</span>
                        <div class="item-actions">
                          <button @click="copyToClipboard(JSON.stringify(item, null, 2))" class="btn-small btn-success">
                            复制
                          </button>
                          <button @click="downloadJson(JSON.stringify(item, null, 2), `item-${index + 1}.json`)"
                            class="btn-small btn-purple">
                            下载
                          </button>
                        </div>
                      </div>
                      <pre class="result-preview">{{ JSON.stringify(item, null, 2) }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 通知提示 -->
      <div v-if="notification.show" :class="['notification', notification.type]">
        {{ notification.message }}
      </div>
    </div>
  </div>
</template>

<script>
  import { ref, computed, watch } from 'vue'

  export default {
    name: 'JsonTool',
    setup() {
      // 响应式数据
      const activeTab = ref('editor')
      const inputJson = ref('')
      const formatType = ref('formatted')
      const validationError = ref('')
      const template = ref('{\n  "id": {{id}},\n  "name": {{name}},\n  "age": {{age}},\n  "email": {{email}},\n  "active": {{boolean}}\n}')
      const generateCount = ref(5)
      const generatedData = ref([])
      const notification = ref({ show: false, message: '', type: 'success' })

      // 标签页配置
      const tabs = [
        { id: 'editor', name: 'JSON 编辑器' },
        { id: 'generator', name: '批量生成' }
      ]

      // 计算输出JSON
      const outputJson = computed(() => {
        if (!inputJson.value.trim() || validationError.value) {
          return ''
        }

        try {
          const parsed = JSON.parse(inputJson.value)
          return formatType.value === 'formatted'
            ? JSON.stringify(parsed, null, 2)
            : JSON.stringify(parsed)
        } catch (error) {
          return ''
        }
      })

      // 验证和格式化JSON
      const validateAndFormat = () => {
        validationError.value = ''

        if (!inputJson.value.trim()) {
          return
        }

        try {
          JSON.parse(inputJson.value)
        } catch (error) {
          validationError.value = `JSON格式错误: ${error.message}`
        }
      }

      // 清空输入
      const clearInput = () => {
        inputJson.value = ''
        validationError.value = ''
      }

      // 文件上传处理
      const handleFileUpload = (event) => {
        const file = event.target.files[0]
        if (file && file.type === 'application/json') {
          const reader = new FileReader()
          reader.onload = (e) => {
            inputJson.value = e.target.result
            validateAndFormat()
          }
          reader.readAsText(file)
        } else {
          showNotification('请选择有效的JSON文件', 'error')
        }
        event.target.value = ''
      }

      // 生成随机数据的辅助函数
      const generateRandomValue = (type) => {
        const names = ['张三', '李四', '王五', '赵六', '钱七', '孙八', '周九', '吴十', '陈一', '林二']
        const emails = ['example.com', 'test.com', 'demo.com', 'sample.com', 'mail.com']

        switch (type.toLowerCase()) {
          case 'id':
            return Math.floor(Math.random() * 10000) + 1
          case 'name':
            return names[Math.floor(Math.random() * names.length)]
          case 'age':
            return Math.floor(Math.random() * 60) + 18
          case 'email':
            const randomName = 'user' + Math.floor(Math.random() * 1000)
            const domain = emails[Math.floor(Math.random() * emails.length)]
            return `${randomName}@${domain}`
          case 'boolean':
            return Math.random() > 0.5
          case 'number':
            return Math.floor(Math.random() * 1000) + 1
          case 'string':
            return '随机字符串' + Math.floor(Math.random() * 1000)
          case 'phone':
            return '1' + Math.floor(Math.random() * 9000000000 + 1000000000)
          case 'date':
            const start = new Date(2020, 0, 1)
            const end = new Date()
            const randomDate = new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()))
            return randomDate.toISOString().split('T')[0]
          default:
            // 如果是未知类型，尝试根据变量名推断
            const varName = type.toLowerCase()
            if (varName.includes('id')) return Math.floor(Math.random() * 10000) + 1
            if (varName.includes('name')) return names[Math.floor(Math.random() * names.length)]
            if (varName.includes('age')) return Math.floor(Math.random() * 60) + 18
            if (varName.includes('email')) return `user${Math.floor(Math.random() * 1000)}@example.com`
            if (varName.includes('phone')) return '1' + Math.floor(Math.random() * 9000000000 + 1000000000)
            if (varName.includes('date')) {
              const start = new Date(2020, 0, 1)
              const end = new Date()
              const randomDate = new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()))
              return randomDate.toISOString().split('T')[0]
            }
            return Math.floor(Math.random() * 100)
        }
      }

      // 生成JSON数据
      const generateJsonData = () => {
        if (!template.value.trim()) {
          showNotification('请输入模板', 'error')
          return
        }

        try {
          // 首先验证模板格式 - 将所有变量替换为合适的测试值来验证JSON结构
          let testTemplate = template.value
          const variableMatches = template.value.match(/\{\{[^}]+\}\}/g)

          if (variableMatches) {
            variableMatches.forEach(match => {
              const varName = match.slice(2, -2).trim()
              // 根据变量类型生成测试值
              let testValue
              if (['id', 'age', 'number'].includes(varName.toLowerCase()) ||
                varName.toLowerCase().includes('id') ||
                varName.toLowerCase().includes('age') ||
                varName.toLowerCase().includes('count') ||
                varName.toLowerCase().includes('num')) {
                testValue = 1 // 数字类型不需要引号
              } else if (['boolean', 'active', 'enabled'].includes(varName.toLowerCase()) ||
                varName.toLowerCase().includes('is') ||
                varName.toLowerCase().includes('has')) {
                testValue = true // 布尔类型不需要引号
              } else {
                testValue = 'test' // 字符串类型，先不加引号
              }

              // 检查模板中变量是否被双引号包裹
              const isQuotedInTemplate = (new RegExp(`(["'])${match.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\1`)).test(template.value)
              if (typeof testValue === 'string' && !isQuotedInTemplate) {
                testValue = `"${testValue}"` // 只有字符串类型且模板中未被引号包裹时才添加引号
              }

              // 转换为字符串进行替换
              const replacementValue = String(testValue)
              testTemplate = testTemplate.replace(new RegExp(match.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g'), replacementValue)
            })
          }

          // 验证模板JSON结构
          JSON.parse(testTemplate)

          const results = []

          for (let i = 0; i < generateCount.value; i++) {
            let jsonString = template.value

            // 如果有变量，则进行替换
            if (variableMatches) {
              // 获取所有唯一变量
              const uniqueVars = [...new Set(variableMatches.map(match => match.slice(2, -2).trim()))]

              // 为每个变量生成值并替换
              uniqueVars.forEach(varName => {
                const regex = new RegExp(`\\{\\{\\s*${varName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*\\}\\}`, 'g')
                const value = generateRandomValue(varName)

                // 根据值类型决定JSON格式
                let replacementValue
                if (typeof value === 'string') {
                  // 字符串需要加引号，并转义特殊字符
                  replacementValue = `"${value.replace(/"/g, '\\"')}"`
                } else if (typeof value === 'boolean') {
                  // 布尔值直接转换
                  replacementValue = String(value)
                } else if (typeof value === 'number') {
                  // 数字直接转换
                  replacementValue = String(value)
                } else {
                  // 其他类型转为字符串
                  replacementValue = `"${String(value).replace(/"/g, '\\"')}"`
                }

                jsonString = jsonString.replace(regex, replacementValue)
              })
            }

            // 解析生成的JSON
            const parsed = JSON.parse(jsonString)
            results.push(parsed)
          }

          generatedData.value = results
          if (variableMatches) {
            showNotification(`成功生成 ${results.length} 条数据`, 'success')
          } else {
            showNotification(`成功复制 ${results.length} 条相同数据`, 'success')
          }
        } catch (error) {
          console.error('生成错误:', error)
          showNotification(`生成失败: ${error.message}`, 'error')
        }
      }

      // 复制到剪贴板
      const copyToClipboard = async (text) => {
        try {
          if (navigator.clipboard && window.isSecureContext) {
            await navigator.clipboard.writeText(text)
          } else {
            // 降级方案
            const textArea = document.createElement('textarea')
            textArea.value = text
            textArea.style.position = 'fixed'
            textArea.style.left = '-999999px'
            textArea.style.top = '-999999px'
            document.body.appendChild(textArea)
            textArea.focus()
            textArea.select()
            document.execCommand('copy')
            textArea.remove()
          }
          showNotification('已复制到剪贴板', 'success')
        } catch (error) {
          showNotification('复制失败', 'error')
        }
      }

      // 下载JSON文件
      const downloadJson = (jsonString, filename) => {
        try {
          const blob = new Blob([jsonString], { type: 'application/json;charset=utf-8' })
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = filename
          a.style.display = 'none'
          document.body.appendChild(a)
          a.click()
          document.body.removeChild(a)
          URL.revokeObjectURL(url)
          showNotification('文件已下载', 'success')
        } catch (error) {
          showNotification('下载失败', 'error')
        }
      }

      // 批量下载生成的数据
      const downloadGeneratedData = () => {
        if (generatedData.value.length === 0) return

        // 创建包含所有数据的数组
        const allData = JSON.stringify(generatedData.value, null, 2)
        downloadJson(allData, 'generated-data.json')
      }

      // 显示通知
      const showNotification = (message, type = 'success') => {
        notification.value = { show: true, message, type }
        setTimeout(() => {
          notification.value.show = false
        }, 3000)
      }

      // 监听格式类型变化
      watch(formatType, () => {
        if (inputJson.value && !validationError.value) {
          validateAndFormat()
        }
      })

      return {
        // 响应式数据
        activeTab,
        inputJson,
        formatType,
        validationError,
        template,
        generateCount,
        generatedData,
        notification,
        tabs,

        // 计算属性
        outputJson,

        // 方法
        validateAndFormat,
        clearInput,
        handleFileUpload,
        generateJsonData,
        copyToClipboard,
        downloadJson,
        downloadGeneratedData,
        showNotification
      }
    }
  }
</script>

<style scoped>
  /* 基础样式重置和变量定义 */
  .json-tool {
    --primary-color: #3b82f6;
    --primary-hover: #2563eb;
    --success-color: #10b981;
    --success-hover: #059669;
    --danger-color: #ef4444;
    --danger-hover: #dc2626;
    --purple-color: #8b5cf6;
    --purple-hover: #7c3aed;
    --secondary-color: #6b7280;
    --secondary-hover: #4b5563;
    --bg-primary: #1f2937;
    --bg-secondary: #374151;
    --bg-tertiary: #4b5563;
    --text-primary: #f9fafb;
    --text-secondary: #d1d5db;
    --text-muted: #9ca3af;
    --border-color: #4b5563;
    --border-hover: #6b7280;
    --shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    --radius: 8px;
    --transition: all 0.2s ease-in-out;

    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    background-color: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
    line-height: 1.6;
  }

  .container {
    max-width: 100%;
    max-height: 100%;
    margin: 0 auto;
    padding: 2rem 1rem;
  }

  /* 头部样式 */
  .header {
    text-align: center;
    margin-bottom: 2rem;
    max-width: 100%;
  }

  .title {
    font-size: 2.5rem;
    font-weight: 700;
    background: linear-gradient(135deg, var(--primary-color), var(--purple-color));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: var(--text-muted);
    font-size: 1.1rem;
    margin: 0;
  }

  /* 主面板样式 */
  .main-panel {
    background-color: var(--bg-secondary);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    overflow: hidden;
    max-width: 100%;
  }

  /* 标签页头部 */
  .tab-header {
    display: flex;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
  }

  .tab-button {
    flex: 1;
    padding: 1rem 1.5rem;
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: var(--transition);
    border-bottom: 3px solid transparent;
  }

  .tab-button:hover {
    color: var(--text-primary);
    background-color: rgba(59, 130, 246, 0.1);
  }

  .tab-button.active {
    color: var(--text-primary);
    background-color: var(--primary-color);
    border-bottom-color: var(--primary-hover);
  }

  /* 标签页内容 */
  .tab-content {
    padding: 2rem;
    max-width: 100%;
    max-height: 100%;
  }

  /* 编辑器标签页 */
  .editor-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }

  @media (max-width: 768px) {
    .editor-grid {
      grid-template-columns: 1fr;
    }
  }

  /* 生成器标签页 */
  .generator-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }

  @media (max-width: 768px) {
    .generator-grid {
      grid-template-columns: 1fr;
    }
  }

  /* 区域样式 */
  .input-section,
  .output-section,
  .template-section,
  .result-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .section-title {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--primary-color);
    margin: 0;
  }

  .help-text {
    color: var(--text-muted);
    font-size: 0.875rem;
    margin: 0;
    line-height: 1.5;
  }

  /* 按钮组 */
  .button-group {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  /* 按钮样式 */
  .btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: var(--radius);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: var(--transition);
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 80px;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: var(--primary-color);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: var(--primary-hover);
  }

  .btn-secondary {
    background-color: var(--secondary-color);
    color: white;
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: var(--secondary-hover);
  }

  .btn-success {
    background-color: var(--success-color);
    color: white;
  }

  .btn-success:hover:not(:disabled) {
    background-color: var(--success-hover);
  }

  .btn-danger {
    background-color: var(--danger-color);
    color: white;
  }

  .btn-danger:hover:not(:disabled) {
    background-color: var(--danger-hover);
  }

  .btn-purple {
    background-color: var(--purple-color);
    color: white;
  }

  .btn-purple:hover:not(:disabled) {
    background-color: var(--purple-hover);
  }

  .btn-small {
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
    min-width: 60px;
  }

  /* 文本区域 */
  .json-textarea {
    width: 100%;
    height: 320px;
    padding: 1rem;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    color: var(--text-primary);
    font-family: 'Fira Code', 'Monaco', 'Cascadia Code', 'Roboto Mono', monospace;
    font-size: 0.875rem;
    line-height: 1.5;
    resize: vertical;
    transition: var(--transition);
  }

  .json-textarea:focus {
    outline: none;
    border-color: var(--primary-color);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .json-textarea.readonly {
    background-color: var(--bg-tertiary);
    cursor: default;
  }

  .template-textarea {
    height: 240px;
  }

  /* 生成器控件 */
  .generator-controls {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .control-label {
    color: var(--text-secondary);
    font-weight: 500;
  }

  .number-input {
    padding: 0.5rem;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    color: var(--text-primary);
    width: 100px;
    font-size: 0.875rem;
  }

  .number-input:focus {
    outline: none;
    border-color: var(--primary-color);
  }

  /* 结果容器 */
  .result-container {
    height: 400px;
    overflow-y: auto;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    padding: 1rem;
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
    font-style: italic;
  }

  .result-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .result-item {
    background-color: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    padding: 1rem;
  }

  .result-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
  }

  .item-label {
    color: var(--primary-color);
    font-weight: 500;
    font-size: 0.875rem;
  }

  .item-actions {
    display: flex;
    gap: 0.5rem;
  }

  .result-preview {
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 0.75rem;
    font-family: 'Fira Code', 'Monaco', 'Cascadia Code', 'Roboto Mono', monospace;
    font-size: 0.75rem;
    line-height: 1.4;
    color: var(--text-secondary);
    overflow-x: auto;
    margin: 0;
    white-space: pre-wrap;
  }

  /* 错误消息 */
  .error-message {
    padding: 0.75rem;
    background-color: rgba(239, 68, 68, 0.1);
    border: 1px solid var(--danger-color);
    border-radius: var(--radius);
    color: var(--danger-color);
  }

  .error-message p {
    margin: 0;
    font-size: 0.875rem;
  }

  /* 通知样式 */
  .notification {
    position: fixed;
    top: 1rem;
    right: 1rem;
    padding: 1rem 1.5rem;
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    font-weight: 500;
    z-index: 1000;
    animation: slideIn 0.3s ease-out;
  }

  .notification.success {
    background-color: var(--success-color);
    color: white;
  }

  .notification.error {
    background-color: var(--danger-color);
    color: white;
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }

    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  /* 自定义滚动条 */
  .result-container::-webkit-scrollbar,
  .json-textarea::-webkit-scrollbar {
    width: 8px;
  }

  .result-container::-webkit-scrollbar-track,
  .json-textarea::-webkit-scrollbar-track {
    background: var(--bg-secondary);
    border-radius: 4px;
  }

  .result-container::-webkit-scrollbar-thumb,
  .json-textarea::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 4px;
  }

  .result-container::-webkit-scrollbar-thumb:hover,
  .json-textarea::-webkit-scrollbar-thumb:hover {
    background: var(--border-hover);
  }

  /* 响应式设计 */
  @media (max-width: 640px) {
    .container {
      padding: 1rem 0.5rem;
    }

    .tab-content {
      padding: 1rem;
    }

    .title {
      font-size: 2rem;
    }

    .section-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .button-group {
      width: 100%;
      justify-content: flex-start;
    }

    .generator-controls {
      flex-direction: column;
      align-items: flex-start;
    }

    .json-textarea {
      height: 240px;
    }

    .result-container {
      height: 300px;
    }
  }
</style>