<template>
  <div class="formatter-config">
    <!-- 大小写转换 -->
    <div class="config-group">
      <label class="config-label">大小写转换</label>
      <div class="radio-group">
        <label class="radio-label">
          <input type="radio" name="caseTransform" v-model="config.caseTransform" value="none" @change="updateConfig" />
          不转换
        </label>
        <label class="radio-label">
          <input type="radio" name="caseTransform" v-model="config.caseTransform" value="allUpper"
            @change="updateConfig" />
          全部大写
        </label>
        <label class="radio-label">
          <input type="radio" name="caseTransform" v-model="config.caseTransform" value="allLower"
            @change="updateConfig" />
          全部小写
        </label>
        <label class="radio-label">
          <input type="radio" name="caseTransform" v-model="config.caseTransform" value="word" @change="updateConfig" />
          单词首字母大写
        </label>
        <label class="radio-label">
          <input type="radio" name="caseTransform" v-model="config.caseTransform" value="line" @change="updateConfig" />
          每行首字母大写
        </label>
      </div>
    </div>

    <!-- 空格处理 -->
    <div class="config-group">
      <label class="config-label">空格处理</label>
      <div class="checkbox-group">
        <label class="checkbox-label">
          <input type="checkbox" v-model="config.removeAllSpaces" @change="updateConfig" />
          清除全部空格
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="config.normalizeSpaces" @change="updateConfig" />
          合并连续空格（单行内）
        </label>
      </div>

      <div class="config-subgroup">
        <label class="config-label small">行首尾空格处理</label>
        <select v-model="config.trimLines" @change="updateConfig" class="form-select small">
          <option value="">不处理</option>
          <option value="left">去除左侧空格</option>
          <option value="right">去除右侧空格</option>
          <option value="both">去除两侧空格</option>
        </select>
      </div>
    </div>

    <!-- 空行处理 -->
    <div class="config-group">
      <label class="config-label">空行处理</label>
      <div class="config-row">
        <label class="checkbox-label" style="flex: 1;">
          <input type="checkbox" v-model="config.removeEmptyLines" @change="updateConfig" />
          清除空行
        </label>
        <div class="input-group" v-if="config.removeEmptyLines">
          <label class="config-label small">保留空行数量</label>
          <input type="number" v-model.number="config.keepEmptyLinesCount" @input="updateConfig" min="0"
            class="form-input small" />
        </div>
      </div>
    </div>

    <!-- 换行处理 -->
    <div class="config-group">
      <label class="config-label">换行处理</label>
      <div class="checkbox-group">
        <label class="checkbox-label">
          <input type="checkbox" v-model="config.removeLineBreaks" @change="updateConfig" />
          去除换行符
        </label>
      </div>

      <div class="config-subgroup">
        <label class="config-label small">换行符替换为</label>
        <select v-model="config.lineBreakReplace" @change="updateConfig" class="form-select small"
          v-if="config.removeLineBreaks">
          <option value="space">空格</option>
          <option value="empty">空字符串</option>
        </select>
      </div>
    </div>

    <!-- 字符处理 -->
    <div class="config-group">
      <label class="config-label">字符处理</label>
      <div class="checkbox-group">
        <label class="checkbox-label">
          <input type="checkbox" v-model="config.removeInvisibleChars" @change="updateConfig" />
          去除不可见字符
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="config.removeDuplicates" @change="updateConfig" />
          行去重
        </label>
      </div>
    </div>

    <!-- 删除字符选项 -->
    <div class="config-group">
      <label class="config-label">删除指定字符</label>
      <div class="config-row">
        <label class="checkbox-label" style="flex: 1;">
          <input type="checkbox" v-model="config.removeChars" @change="updateConfig" />
          启用删除字符功能
        </label>
      </div>

      <div class="remove-chars-options" v-if="config.removeChars">
        <div class="config-row">
          <div class="config-subgroup" style="flex: 1;">
            <label class="config-label small">匹配模式</label>
            <select v-model="config.matchMode" @change="updateConfig" class="form-select">
              <option value="string">字符串</option>
              <option value="regex">正则表达式</option>
            </select>
          </div>

          <div class="config-subgroup" style="flex: 2;">
            <label class="config-label small">要删除的内容</label>
            <input type="text" v-model="config.matchContent" @input="updateConfig" placeholder="输入要删除的字符或正则表达式"
              class="form-input" />
          </div>
        </div>

        <div class="checkbox-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="config.global" @change="updateConfig" />
            全局匹配
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="config.ignoreCase" @change="updateConfig" />
            忽略大小写
          </label>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import { ref, watch } from 'vue'

  export default {
    name: 'FormatterConfig',
    props: {
      modelValue: {
        type: Object,
        default: () => ({})
      }
    },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
      const config = ref({
        caseTransform: 'none',
        removeAllSpaces: false,
        normalizeSpaces: false,
        trimLines: '',
        removeEmptyLines: false,
        keepEmptyLinesCount: 0,
        removeLineBreaks: false,
        lineBreakReplace: 'space',
        removeDuplicates: false,
        removeChars: false,
        removeInvisibleChars: false,
        matchMode: 'string',
        matchContent: '',
        global: true,
        ignoreCase: false
      })

      // 初始化配置
      watch(() => props.modelValue, (newVal) => {
        if (newVal && Object.keys(newVal).length > 0) {
          config.value = { ...config.value, ...newVal }
        }
      }, { immediate: true })

      const updateConfig = () => {
        emit('update:modelValue', { ...config.value })
      }

      return {
        config,
        updateConfig
      }
    }
  }
</script>

<style scoped>
  .formatter-config {
    padding: 10px 0;
  }

  .config-group {
    margin-bottom: 15px;
  }

  .config-label {
    display: block;
    margin-bottom: 5px;
    font-weight: 500;
    color: #333;
    font-size: 14px;
  }

  .config-label.small {
    font-size: 12px;
    font-weight: normal;
    margin-bottom: 3px;
  }

  .form-select {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
    background: white;
  }

  .form-select.small {
    padding: 4px 8px;
    font-size: 12px;
  }

  .form-select:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  .checkbox-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: #333;
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
  }

  .radio-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 10px;
    padding-left: 5px;
  }

  .radio-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: #333;
    cursor: pointer;
  }

  .radio-label input[type="radio"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
  }

  /* 新增样式 */
  .config-row {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-bottom: 8px;
  }

  .config-subgroup {
    margin-bottom: 8px;
  }

  .input-group {
    display: flex;
    gap: 5px;
    align-items: center;
  }

  .form-input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
    background: white;
  }

  .form-input.small {
    padding: 4px 8px;
    font-size: 12px;
    width: auto;
    min-width: 60px;
  }

  .form-input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  .remove-chars-options {
    margin-top: 8px;
    padding-left: 10px;
    border-left: 1px solid #eee;
  }
</style>