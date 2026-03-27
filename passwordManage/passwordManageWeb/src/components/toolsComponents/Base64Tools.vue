<template>
  <div class="base64-tool">
    <h3>Base64 编码/解码</h3>

    <!-- 输入显示模式选择 -->
    <div class="display-mode-section">
      <label for="inputDisplayMode">输入格式:</label>
      <select id="inputDisplayMode" v-model="inputDisplayMode">
        <option value="text">文本</option>
        <option value="base64">Base64</option>
        <option value="hex">十六进制</option>
        <option value="binary">二进制</option>
      </select>
    </div>

    <!-- 原输出显示模式选择保持不变 -->
    <div class="display-mode-section">
      <label for="outputDisplayMode">输出显示模式:</label>
      <select id="outputDisplayMode" v-model="outputDisplayMode">
        <option value="text">文本</option>
        <option value="base64">Base64</option>
        <option value="hex">十六进制</option>
        <option value="binary">二进制</option>
      </select>
    </div>

    <!-- 输入区域 -->
    <div class="input-section">
      <textarea v-model="inputText" placeholder="请输入需要编码或解码的内容..." rows="5"></textarea>
    </div>

    <!-- 按钮操作 -->
    <div class="button-section">
      <button @click="encode">编码</button>
      <button @click="decode">解码</button>
      <button @click="clear">清空</button>
      <button @click="fillInputWithOutput">使用结果填充输入</button>
    </div>

    <!-- 输出区域 -->
    <div class="output-section">
      <!-- 原错误绑定：:value="formattedOutputText" -->
      <textarea :value="outputText" placeholder="结果将显示在此处..." rows="5" readonly></textarea>
    </div>
  </div>
</template>

<script setup>
  import { ref, watch } from 'vue'
  import { toByteArray, fromByteArray } from 'base64-js'
  import { ElMessage } from 'element-plus'

  // 显示文本（绑定输入输出框）
  const inputText = ref('')  // 输入框显示文本
  const outputText = ref('') // 输出框显示文本

  // 实际二进制数据
  const inputRaw = ref(new Uint8Array())  // 输入的原始二进制数据
  const outputRaw = ref(new Uint8Array()) // 输出的原始二进制数据

  // 格式模式
  const inputDisplayMode = ref('text')  // 输入显示格式
  const outputDisplayMode = ref('text') // 输出显示格式

  // 文本 <-> 二进制 
  const textToRaw = (text, mode) => {
    try {
      let bytes
      switch (mode) {
        case 'text': return new TextEncoder().encode(text)
        case 'base64': {
          const cleanText = text.trim().replace(/[^A-Za-z0-9+/]/g, '')
          const paddedText = cleanText.padEnd(cleanText.length + (4 - cleanText.length % 4) % 4, '=')
          // 返回Base64字符串的二进制形式
          return new TextEncoder().encode(paddedText)
        }
        case 'hex': {
          const cleanHex = text.replace(/[^0-9a-fA-F]/g, '')
          if (cleanHex.length % 2 !== 0) throw new Error('十六进制数据长度必须为偶数')
          return new Uint8Array(cleanHex.match(/.{2}/g).map(h => parseInt(h, 16)))
        }
        case 'binary': {
          const cleanBin = text.replace(/[^01]/g, '')
          if (cleanBin.length % 8 !== 0) throw new Error('二进制数据长度必须为8的倍数')
          return new Uint8Array(cleanBin.match(/.{8}/g).map(b => parseInt(b, 2)))
        }
        default: return new Uint8Array()
      }
    } catch (error) {
      ElMessage({
        message: `输入解析失败（${mode}格式）：${error.message}`,
        type: 'error',
        grouping: true,
      })
      return new Uint8Array()
    }
  }

  const rawToText = (raw, mode) => {
    try {
      switch (mode) {
        case 'text': return new TextDecoder().decode(raw)
        case 'base64': return fromByteArray(raw)
        case 'hex': return Array.from(raw).map(b => b.toString(16).padStart(2, '0')).join('')
        case 'binary': return Array.from(raw).map(b => b.toString(2).padStart(8, '0')).join(' ')
        default: return ''
      }
    } catch {
      return '[无效数据]'
    }
  }

  // 输出同步：输出二进制数据/格式变化 -> 更新显示文本
  watch([outputRaw, outputDisplayMode], () => {
    outputText.value = rawToText(outputRaw.value, outputDisplayMode.value)
  })

  // 编码操作：将输入二进制数据转换为Base64二进制数据
  const encode = () => {
    // 按当前输入格式解析输入内容
    const parsedInput = textToRaw(inputText.value, inputDisplayMode.value)
    if (parsedInput.length === 0) { // 解析失败
      ElMessage({
        message: '输入解析失败，请检查格式',
        type: 'error',
        grouping: true,
      })
      return
    }
    // 解析成功，执行编码
    const base64Str = fromByteArray(parsedInput)
    outputRaw.value = new TextEncoder().encode(base64Str)
  }

  // 解码操作
  const decode = () => {
    // 按当前输入格式解析输入内容
    const parsedInput = textToRaw(inputText.value, inputDisplayMode.value)
    if (parsedInput.length === 0) {
      ElMessage({
        message: '输入解析失败，请检查格式',
        type: 'error',
        grouping: true,
      })
      return
    }

    // 将解析后的二进制数据转换为字符串
    const base64Str = new TextDecoder().decode(parsedInput)

    // 尝试Base64解码
    try {
      outputRaw.value = toByteArray(base64Str)
    } catch (error) {
      ElMessage({
        message: 'Base64解码失败，请检查输入内容',
        type: 'error',
        grouping: true,
      })
      outputRaw.value = new Uint8Array()
    }
  }

  // 填充输入：将输出文本覆盖输入，并同步格式
  const fillInputWithOutput = () => {
    inputText.value = outputText.value
    inputDisplayMode.value = outputDisplayMode.value
  }

  // 清空操作
  const clear = () => {
    inputText.value = ''
    outputText.value = ''
    inputRaw.value = new Uint8Array()
    outputRaw.value = new Uint8Array()
  }
</script>

<style scoped>
  .base64-tool {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  textarea {
    width: 100%;
    padding: 10px;
    font-size: 14px;
    border: 1px solid #ccc;
    border-radius: 4px;
    resize: none;
  }

  .button-section {
    display: flex;
    gap: 10px;
    justify-content: flex-start;
  }

  button {
    padding: 10px 15px;
    font-size: 14px;
    border: none;
    background-color: #007bff;
    color: white;
    cursor: pointer;
    border-radius: 4px;
    transition: background-color 0.3s ease;
  }

  button:hover {
    background-color: #0056b3;
  }

  .output-section {
    display: flex;
    flex-direction: column;
  }
</style>