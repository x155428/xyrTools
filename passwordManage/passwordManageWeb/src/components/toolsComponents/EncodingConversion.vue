<template>
  <div class="converter-container">
    <div class="converter-header">
      <h2>进制与字符编码转换器</h2>
      <p>支持多种进制转换和字符编码格式展示</p>
    </div>

    <!-- 标签导航 -->
    <div class="tab-navigation">
      <button :class="{ active: activeTab === 'base' }" @click="activeTab = 'base'" class="tab-button">
        <span class="tab-icon">🔢</span>
        进制转换
      </button>
      <button :class="{ active: activeTab === 'encoding' }" @click="activeTab = 'encoding'" class="tab-button">
        <span class="tab-icon">📝</span>
        字符编码
      </button>
    </div>

    <!-- 进制转换 -->
    <div v-if="activeTab === 'base'" class="tab-content">
      <div class="input-section">
        <div class="input-group">
          <label>输入数值</label>
          <input v-model="baseInput" @input="convertBaseRealtime" placeholder="请输入要转换的数值..." class="input-field" />
        </div>

        <div class="conversion-controls">
          <div class="base-selector">
            <label>源进制</label>
            <select v-model="fromBase" @change="convertBaseRealtime" class="select-field">
              <option value="2">二进制 (Binary)</option>
              <option value="8">八进制 (Octal)</option>
              <option value="10">十进制 (Decimal)</option>
              <option value="16">十六进制 (Hexadecimal)</option>
              <option value="32">32进制</option>
              <option value="36">36进制</option>
            </select>
          </div>

          <div class="arrow-icon">→</div>

          <div class="base-selector">
            <label>目标进制</label>
            <select v-model="toBase" @change="convertBaseRealtime" class="select-field">
              <option value="2">二进制 (Binary)</option>
              <option value="8">八进制 (Octal)</option>
              <option value="10">十进制 (Decimal)</option>
              <option value="16">十六进制 (Hexadecimal)</option>
              <option value="32">32进制</option>
              <option value="36">36进制</option>
            </select>
          </div>
        </div>
      </div>

      <div class="result-section">
        <h3>转换结果</h3>
        <div v-if="baseError" class="error-message">
          {{ baseError }}
        </div>
        <div v-else class="result-display">
          <div class="result-item">
            <span class="result-label">{{ getBaseLabel(toBase) }}:</span>
            <span class="result-value">{{ baseResult }}</span>
            <button @click="copyToClipboard(baseResult)" class="copy-btn" title="复制">📋</button>
          </div>
        </div>

        <!-- 所有进制展示 -->
        <div v-if="baseInput && !baseError" class="all-bases">
          <h4>所有进制表示</h4>
          <div class="bases-grid">
            <div v-for="base in allBases" :key="base.value" class="base-item">
              <span class="base-label">{{ base.label }}:</span>
              <span class="base-value">{{ convertToBase(baseInput, fromBase, base.value) }}</span>
              <button @click="copyToClipboard(convertToBase(baseInput, fromBase, base.value))"
                class="copy-btn-small">📋</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 字符编码转换 -->
    <div v-if="activeTab === 'encoding'" class="tab-content">
      <div class="input-section">
        <div class="input-group">
          <label>输入文本或编码数据</label>
          <textarea v-model="encodingInput" @input="convertEncodingRealtime" placeholder="输入文本或十六进制数据..."
            class="textarea-field" rows="4"></textarea>
        </div>

        <div class="encoding-controls">
          <div class="mode-selector">
            <label>转换模式</label>
            <select v-model="encodingMode" @change="convertEncodingRealtime" class="select-field">
              <option value="text-to-encoding">文本 → 编码</option>
              <option value="encoding-to-text">编码 → 文本</option>
            </select>
          </div>

          <div v-if="encodingMode === 'text-to-encoding'" class="encoding-selector">
            <label>源编码</label>
            <select v-model="sourceEncoding" @change="convertEncodingRealtime" class="select-field">
              <option value="utf-8">UTF-8</option>
              <option value="utf-16le">UTF-16 LE</option>
              <option value="utf-16be">UTF-16 BE</option>
              <option value="ascii">ASCII</option>
              <option value="latin1">Latin-1</option>
            </select>
          </div>
        </div>
      </div>

      <div class="result-section">
        <h3>编码结果展示</h3>
        <div v-if="encodingError" class="error-message">
          {{ encodingError }}
        </div>
        <div v-else class="encoding-results">
          <div v-if="encodingMode === 'text-to-encoding'" class="encoding-grid">
            <div v-for="(result, encoding) in encodingResults" :key="encoding" class="encoding-item">
              <div class="encoding-header">
                <span class="encoding-name">{{ getEncodingLabel(encoding) }}</span>
                <button @click="copyToClipboard(result)" class="copy-btn" title="复制">📋</button>
              </div>
              <div class="encoding-content">{{ result }}</div>
            </div>
          </div>

          <div v-else class="decoding-results">
            <div v-for="(result, encoding) in decodingResults" :key="encoding" class="encoding-item">
              <div class="encoding-header">
                <span class="encoding-name">{{ getEncodingLabel(encoding) }}</span>
                <button @click="copyToClipboard(result)" class="copy-btn" title="复制">📋</button>
              </div>
              <div class="encoding-content">{{ result }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 复制提示 -->
    <div v-if="copyMessage" class="copy-toast">
      {{ copyMessage }}
    </div>
  </div>
</template>

<script>
  import { ref, reactive } from 'vue'

  export default {
    name: 'ConverterComponent',
    setup() {
      const activeTab = ref('base')

      // 进制转换相关
      const baseInput = ref('')
      const fromBase = ref('10')
      const toBase = ref('16')
      const baseResult = ref('')
      const baseError = ref('')

      // 字符编码相关
      const encodingInput = ref('')
      const encodingMode = ref('text-to-encoding')
      const sourceEncoding = ref('utf-8')
      const encodingResults = reactive({})
      const decodingResults = reactive({})
      const encodingError = ref('')

      // 复制提示
      const copyMessage = ref('')

      // 进制配置
      const allBases = [
        { value: '2', label: '二进制' },
        { value: '8', label: '八进制' },
        { value: '10', label: '十进制' },
        { value: '16', label: '十六进制' },
        { value: '32', label: '32进制' },
        { value: '36', label: '36进制' }
      ]

      // 编码配置
      const encodings = ['ascii', 'utf-8', 'utf-16le', 'utf-16be', 'latin1', 'hex', 'binary']

      // 工具函数
      const getBaseLabel = (base) => {
        const labels = {
          '2': '二进制',
          '8': '八进制',
          '10': '十进制',
          '16': '十六进制',
          '32': '32进制',
          '36': '36进制'
        }
        return labels[base] || `${base}进制`
      }

      const getEncodingLabel = (encoding) => {
        const labels = {
          'utf-8': 'UTF-8',
          'utf-16le': 'UTF-16 LE',
          'utf-16be': 'UTF-16 BE',
          'ascii': 'ASCII',
          'latin1': 'Latin-1',
          'hex': '十六进制',
          'binary': '二进制'
        }
        return labels[encoding] || encoding.toUpperCase()
      }

      // 进制转换函数
      const convertToBase = (value, fromBase, toBase) => {
        try {
          if (!value.trim()) return ''

          // 清理输入值
          let cleanValue = value.trim()
          if (fromBase === '2' && cleanValue.startsWith('0b')) cleanValue = cleanValue.slice(2)
          if (fromBase === '8' && cleanValue.startsWith('0o')) cleanValue = cleanValue.slice(2)
          if (fromBase === '16' && cleanValue.startsWith('0x')) cleanValue = cleanValue.slice(2)

          // 验证输入
          const validChars = '0123456789abcdefghijklmnopqrstuvwxyz'.slice(0, parseInt(fromBase))
          if (!cleanValue.toLowerCase().split('').every(char => validChars.includes(char))) {
            return '无效输入'
          }

          const decimal = parseInt(cleanValue, parseInt(fromBase))
          if (isNaN(decimal)) return '转换失败'

          let result = decimal.toString(parseInt(toBase)).toUpperCase()

          // 添加前缀
          if (toBase === '2') result = '0b' + result
          if (toBase === '8') result = '0o' + result
          if (toBase === '16') result = '0x' + result

          return result
        } catch (error) {
          return '转换错误'
        }
      }

      const convertBaseRealtime = () => {
        baseError.value = ''

        if (!baseInput.value.trim()) {
          baseResult.value = ''
          return
        }

        try {
          const result = convertToBase(baseInput.value, fromBase.value, toBase.value)
          if (result.includes('无效') || result.includes('失败') || result.includes('错误')) {
            baseError.value = result
            baseResult.value = ''
          } else {
            baseResult.value = result
          }
        } catch (error) {
          baseError.value = '转换过程中发生错误'
          baseResult.value = ''
        }
      }

      // 字符编码转换函数
      const stringToBytes = (str, encoding) => {
        if (encoding === 'hex') {
          const hex = str.replace(/\s+/g, '').replace(/0x/g, '')
          const bytes = []
          for (let i = 0; i < hex.length; i += 2) {
            bytes.push(parseInt(hex.substr(i, 2), 16))
          }
          return new Uint8Array(bytes)
        }

        if (encoding === 'binary') {
          const binary = str.replace(/\s+/g, '')
          const bytes = []
          for (let i = 0; i < binary.length; i += 8) {
            bytes.push(parseInt(binary.substr(i, 8), 2))
          }
          return new Uint8Array(bytes)
        }

        const encoder = new TextEncoder()
        return encoder.encode(str)
      }

      const bytesToString = (bytes, encoding) => {
        if (encoding === 'hex') {
          return Array.from(bytes)
            .map(byte => byte.toString(16).padStart(2, '0').toUpperCase())
            .join(' ')
        }

        if (encoding === 'binary') {
          return Array.from(bytes)
            .map(byte => byte.toString(2).padStart(8, '0'))
            .join(' ')
        }

        try {
          const decoder = new TextDecoder(encoding)
          return decoder.decode(bytes)
        } catch {
          return '解码失败'
        }
      }

      const convertEncodingRealtime = () => {
        encodingError.value = ''

        if (!encodingInput.value.trim()) {
          Object.keys(encodingResults).forEach(key => delete encodingResults[key])
          Object.keys(decodingResults).forEach(key => delete decodingResults[key])
          return
        }

        try {
          if (encodingMode.value === 'text-to-encoding') {
            // 文本转编码
            const inputBytes = stringToBytes(encodingInput.value, sourceEncoding.value)

            encodings.forEach(encoding => {
              try {
                encodingResults[encoding] = bytesToString(inputBytes, encoding)
              } catch (error) {
                encodingResults[encoding] = '转换失败'
              }
            })
          } else {
            // 编码转文本
            encodings.forEach(encoding => {
              try {
                const bytes = stringToBytes(encodingInput.value, 'hex')
                decodingResults[encoding] = bytesToString(bytes, encoding)
              } catch (error) {
                decodingResults[encoding] = '解码失败'
              }
            })
          }
        } catch (error) {
          encodingError.value = '转换过程中发生错误'
        }
      }

      // 复制到剪贴板
      const copyToClipboard = async (text) => {
        try {
          await navigator.clipboard.writeText(text)
          copyMessage.value = '已复制到剪贴板'
          setTimeout(() => {
            copyMessage.value = ''
          }, 2000)
        } catch (error) {
          copyMessage.value = '复制失败'
          setTimeout(() => {
            copyMessage.value = ''
          }, 2000)
        }
      }

      return {
        // 响应式数据
        activeTab,
        baseInput,
        fromBase,
        toBase,
        baseResult,
        baseError,
        encodingInput,
        encodingMode,
        sourceEncoding,
        encodingResults,
        decodingResults,
        encodingError,
        copyMessage,
        allBases,

        // 方法
        getBaseLabel,
        getEncodingLabel,
        convertToBase,
        convertBaseRealtime,
        convertEncodingRealtime,
        copyToClipboard
      }
    }
  }
</script>

<style scoped>
  .converter-container {
    max-width: 100%;
    margin: 0 auto;
    padding: 20px;
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 16px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
    color: #333;
    max-height: 90vh;
    overflow-y: auto;
  }

  .converter-header {
    text-align: center;
    margin-bottom: 30px;
    color: white;
  }

  .converter-header h2 {
    margin: 0 0 10px 0;
    font-size: 28px;
    font-weight: 600;
  }

  .converter-header p {
    margin: 0;
    font-size: 16px;
    opacity: 0.9;
  }

  .tab-navigation {
    display: flex;
    gap: 8px;
    margin-bottom: 30px;
    background: rgba(255, 255, 255, 0.1);
    padding: 8px;
    border-radius: 12px;
    backdrop-filter: blur(10px);
  }

  .tab-button {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 20px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: white;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
  }

  .tab-button:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .tab-button.active {
    background: white;
    color: #667eea;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .tab-icon {
    font-size: 18px;
  }

  .tab-content {
    background: white;
    border-radius: 12px;
    padding: 30px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  }

  .input-section {
    margin-bottom: 30px;
  }

  .input-group {
    margin-bottom: 20px;
  }

  .input-group label {
    display: block;
    margin-bottom: 8px;
    font-weight: 600;
    color: #374151;
  }

  .input-field,
  .textarea-field,
  .select-field {
    width: 100%;
    padding: 12px 16px;
    border: 2px solid var(--border-color-light);
    border-radius: 8px;
    font-size: 16px;
    transition: all 0.3s ease;
    box-sizing: border-box;
  }

  .input-field:focus,
  .textarea-field:focus,
  .select-field:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }

  .textarea-field {
    resize: vertical;
    min-height: 100px;
    font-family: 'Courier New', monospace;
  }

  .conversion-controls {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    gap: 20px;
    align-items: end;
  }

  .encoding-controls {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
  }

  .base-selector,
  .mode-selector,
  .encoding-selector {
    display: flex;
    flex-direction: column;
  }

  .arrow-icon {
    font-size: 24px;
    color: #667eea;
    font-weight: bold;
    text-align: center;
    padding: 12px 0;
  }

  .result-section {
    border-top: 2px solid #f3f4f6;
    padding-top: 30px;
  }

  .result-section h3,
  .result-section h4 {
    margin: 0 0 20px 0;
    color: #374151;
    font-weight: 600;
  }

  .error-message {
    padding: 16px;
    background: #fef2f2;
    border: 1px solid var(--color-danger);
    border-radius: 8px;
    color: var(--color-danger);
    font-weight: 500;
  }

  .result-display {
    background: var(--bg-muted);
    border-radius: 8px;
    padding: 20px;
    border: 1px solid var(--border-color-light);
  }

  .result-item {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .result-label {
    font-weight: 600;
    color: var(--text-secondary);
    min-width: 100px;
  }

  .result-value {
    flex: 1;
    font-family: 'Courier New', monospace;
    font-size: 16px;
    color: var(--text-main);
    background: var(--bg-container);
    padding: 8px 12px;
    border-radius: 6px;
    border: 1px solid var(--border-color-light);
  }

  .all-bases {
    margin-top: 30px;
  }

  .bases-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 16px;
  }

  .base-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: var(--bg-muted);
    border-radius: 8px;
    border: 1px solid var(--border-color-light);
  }

  .base-label {
    font-weight: 600;
    color: var(--text-secondary);
    min-width: 80px;
  }

  .base-value {
    flex: 1;
    font-family: 'Courier New', monospace;
    color: var(--text-main);
  }

  .encoding-results {
    display: grid;
    gap: 16px;
  }

  .encoding-grid {
    display: grid;
    gap: 16px;
  }

  .encoding-item {
    background: var(--bg-muted);
    border-radius: 8px;
    border: 1px solid var(--border-color-light);
    overflow: hidden;
  }

  .encoding-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: var(--bg-hover);
    border-bottom: 1px solid var(--border-color-light);
  }

  .encoding-name {
    font-weight: 600;
    color: var(--text-main);
  }

  .encoding-content {
    padding: 16px;
    font-family: 'Courier New', monospace;
    font-size: 14px;
    line-height: 1.6;
    color: var(--text-main);
    word-break: break-all;
    white-space: pre-wrap;
  }

  .copy-btn,
  .copy-btn-small {
    padding: 6px 10px;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.3s ease;
  }

  .copy-btn-small {
    padding: 4px 8px;
    font-size: 12px;
  }

  .copy-btn:hover,
  .copy-btn-small:hover {
    background: var(--color-primary-hover);
    transform: translateY(-1px);
  }

  .copy-toast {
    position: fixed;
    top: 20px;
    right: 20px;
    background: var(--color-success);
    color: white;
    padding: 12px 20px;
    border-radius: 8px;
    box-shadow: var(--shadow-md);
    z-index: 1000;
    animation: slideIn 0.3s ease;
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

  @media (max-width: 768px) {
    .converter-container {
      padding: 16px;
      margin: 10px;
    }

    .conversion-controls {
      grid-template-columns: 1fr;
      gap: 16px;
    }

    .arrow-icon {
      transform: rotate(90deg);
    }

    .encoding-controls {
      grid-template-columns: 1fr;
    }

    .bases-grid {
      grid-template-columns: 1fr;
    }

    .base-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }

    .result-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }
  }
</style>