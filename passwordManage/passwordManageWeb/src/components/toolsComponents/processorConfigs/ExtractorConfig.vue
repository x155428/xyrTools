<template>
  <div class="extractor-config">
    <div class="config-group">
      <label class="config-label">提取模式</label>
      <select v-model="config.patternType" @change="onPatternTypeChange" class="form-select">
        <option value="custom">自定义正则</option>
        <option value="ip">IP地址</option>
        <option value="ipv4">IPv4地址</option>
        <option value="ipv6">IPv6地址</option>
        <option value="url">URL</option>
        <option value="email">邮箱地址</option>
        <option value="phone">手机号码</option>
        <option value="number">数字</option>
        <option value="hex">十六进制</option>
      </select>
    </div>

    <div class="config-group">
      <label class="config-label">正则表达式</label>
      <input 
        v-model="config.pattern" 
        @input="updateConfig" 
        placeholder="输入正则表达式" 
        class="form-input"
      />
      <small class="help-text">使用正则表达式提取数据，注意转义字符</small>
    </div>

    <div class="config-group">
      <label class="config-label">提取选项</label>
      <div class="checkbox-group">
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.globalMatch" 
            @change="updateConfig"
          />
          全局匹配（提取所有匹配项）
        </label>
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.ignoreCase" 
            @change="updateConfig"
          />
          忽略大小写
        </label>
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.distinct" 
            @change="updateConfig"
          />
          去重（相同的匹配项只保留一个）
        </label>
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.showLineNumbers" 
            @change="updateConfig"
          />
          显示行号
        </label>
      </div>
    </div>

    <div class="config-group">
      <label class="config-label">捕获组选择</label>
      <select v-model="config.captureGroup" @change="updateConfig" class="form-select">
        <option value="all">全部捕获组</option>
        <option value="0">完整匹配</option>
        <option value="1">第一个捕获组</option>
        <option value="2">第二个捕获组</option>
        <option value="3">第三个捕获组</option>
      </select>
      <small class="help-text">选择要提取的捕获组，默认为完整匹配</small>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'

export default {
  name: 'ExtractorConfig',
  props: {
    modelValue: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    // 预设的正则表达式模板
    const patternTemplates = {
      ip: '((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}|([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}',
      ipv4: '(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)',
      ipv6: '([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}',
      url: 'https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&\\/\\/=]*)',
      email: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}',
      phone: '1[3-9]\\d{9}',
      number: '-?\\d+(\\.\\d+)?',
      hex: '#?([a-fA-F0-9]{6}|[a-fA-F0-9]{3})'
    }

    const config = ref({
      patternType: 'custom',
      pattern: '',
      globalMatch: true,
      ignoreCase: false,
      distinct: false,
      showLineNumbers: false,
      captureGroup: '0'
    })

    // 初始化配置
    watch(() => props.modelValue, (newVal) => {
      if (newVal && Object.keys(newVal).length > 0) {
        config.value = { ...config.value, ...newVal }
        // 如果有预设的patternType，则更新pattern
        if (newVal.patternType && patternTemplates[newVal.patternType]) {
          config.value.pattern = patternTemplates[newVal.patternType]
        }
      }
    }, { immediate: true })

    const updateConfig = () => {
      emit('update:modelValue', { ...config.value })
    }

    // 当选择预设模式时，更新正则表达式
    const onPatternTypeChange = () => {
      if (config.value.patternType !== 'custom' && patternTemplates[config.value.patternType]) {
        config.value.pattern = patternTemplates[config.value.patternType]
      }
      updateConfig()
    }

    return {
      config,
      updateConfig,
      onPatternTypeChange
    }
  }
}
</script>

<style scoped>
.extractor-config {
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

.form-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background: white;
}

.form-select:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  font-family: monospace;
  background: white;
}

.form-input:focus {
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

.help-text {
  display: block;
  margin-top: 5px;
  font-size: 12px;
  color: #666;
}
</style>