<template>
  <div class="url-tool">
    <h3>URL 编码 / 解码工具（多模式）</h3>

    <!-- 输入 -->
    <div class="input-section">
      <textarea v-model="inputText" placeholder="请输入需要编码或解码的文本..." rows="5"></textarea>
    </div>

    <!-- 编码模式 -->
    <div class="encode-mode">
      <label for="encodeMode">编/解码方式：</label>
      <select id="encodeMode" v-model="encodeMode">
        <option value="standard">标准</option>
        <option value="form">Form-urlencoded（空格<-> +）</option>
        <option value="full">全部（包括保留字符）</option>
      </select>
    </div>

    <!-- 按钮 -->
    <div class="button-section">
      <button @click="encode">编码</button>
      <button @click="decode">解码</button>
      <button @click="clear">清空</button>
    </div>

    <!-- 输出 -->
    <div class="output-section">
      <textarea :value="outputText" placeholder="结果将显示在此处..." rows="5" readonly></textarea>
    </div>
  </div>
</template>


<script setup>
  import { ref } from 'vue'

  const inputText = ref('')
  const outputText = ref('')
  const encodeMode = ref('standard')

  // 编码函数
  const encode = () => {
    try {
      const text = inputText.value.trim()
      if (!text) {
        outputText.value = ''
        return
      }

      let result = ''
      switch (encodeMode.value) {
        case 'standard':
          result = encodeURIComponent(text)
          break
        case 'form':
          result = new URLSearchParams({ v: text }).toString().slice(2)
          break
        case 'full':
          const encoder = new TextEncoder()
          const bytes = encoder.encode(text)
          result = [...bytes].map(b => '%' + b.toString(16).toUpperCase().padStart(2, '0')).join('')
          break
        default:
          result = text
      }

      outputText.value = result
    } catch (e) {
      outputText.value = '编码失败：' + e.message
    }
  }

  // 解码函数
  const decode = () => {
    try {
      let text = inputText.value.trim()
      if (!text) {
        outputText.value = ''
        return
      }

      switch (encodeMode.value) {
        case 'form':
          // 替换 + 为 空格
          text = text.replace(/\+/g, ' ')
          break
        case 'full':
          break
      }

      outputText.value = decodeURIComponent(text)
    } catch (e) {
      outputText.value = '解码失败：' + e.message
    }
  }

  // 清空
  const clear = () => {
    inputText.value = ''
    outputText.value = ''
  }
</script>


<style scoped>
  .url-tool {
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

  .encode-mode {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  select {
    padding: 5px 10px;
    font-size: 14px;
    border: 1px solid #ccc;
    border-radius: 4px;
  }

  .output-section {
    display: flex;
    flex-direction: column;
  }
</style>