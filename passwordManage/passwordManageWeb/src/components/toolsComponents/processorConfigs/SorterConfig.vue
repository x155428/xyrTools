<template>
  <div class="sorter-config">
    <div class="config-group">
      <label class="config-label">排序模式</label>
      <select v-model="config.sortMode" @change="updateConfig" class="form-select">
        <option value="dictionary">字典序</option>
        <option value="numeric">数值</option>
        <option value="mixed">混合（优先数值）</option>
      </select>
    </div>

    <div class="config-group">
      <label class="config-label">字段分隔符</label>
      <input 
        v-model="config.fieldSeparator" 
        @input="updateConfig"
        placeholder="分隔字段的字符，默认空格"
        class="form-input"
      />
    </div>

    <div class="config-group">
      <label class="config-label">排序字段</label>
      <input 
        v-model="sortFieldsInput" 
        @input="updateSortFields"
        placeholder="输入字段编号，如：1 或 2,1（主次顺序）"
        class="form-input"
      />
      <small class="help-text">字段从1开始，多个字段用逗号分隔</small>
    </div>

    <div class="config-group">
      <label class="config-label">排序方向</label>
      <div class="radio-group">
        <label class="radio-label">
          <input 
            type="radio" 
            v-model="config.ascending" 
            :value="true"
            @change="updateConfig"
          />
          升序
        </label>
        <label class="radio-label">
          <input 
            type="radio" 
            v-model="config.ascending" 
            :value="false"
            @change="updateConfig"
          />
          降序
        </label>
      </div>
    </div>

    <div class="config-group">
      <label class="config-label">排序选项</label>
      <div class="checkbox-group">
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
            v-model="config.stable" 
            @change="updateConfig"
          />
          稳定排序（保留原始顺序）
        </label>
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.trimFields" 
            @change="updateConfig"
          />
          去除字段前空格
        </label>
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.removeDuplicates" 
            @change="updateConfig"
          />
          去重（按排序字段）
        </label>
      </div>
    </div>

    <div class="config-group">
      <label class="config-label">空值处理</label>
      <select v-model="config.emptyHandling" @change="updateConfig" class="form-select">
        <option value="first">空值在前</option>
        <option value="last">空值在后</option>
        <option value="ignore">忽略空行</option>
      </select>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'

export default {
  name: 'SorterConfig',
  props: {
    modelValue: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const config = ref({
      sortMode: 'dictionary',
      fieldSeparator: ' ',
      sortFields: [1],
      ascending: true,
      ignoreCase: false,
      stable: false,
      trimFields: true,
      removeDuplicates: false,
      emptyHandling: 'last'
    })

    const sortFieldsInput = ref('1')

    const updateConfig = () => {
      emit('update:modelValue', { ...config.value })
    }

    const updateSortFieldsInput = () => {
      if (config.value.sortFields && config.value.sortFields.length > 0) {
        sortFieldsInput.value = config.value.sortFields.join(',')
      }
    }

    // 初始化配置
    watch(() => props.modelValue, (newVal) => {
      if (newVal && Object.keys(newVal).length > 0) {
        config.value = { ...config.value, ...newVal }
        updateSortFieldsInput()
      }
    }, { immediate: true })

    const updateSortFields = () => {
      const fields = sortFieldsInput.value
        .split(',')
        .map(f => Number(f.trim()))
        .filter(f => !isNaN(f))
      
      config.value.sortFields = fields
      updateConfig()
    }

    return {
      config,
      sortFieldsInput,
      updateConfig,
      updateSortFields
    }
  }
}
</script>

<style scoped>
.sorter-config {
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

.form-select,
.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background: white;
}

.form-select:focus,
.form-input:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.help-text {
  display: block;
  margin-top: 4px;
  color: #666;
  font-size: 12px;
}

.radio-group {
  display: flex;
  gap: 20px;
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
</style>
