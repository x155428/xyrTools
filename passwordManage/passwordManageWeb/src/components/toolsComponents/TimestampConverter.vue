<template>
  <div class="timestamp-converter-container">
    <div class="converter-header">
      <h2>时间戳互转工具</h2>
      <p>支持时间戳与日期时间的相互转换</p>
    </div>

    <!-- 标签导航 -->
    <div class="tab-navigation">
      <button :class="{ active: activeTab === 'timestampToDate' }" @click="activeTab = 'timestampToDate'"
        class="tab-button">
        <span class="tab-icon">🔢</span>
        时间戳 → 日期
      </button>
      <button :class="{ active: activeTab === 'dateToTimestamp' }" @click="activeTab = 'dateToTimestamp'"
        class="tab-button">
        <span class="tab-icon">📅</span>
        日期 → 时间戳
      </button>
    </div>

    <!-- 时间戳转日期 -->
    <div v-if="activeTab === 'timestampToDate'" class="tab-content">
      <div class="input-section">
        <div class="input-group">
          <label>输入时间戳</label>
          <div class="input-with-buttons">
            <input v-model="timestampInput" @input="convertTimestampToDate" placeholder="请输入Unix时间戳..."
              class="input-field" />
            <button @click="getCurrentTimestamp" class="quick-btn" title="获取当前时间戳">当前时间戳</button>
          </div>
          <div class="format-options">
            <label>时间戳格式</label>
            <div class="radio-group">
              <label class="radio-option">
                <input type="radio" v-model="timestampFormat" value="seconds" @change="convertTimestampToDate" />
                秒级 (10位)
              </label>
              <label class="radio-option">
                <input type="radio" v-model="timestampFormat" value="milliseconds" @change="convertTimestampToDate" />
                毫秒级 (13位)
              </label>
            </div>
          </div>
          <div class="format-options">
            <label>选择时区</label>
            <div class="timezone-selector">
              <select v-model="timeZone" @change="convertTimestampToDate" class="timezone-select">
                <option v-for="tz in timeZones" :key="tz.value" :value="tz.value">{{ tz.label }}</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div class="result-section">
        <h3>转换结果</h3>
        <div v-if="timestampError" class="error-message">
          {{ timestampError }}
        </div>
        <div v-else-if="dateResult" class="result-display">
          <div class="result-item">
            <span class="result-label">标准格式:</span>
            <span class="result-value">{{ dateResult.standard }}</span>
            <button @click="copyToClipboard(dateResult.standard)" class="copy-btn" title="复制">📋</button>
          </div>
          <div class="result-item">
            <span class="result-label">详细格式:</span>
            <span class="result-value">{{ dateResult.detail }}</span>
            <button @click="copyToClipboard(dateResult.detail)" class="copy-btn" title="复制">📋</button>
          </div>
          <div class="result-item">
            <span class="result-label">ISO格式:</span>
            <span class="result-value">{{ dateResult.iso }}</span>
            <button @click="copyToClipboard(dateResult.iso)" class="copy-btn" title="复制">📋</button>
          </div>
          <div class="result-item">
            <span class="result-label">相对时间:</span>
            <span class="result-value">{{ dateResult.relative }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 日期转时间戳 -->
    <div v-if="activeTab === 'dateToTimestamp'" class="tab-content">
      <div class="input-section">
        <div class="input-group">
          <label>选择日期时间</label>
          <input type="datetime-local" v-model="dateInput" @input="convertDateToTimestamp" class="input-field" />
          <div class="quick-buttons">
            <button @click="setCurrentDateTime" class="quick-btn" title="设置为当前时间">当前时间</button>
          </div>
          <div class="format-options">
            <label>输出时间戳格式</label>
            <div class="radio-group">
              <label class="radio-option">
                <input type="radio" v-model="outputTimestampFormat" value="seconds" @change="convertDateToTimestamp" />
                秒级 (10位)
              </label>
              <label class="radio-option">
                <input type="radio" v-model="outputTimestampFormat" value="milliseconds"
                  @change="convertDateToTimestamp" />
                毫秒级 (13位)
              </label>
            </div>
          </div>
          <div class="format-options">
            <label>选择时区</label>
            <div class="timezone-selector">
              <select v-model="timeZone" @change="convertDateToTimestamp" class="timezone-select">
                <option v-for="tz in timeZones" :key="tz.value" :value="tz.value">{{ tz.label }}</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div class="result-section">
        <h3>转换结果</h3>
        <div v-if="dateError" class="error-message">
          {{ dateError }}
        </div>
        <div v-else-if="timestampResult" class="result-display">
          <div class="result-item">
            <span class="result-label">{{ outputTimestampFormat === 'seconds' ? '秒级时间戳' : '毫秒级时间戳' }}:</span>
            <span class="result-value">{{ timestampResult }}</span>
            <button @click="copyToClipboard(timestampResult)" class="copy-btn" title="复制">📋</button>
          </div>
          <div v-if="outputTimestampFormat === 'seconds'" class="result-item">
            <span class="result-label">毫秒级时间戳:</span>
            <span class="result-value">{{ timestampResult * 1000 }}</span>
            <button @click="copyToClipboard(timestampResult * 1000)" class="copy-btn" title="复制">📋</button>
          </div>
          <div v-else class="result-item">
            <span class="result-label">秒级时间戳:</span>
            <span class="result-value">{{ Math.floor(timestampResult / 1000) }}</span>
            <button @click="copyToClipboard(Math.floor(timestampResult / 1000))" class="copy-btn" title="复制">📋</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref } from 'vue'

  // 活动标签
  const activeTab = ref('timestampToDate')

  // 时间戳转日期相关
  const timestampInput = ref('')
  const timestampFormat = ref('seconds') // seconds 或 milliseconds
  const dateResult = ref(null)
  const timestampError = ref('')

  // 日期转时间戳相关
  const dateInput = ref('')
  const outputTimestampFormat = ref('seconds')
  const timestampResult = ref(null)
  const dateError = ref('')

  // 时区设置，默认东八区(UTC+8)
  const timeZone = ref('8')
  // 时区选项（UTC-12到UTC+12）
  const timeZones = Array.from({ length: 25 }, (_, i) => {
    const offset = i - 12
    // 格式化显示为UTC±HH:00格式
    const sign = offset >= 0 ? '+' : ''
    const hours = Math.abs(offset)
    const display = `UTC${sign}${hours.toString().padStart(2, '0')}:00`
    return { value: offset.toString(), label: display }
  })

  // 时间戳转日期
  const convertTimestampToDate = () => {
    if (!timestampInput.value) {
      dateResult.value = null
      timestampError.value = ''
      return
    }

    // 验证时间戳格式
    const timestampPattern = /^\d+$/;
    if (!timestampPattern.test(timestampInput.value)) {
      timestampError.value = '请输入有效的数字时间戳'
      dateResult.value = null
      return
    }

    try {
      let timestamp = parseInt(timestampInput.value)

      // 如果是秒级时间戳但长度小于10位，可能是需要补零的情况
      if (timestampFormat.value === 'seconds' && timestampInput.value.length < 10) {
        // 补零到10位
        const paddedTimestamp = timestampInput.value.padEnd(10, '0')
        timestamp = parseInt(paddedTimestamp)
      }

      // 如果是秒级时间戳，转换为毫秒级
      if (timestampFormat.value === 'seconds') {
        timestamp = timestamp * 1000
      }

      const date = new Date(timestamp)

      // 检查日期是否有效
      if (isNaN(date.getTime())) {
        timestampError.value = '无效的时间戳'
        dateResult.value = null
        return
      }

      // 格式化日期（使用用户选择的时区）
      const standard = formatDate(date, 'yyyy-MM-dd HH:mm:ss')
      const detail = formatDate(date, 'yyyy年MM月dd日 HH时mm分ss秒')
      const iso = date.toISOString()
      const relative = getRelativeTime(date)

      dateResult.value = {
        standard,
        detail,
        iso,
        relative
      }

      timestampError.value = ''
    } catch (error) {
      timestampError.value = '转换失败: ' + error.message
      dateResult.value = null
    }
  }

  // 日期转时间戳
  const convertDateToTimestamp = () => {
    if (!dateInput.value) {
      timestampResult.value = null
      dateError.value = ''
      return
    }

    try {
      // 用户输入的datetime-local是本地时间
      const localDate = new Date(dateInput.value)

      // 获取UTC时间（减去浏览器本地时区偏移）
      const utcTime = localDate.getTime() - (localDate.getTimezoneOffset() * 60 * 1000)

      // 现在我们有了UTC时间，需要根据用户选择的时区进行调整
      // 例如：如果用户选择UTC+8，而输入的是本地时间，我们需要将UTC时间转换为用户选择的时区对应的UTC时间
      const userTimezoneOffset = parseInt(timeZone.value) * 60 * 60 * 1000
      // 浏览器本地时间的时区偏移（分钟）转换为毫秒
      const browserTimezoneOffset = localDate.getTimezoneOffset() * 60 * 1000

      // 计算用户选择的时区下的UTC时间
      // 公式：目标UTC时间 = 本地时间的UTC时间 - 浏览器时区偏移 + 用户选择的时区偏移
      const targetUtcTime = utcTime - browserTimezoneOffset + userTimezoneOffset
      const date = new Date(targetUtcTime)

      // 检查日期是否有效
      if (isNaN(date.getTime())) {
        dateError.value = '无效的日期时间'
        timestampResult.value = null
        return
      }

      // 根据选择的格式返回时间戳
      const timestamp = outputTimestampFormat.value === 'seconds'
        ? Math.floor(date.getTime() / 1000)
        : date.getTime()

      timestampResult.value = timestamp
      dateError.value = ''
    } catch (error) {
      dateError.value = '转换失败: ' + error.message
      timestampResult.value = null
    }
  }

  // 获取当前时间戳（考虑用户选择的时区）
  const getCurrentTimestamp = () => {
    // 获取当前的UTC时间戳
    const timestamp = timestampFormat.value === 'seconds'
      ? Math.floor(Date.now() / 1000)
      : Date.now()

    timestampInput.value = timestamp.toString()
    convertTimestampToDate()
  }

  // 设置当前日期时间（考虑用户选择的时区）
  const setCurrentDateTime = () => {
    const now = new Date()

    // 获取UTC时间，并根据用户选择的时区进行调整
    const nowUtc = now.getTime() - (now.getTimezoneOffset() * 60 * 1000)
    const userTimezoneOffset = parseInt(timeZone.value) * 60 * 60 * 1000
    const nowInUserTimezone = new Date(nowUtc + userTimezoneOffset)

    // 格式化为YYYY-MM-DDTHH:MM格式
    const formattedDate = nowInUserTimezone.toISOString().slice(0, 16)
    dateInput.value = formattedDate
    convertDateToTimestamp()
  }

  // 格式化日期（使用用户选择的时区）
  const formatDate = (date, format) => {
    // 获取UTC时间，并添加用户选择的时区偏移量
    const utcTimestamp = date.getTime()
    const timezoneOffsetHours = parseInt(timeZone.value)
    // 创建指定时区的时间对象
    const timezoneDate = new Date(utcTimestamp + timezoneOffsetHours * 60 * 60 * 1000)

    const year = timezoneDate.getUTCFullYear()
    const month = String(timezoneDate.getUTCMonth() + 1).padStart(2, '0')
    const day = String(timezoneDate.getUTCDate()).padStart(2, '0')
    const hours = String(timezoneDate.getUTCHours()).padStart(2, '0')
    const minutes = String(timezoneDate.getUTCMinutes()).padStart(2, '0')
    const seconds = String(timezoneDate.getUTCSeconds()).padStart(2, '0')

    return format
      .replace('yyyy', year)
      .replace('MM', month)
      .replace('dd', day)
      .replace('HH', hours)
      .replace('mm', minutes)
      .replace('ss', seconds)
  }

  // 获取相对时间（考虑用户选择的时区）
  const getRelativeTime = (date) => {
    // 获取当前时间，并根据用户选择的时区进行调整
    const now = new Date()
    const nowUtc = now.getTime() - (now.getTimezoneOffset() * 60 * 1000)
    const userTimezoneOffset = parseInt(timeZone.value) * 60 * 60 * 1000
    const nowInUserTimezone = new Date(nowUtc + userTimezoneOffset)

    // 计算两个时间的差值（毫秒）
    const diff = nowInUserTimezone.getTime() - date.getTime()

    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    const months = Math.floor(days / 30)
    const years = Math.floor(days / 365)

    if (years > 0) return `${years}年前`
    if (months > 0) return `${months}个月前`
    if (days > 0) return `${days}天前`
    if (hours > 0) return `${hours}小时前`
    if (minutes > 0) return `${minutes}分钟前`
    if (seconds > 0) return `${seconds}秒前`
    return '刚刚'
  }

  // 复制到剪贴板
  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text.toString())
      // 可以添加一个临时提示，但为了简化，这里不添加
    } catch (err) {
      console.error('复制失败:', err)
    }
  }
</script>

<style scoped>
  .timestamp-converter-container {
    padding: 20px;
    background: #f8f9fa;
    border-radius: 8px;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .converter-header {
    text-align: center;
    margin-bottom: 20px;
  }

  .converter-header h2 {
    margin: 0 0 8px 0;
    color: #333;
  }

  .converter-header p {
    margin: 0;
    color: #666;
    font-size: 14px;
  }

  .tab-navigation {
    display: flex;
    margin-bottom: 20px;
    border-bottom: 1px solid #ddd;
  }

  .tab-button {
    flex: 1;
    padding: 12px 16px;
    background: #fff;
    border: 1px solid #ddd;
    border-bottom: none;
    cursor: pointer;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.3s;
  }

  .tab-button:first-child {
    border-radius: 6px 0 0 0;
  }

  .tab-button:last-child {
    border-radius: 0 6px 0 0;
  }

  .tab-button.active {
    background: #1890ff;
    color: white;
    border-color: #1890ff;
  }

  .tab-icon {
    font-size: 16px;
  }

  .tab-content {
    flex: 1;
    overflow-y: auto;
    background: #fff;
    padding: 20px;
    border-radius: 0 0 6px 6px;
    border: 1px solid #ddd;
    border-top: none;
  }

  .input-section {
    margin-bottom: 20px;
  }

  .input-group {
    margin-bottom: 16px;
  }

  .input-group label {
    display: block;
    margin-bottom: 6px;
    font-weight: 500;
    color: #333;
  }

  .input-field {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
  }

  .input-with-buttons {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .input-with-buttons .input-field {
    flex: 1;
  }

  .quick-btn {
    padding: 8px 16px;
    background: #f0f0f0;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
    transition: background-color 0.3s;
  }

  .quick-btn:hover {
    background: #e0e0e0;
  }

  .quick-buttons {
    margin-top: 10px;
    display: flex;
    gap: 10px;
  }

  .format-options {
    margin-top: 10px;
  }

  .timezone-selector {
    margin-top: 6px;
  }

  .timezone-select {
    width: 100%;
    padding: 6px 10px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
  }

  .radio-group {
    display: flex;
    gap: 20px;
    margin-top: 6px;
  }

  .radio-option {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
  }

  .result-section {
    margin-top: 20px;
  }

  .result-section h3 {
    margin: 0 0 12px 0;
    color: #333;
    font-size: 16px;
  }

  .error-message {
    padding: 12px;
    background: #fff2f0;
    border: 1px solid #ffccc7;
    border-radius: 4px;
    color: #f5222d;
  }

  .result-display {
    background: #f8f9fa;
    border: 1px solid #e9ecef;
    border-radius: 4px;
    padding: 16px;
  }

  .result-item {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e9ecef;
  }

  .result-item:last-child {
    margin-bottom: 0;
    padding-bottom: 0;
    border-bottom: none;
  }

  .result-label {
    font-weight: 500;
    color: #666;
    min-width: 100px;
  }

  .result-value {
    flex: 1;
    color: #333;
    word-break: break-all;
  }

  .copy-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: 2px;
    transition: background-color 0.3s;
  }

  .copy-btn:hover {
    background: rgba(0, 0, 0, 0.1);
  }
</style>