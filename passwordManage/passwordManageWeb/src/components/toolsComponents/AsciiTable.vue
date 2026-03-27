<template>
  <div class="ascii-table-container">
    <div class="ascii-table-header">
      <h3>ASCII码表</h3>
      <div class="search-container">
        <input v-model="searchTerm" @input="filterAsciiTable" placeholder="搜索字符、十进制或十六进制值..." class="search-input" />
        <select v-model="displayMode" @change="generateAsciiTable" class="display-mode-select">
          <option value="all">全部显示</option>
          <option value="printable">仅可打印字符</option>
          <option value="control">控制字符</option>
        </select>
        <select v-model="searchScope" @change="filterAsciiTable" class="search-scope-select">
          <option value="all">全部搜索</option>
          <option value="char">仅搜索字符</option>
          <option value="dec">仅搜索十进制</option>
          <option value="hex">仅搜索十六进制</option>
        </select>
      </div>
    </div>

    <div class="ascii-table-description">
      <p>ASCII（American Standard Code for Information Interchange）是基于拉丁字母的一套电脑编码系统，包含128个字符。</p>
      <p>下表显示了0-127范围内的所有ASCII字符及其对应的十进制、十六进制和二进制值。</p>
    </div>

    <div class="ascii-table-wrapper">
      <table class="ascii-table">
        <thead>
          <tr>
            <th>十进制</th>
            <th>十六进制</th>
            <th>二进制</th>
            <th>字符</th>
            <th>描述</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="char in filteredAsciiChars" :key="char.dec">
            <td class="dec-value">{{ char.dec }}</td>
            <td class="hex-value">0x{{ char.hex }}</td>
            <td class="bin-value">{{ char.bin }}</td>
            <td class="char-value" :class="{ 'control-char': char.isControl }">
              <span v-if="char.isControl">{{ char.display }}</span>
              <span v-else>{{ char.char }}</span>
            </td>
            <td class="desc-value">{{ char.description }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="filteredAsciiChars.length === 0" class="no-results">
      没有找到匹配的ASCII字符
    </div>

    <div class="ascii-table-footer">
      <p>提示：一个静态页面。</p>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted, computed } from 'vue'
  import { ElMessage } from 'element-plus'

  // 响应式数据
  const searchTerm = ref('')
  const displayMode = ref('all')
  const searchScope = ref('all')
  const asciiChars = ref([])
  const filteredAsciiChars = ref([])

  // ASCII控制字符描述映射
  const controlCharDescriptions = {
    0: 'NUL (Null character)',
    1: 'SOH (Start of Heading)',
    2: 'STX (Start of Text)',
    3: 'ETX (End of Text)',
    4: 'EOT (End of Transmission)',
    5: 'ENQ (Enquiry)',
    6: 'ACK (Acknowledgment)',
    7: 'BEL (Bell)',
    8: 'BS (Backspace)',
    9: 'HT (Horizontal Tab)',
    10: 'LF (Line Feed)',
    11: 'VT (Vertical Tab)',
    12: 'FF (Form Feed)',
    13: 'CR (Carriage Return)',
    14: 'SO (Shift Out)',
    15: 'SI (Shift In)',
    16: 'DLE (Data Link Escape)',
    17: 'DC1 (Device Control 1)',
    18: 'DC2 (Device Control 2)',
    19: 'DC3 (Device Control 3)',
    20: 'DC4 (Device Control 4)',
    21: 'NAK (Negative Acknowledgment)',
    22: 'SYN (Synchronous Idle)',
    23: 'ETB (End of Transmission Block)',
    24: 'CAN (Cancel)',
    25: 'EM (End of Medium)',
    26: 'SUB (Substitute)',
    27: 'ESC (Escape)',
    28: 'FS (File Separator)',
    29: 'GS (Group Separator)',
    30: 'RS (Record Separator)',
    31: 'US (Unit Separator)',
    127: 'DEL (Delete)'
  }

  // 生成ASCII字符表
  const generateAsciiTable = () => {
    const chars = []

    for (let i = 0; i <= 127; i++) {
      const isControl = i < 32 || i === 127
      const char = {
        dec: i,
        hex: i.toString(16).toUpperCase().padStart(2, '0'),
        bin: i.toString(2).padStart(8, '0'),
        char: isControl ? '' : String.fromCharCode(i),
        isControl: isControl,
        display: isControl ? getControlCharDisplay(i) : String.fromCharCode(i),
        description: isControl ? controlCharDescriptions[i] : `可打印字符"${String.fromCharCode(i)}"`
      }
      chars.push(char)
    }

    asciiChars.value = chars
    filterAsciiTable()
  }

  // 获取控制字符的显示表示
  const getControlCharDisplay = (code) => {
    const controlDisplay = {
      0: 'NUL',
      1: 'SOH',
      2: 'STX',
      3: 'ETX',
      4: 'EOT',
      5: 'ENQ',
      6: 'ACK',
      7: 'BEL',
      8: 'BS',
      9: 'HT',
      10: 'LF',
      11: 'VT',
      12: 'FF',
      13: 'CR',
      14: 'SO',
      15: 'SI',
      16: 'DLE',
      17: 'DC1',
      18: 'DC2',
      19: 'DC3',
      20: 'DC4',
      21: 'NAK',
      22: 'SYN',
      23: 'ETB',
      24: 'CAN',
      25: 'EM',
      26: 'SUB',
      27: 'ESC',
      28: 'FS',
      29: 'GS',
      30: 'RS',
      31: 'US',
      127: 'DEL'
    }
    return controlDisplay[code] || ''
  }

  // 过滤ASCII表
  const filterAsciiTable = () => {
    let filtered = [...asciiChars.value]

    // 应用显示模式过滤
    if (displayMode.value === 'printable') {
      filtered = filtered.filter(char => !char.isControl)
    } else if (displayMode.value === 'control') {
      filtered = filtered.filter(char => char.isControl)
    }

    // 应用搜索过滤
    if (searchTerm.value.trim()) {
      const term = searchTerm.value.trim().toLowerCase()

      if (searchScope.value === 'all') {
        // 全部搜索，但排除描述中可能包含的干扰词
        filtered = filtered.filter(char =>
          char.dec.toString().includes(term) ||
          char.hex.toLowerCase().includes(term) ||
          (!char.isControl && char.char && String(char.char).toLowerCase().includes(term))
        )
      } else if (searchScope.value === 'char') {
        // 仅搜索字符值
        filtered = filtered.filter(char =>
          !char.isControl && char.char && String(char.char).toLowerCase().includes(term)
        )
      } else if (searchScope.value === 'dec') {
        // 仅搜索十进制值
        filtered = filtered.filter(char =>
          char.dec.toString().includes(term)
        )
      } else if (searchScope.value === 'hex') {
        // 仅搜索十六进制值
        filtered = filtered.filter(char =>
          char.hex.toLowerCase().includes(term)
        )
      }
    }

    filteredAsciiChars.value = filtered
  }

  // 复制ASCII字符信息
  const copyToClipboard = (char) => {
    const text = `ASCII字符信息：\n字符：${char.isControl ? char.display : char.char}\n十进制：${char.dec}\n十六进制：0x${char.hex}\n二进制：${char.bin}\n描述：${char.description}`

    navigator.clipboard.writeText(text)
      .then(() => {
        ElMessage({
          message: 'ASCII字符信息已复制到剪贴板',
          type: 'success',
          grouping: true,
        })
      })
      .catch(err => {
        console.error('复制失败:', err)
        ElMessage({
          message: '复制失败，请手动复制',
          type: 'error',
          grouping: true,
        })
      })
  }

  // 组件挂载时生成ASCII表
  onMounted(() => {
    generateAsciiTable()
  })
</script>

<style scoped>
  /* @import './../../assets/theme.css';*/

  .ascii-table-container {
    width: 100%;
    height: 80vh;
    display: flex;
    flex-direction: column;
    background-color: var(--bg-container);
    border-radius: 4px;
    overflow: hidden;
  }

  .ascii-table-header {
    padding: 10px 15px;
    background-color: var(--bg-muted);
    border-bottom: 1px solid var(--border-color-light);
  }

  .ascii-table-header h3 {
    margin: 0 0 10px 0;
    color: var(--text-main);
    font-size: 18px;
  }

  .search-container {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .search-input {
    flex: 1;
    min-width: 200px;
    padding: 8px 12px;
    border: 1px solid var(--border-color-light);
    border-radius: 4px;
    font-size: 14px;
    background-color: var(--bg-container);
    color: var(--text-main);
  }

  .display-mode-select,
  .search-scope-select {
    padding: 8px 12px;
    border: 1px solid var(--border-color-light);
    border-radius: 4px;
    font-size: 14px;
    background-color: var(--bg-container);
    color: var(--text-main);
  }

  @media (max-width: 768px) {
    .search-scope-select {
      width: 100%;
    }
  }

  .ascii-table-description {
    padding: 10px 15px;
    background-color: var(--bg-hover);
    border-bottom: 1px solid var(--border-color-light);
    font-size: 13px;
    line-height: 1.5;
  }

  .ascii-table-description p {
    margin: 5px 0;
    color: var(--text-secondary);
  }

  .ascii-table-wrapper {
    flex: 1;
    overflow: auto;
    padding: 10px;
  }

  .ascii-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
    table-layout: fixed;
  }

  .ascii-table th {
    background-color: var(--bg-muted);
    color: var(--text-main);
    padding: 10px;
    text-align: left;
    border-bottom: 2px solid var(--border-color-light);
    position: sticky;
    top: 0;
    z-index: 10;
  }

  .ascii-table td {
    padding: 8px 10px;
    border-bottom: 1px solid var(--border-color-light);
    color: var(--text-main);
  }

  .ascii-table tr:hover {
    background-color: var(--bg-hover);
    cursor: pointer;
  }

  .dec-value {
    width: 80px;
    text-align: center;
    font-family: monospace;
  }

  .hex-value {
    width: 100px;
    text-align: center;
    font-family: monospace;
    color: var(--color-primary);
  }

  .bin-value {
    width: 120px;
    text-align: center;
    font-family: monospace;
    font-size: 12px;
  }

  .char-value {
    width: 60px;
    text-align: center;
    font-family: monospace;
    font-size: 16px;
    font-weight: bold;
  }

  .control-char {
    color: var(--color-danger);
    font-size: 12px;
  }

  .desc-value {
    flex: 1;
    word-break: break-word;
    color: var(--text-secondary);
  }

  .no-results {
    text-align: center;
    padding: 40px;
    color: var(--text-secondary);
    font-style: italic;
  }

  .ascii-table-footer {
    padding: 10px 15px;
    background-color: var(--bg-muted);
    border-top: 1px solid var(--border-color-light);
    font-size: 12px;
    color: var(--text-secondary);
    text-align: center;
  }

  /* 响应式布局 */
  @media (max-width: 768px) {
    .search-container {
      flex-direction: column;
    }

    .search-input,
    .display-mode-select {
      width: 100%;
    }

    .ascii-table {
      font-size: 12px;
    }

    .ascii-table th,
    .ascii-table td {
      padding: 6px 8px;
    }

    .bin-value {
      display: none;
    }

    .hex-value {
      width: 80px;
    }
  }

  @media (max-width: 480px) {
    .hex-value {
      display: none;
    }

    .dec-value {
      width: 60px;
    }
  }
</style>