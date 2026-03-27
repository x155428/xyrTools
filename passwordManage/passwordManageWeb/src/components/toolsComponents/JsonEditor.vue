<template>
  <div class="json-editor-tool">
    <div class="container">
      <header class="header">
        <h1 class="title">JSON编辑器</h1>
        <p class="subtitle">格式化、压缩、验证和编辑JSON数据</p>
      </header>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="tool-group">
          <input type="file" accept=".json" @change="handleFileUpload" ref="fileInput" style="display: none">
          <el-button type="primary" @click="$refs.fileInput.click()" :icon="DocumentAdd">
            上传文件
          </el-button>
          <el-button @click="setSampleData" :icon="Operation">
            示例数据
          </el-button>
          <el-button @click="clearAll" :icon="Delete" type="danger">
            清空
          </el-button>
        </div>

        <div class="tool-group">
          <el-button @click="formatJson" :icon="RefreshLeft" :disabled="!hasValidJson">
            格式化
          </el-button>
          <el-button @click="compressJson" :icon="RefreshRight" :disabled="!hasValidJson">
            压缩
          </el-button>
          <el-button @click="copyToClipboard" :icon="CopyDocument" :disabled="!hasContent">
            复制
          </el-button>
          <el-button @click="downloadJson" :icon="Download" :disabled="!hasContent">
            下载
          </el-button>
        </div>

        <div class="tool-group">
          <el-button @click="validateJson" :icon="Check"
            :class="{ 'is-success': isValidJson && hasContent, 'is-danger': !isValidJson && hasContent }">
            {{ validationStatus }}
          </el-button>
        </div>
      </div>

      <!-- 编辑器区域 -->
      <div class="editor-container">
        <v-ace-editor v-model:value="jsonContent" @init="editorInit" @input="onContentChange" lang="json"
          theme="monokai" style="height: 500px; width: 100%" :options="editorOptions" class="json-ace-editor" />
      </div>

      <!-- 错误提示 -->
      <div v-if="errorMessage" class="error-message">
        <el-alert :title="errorMessage" type="error" show-icon :closable="false" />
      </div>

      <!-- 统计信息 -->
      <div class="statistics">
        <span>字符数: {{ jsonContent.length }}</span>
        <span v-if="isValidJson">对象/数组: {{ objectCount }}</span>
        <span v-if="isValidJson">键值对: {{ keyValueCount }}</span>
        <span v-if="isValidJson">深度: {{ maxDepth }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, computed, onMounted } from 'vue'
  import { VAceEditor } from 'vue3-ace-editor'
  import * as ace from 'ace-builds'
  import {
    DocumentAdd,
    Operation,
    Delete,
    RefreshLeft,
    RefreshRight,
    CopyDocument,
    Download,
    Check
  } from '@element-plus/icons-vue'
  import { ElMessage, ElButton } from 'element-plus'

  // 引入ace编辑器主题和语言模式
  import 'ace-builds/src-noconflict/theme-monokai'
  import 'ace-builds/src-noconflict/mode-json'
  import 'ace-builds/src-noconflict/ext-language_tools'

  // 响应式数据
  const jsonContent = ref('')
  const errorMessage = ref('')
  const isValidJson = ref(false)
  const editorInstance = ref(null)

  // 编辑器选项
  const editorOptions = {
    enableBasicAutocompletion: true,
    enableSnippets: true,
    enableLiveAutocompletion: true,
    tabSize: 2,
    useSoftTabs: true,
    showPrintMargin: false,
    fontSize: 14,
    lineNumbers: true,
    highlightActiveLine: true,
    wrap: true
  }

  // 计算属性
  const hasContent = computed(() => jsonContent.value.trim().length > 0)
  const hasValidJson = computed(() => hasContent.value && isValidJson.value)

  const validationStatus = computed(() => {
    if (!hasContent.value) return '未输入'
    return isValidJson.value ? '格式正确' : '格式错误'
  })

  // 统计信息相关
  const objectCount = ref(0)
  const keyValueCount = ref(0)
  const maxDepth = ref(0)

  // 编辑器初始化
  const editorInit = (editor) => {
    editorInstance.value = editor
    // 设置编辑器自动补全
    const langTools = ace.require('ace/ext/language_tools')

    // 自定义JSON自动补全
    const jsonCompleter = {
      getCompletions: function (editor, session, pos, prefix, callback) {
        try {
          const line = session.getLine(pos.row)
          // 检查当前是否在字符串内
          const inString = line.substring(0, pos.column).match(/"[^"\\]*(?:\\.[^"\\]*)*$/)

          if (inString) {
            // 在字符串内，不提供补全
            callback(null, [])
            return
          }

          // 简单的JSON关键词补全
          const completions = [
            { name: 'true', value: 'true', score: 100, meta: 'boolean' },
            { name: 'false', value: 'false', score: 100, meta: 'boolean' },
            { name: 'null', value: 'null', score: 100, meta: 'null' }
          ]
          callback(null, completions)
        } catch (e) {
          callback(null, [])
        }
      }
    }

    langTools.addCompleter(jsonCompleter)
  }

  // 内容变化时验证JSON
  const onContentChange = () => {
    validateJson()
  }

  // 验证JSON
  const validateJson = () => {
    errorMessage.value = ''
    isValidJson.value = false

    if (!jsonContent.value.trim()) {
      resetStatistics()
      return
    }

    try {
      const parsed = JSON.parse(jsonContent.value)
      isValidJson.value = true

      // 计算统计信息
      calculateStatistics(parsed)
    } catch (error) {
      isValidJson.value = false

      // 格式化错误信息
      let errorMsg = error.message
      // 尝试从错误消息中提取行号和列号
      const match = errorMsg.match(/position (\d+)/)
      if (match) {
        const position = parseInt(match[1])
        const lines = jsonContent.value.substring(0, position).split('\n')
        const line = lines.length
        const column = lines[lines.length - 1].length + 1
        errorMsg = `JSON格式错误 (第${line}行, 第${column}列): ${errorMsg}`
      }

      errorMessage.value = errorMsg
      resetStatistics()
    }
  }

  // 重置统计信息
  const resetStatistics = () => {
    objectCount.value = 0
    keyValueCount.value = 0
    maxDepth.value = 0
  }

  // 计算JSON统计信息
  const calculateStatistics = (obj, depth = 1) => {
    if (!obj || typeof obj !== 'object') return

    // 更新最大深度
    if (depth > maxDepth.value) {
      maxDepth.value = depth
    }

    // 增加对象/数组计数
    objectCount.value++

    // 如果是对象，计算键值对数量
    if (obj.constructor === Object) {
      keyValueCount.value += Object.keys(obj).length

      // 递归计算嵌套对象
      for (const key in obj) {
        if (obj.hasOwnProperty(key)) {
          calculateStatistics(obj[key], depth + 1)
        }
      }
    }
    // 如果是数组，递归计算每个元素
    else if (Array.isArray(obj)) {
      for (const item of obj) {
        calculateStatistics(item, depth + 1)
      }
    }
  }

  // 格式化JSON
  const formatJson = () => {
    try {
      const parsed = JSON.parse(jsonContent.value)
      jsonContent.value = JSON.stringify(parsed, null, 2)
      isValidJson.value = true
      errorMessage.value = ''
      ElMessage({
        message: 'JSON已格式化',
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      ElMessage({
        message: '格式化失败，请检查JSON格式',
        type: 'error',
        grouping: true,
      })
    }
  }

  // 压缩JSON
  const compressJson = () => {
    try {
      const parsed = JSON.parse(jsonContent.value)
      jsonContent.value = JSON.stringify(parsed)
      isValidJson.value = true
      errorMessage.value = ''
      ElMessage({
        message: 'JSON已压缩',
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      ElMessage({
        message: '压缩失败，请检查JSON格式',
        type: 'error',
        grouping: true,
      })
    }
  }

  // 复制到剪贴板
  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(jsonContent.value)
      ElMessage({
        message: '已复制到剪贴板',
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      // 降级方案
      const textArea = document.createElement('textarea')
      textArea.value = jsonContent.value
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()

      try {
        document.execCommand('copy')
        ElMessage({
          message: '已复制到剪贴板',
          type: 'success',
          grouping: true,
        })
      } catch (fallbackError) {
        ElMessage({
          message: '复制失败，请手动复制',
          type: 'error',
          grouping: true,
        })
      }

      document.body.removeChild(textArea)
    }
  }

  // 下载JSON文件
  const downloadJson = () => {
    try {
      const blob = new Blob([jsonContent.value], { type: 'application/json;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = isValidJson.value ? 'formatted.json' : 'json-data.txt'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      ElMessage({
        message: '文件已下载',
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      ElMessage({
        message: '下载失败',
        type: 'error',
        grouping: true,
      })
    }
  }

  // 处理文件上传
  const handleFileUpload = (event) => {
    const file = event.target.files[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      jsonContent.value = e.target.result
      validateJson()
      ElMessage({
        message: `已加载文件: ${file.name}`,
        type: 'success',
        grouping: true,
      })
    }
    reader.onerror = () => {
      ElMessage({
        message: '文件读取失败',
        type: 'error',
        grouping: true,
      })
    }
    reader.readAsText(file)

    // 清空input，允许重复上传同一文件
    event.target.value = ''
  }

  // 设置示例数据
  const setSampleData = () => {
    const sampleData = {
      "name": "JSON编辑器示例",
      "version": "1.0.0",
      "features": [
        "格式化",
        "压缩",
        "验证",
        "文件上传下载",
        "智能编辑"
      ],
      "settings": {
        "theme": "monokai",
        "fontSize": 14,
        "tabSize": 2
      },
      "isActive": true,
      "lastUpdated": new Date().toISOString()
    }

    jsonContent.value = JSON.stringify(sampleData, null, 2)
    validateJson()
    ElMessage({
      message: '已加载示例数据',
      type: 'success',
      grouping: true,
    })
  }

  // 清空所有内容
  const clearAll = () => {
    jsonContent.value = ''
    errorMessage.value = ''
    isValidJson.value = false
    resetStatistics()
    ElMessage({
      message: '内容已清空',
      type: 'info',
      grouping: true,
    })
  }

  // 组件挂载时初始化
  onMounted(() => {
    // 可以在这里添加一些初始化逻辑
  })
</script>

<style scoped>
  .json-editor-tool {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .container {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 16px;
    background-color: var(--bg-container);
  }

  .header {
    text-align: center;
    margin-bottom: 16px;
  }

  .title {
    font-size: 24px;
    font-weight: bold;
    margin: 0 0 8px 0;
    color: var(--text-main);
  }

  .subtitle {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    margin-bottom: 16px;
    padding: 8px;
    background-color: var(--bg-muted);
    border-radius: 4px;
  }

  .tool-group {
    display: flex;
    gap: 8px;
  }

  .editor-container {
    flex: 1;
    overflow: hidden;
    border-radius: 4px;
    border: 1px solid var(--border-color-light);
    background-color: #272822;
  }

  .json-ace-editor {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  }

  .error-message {
    margin-top: 16px;
  }

  .statistics {
    display: flex;
    justify-content: space-around;
    margin-top: 12px;
    padding: 8px;
    background-color: var(--bg-muted);
    border-radius: 4px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .statistics span {
    padding: 4px 8px;
  }

  /* 自定义按钮样式 */
  .is-success {
    background-color: var(--color-success) !important;
    border-color: var(--color-success) !important;
  }

  .is-danger {
    background-color: var(--color-danger) !important;
    border-color: var(--color-danger) !important;
  }

  /* 响应式调整 */
  @media (max-width: 768px) {
    .toolbar {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }

    .tool-group {
      justify-content: center;
    }
  }
</style>