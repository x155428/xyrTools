<template>
  <div class="inserter-config">
    <div class="config-group">
      <label class="config-label">插入模式</label>
      <select v-model="config.mode" @change="updateConfig" class="form-select">
        <option value="character">按字符插入</option>
        <option value="position">按位置插入</option>
      </select>
    </div>

    <!-- 按字符插入配置 -->
    <div v-if="config.mode === 'character'" class="config-group">
      <label class="config-label">目标字符</label>
      <input 
        v-model="config.targetChar" 
        @input="updateConfig"
        placeholder="输入要插入的目标字符"
        class="form-input"
      />
      
      <label class="config-label">插入位置</label>
      <select v-model="config.position" @change="updateConfig" class="form-select">
        <option value="before">在字符前插入</option>
        <option value="after">在字符后插入</option>
      </select>
    </div>

    <!-- 按位置插入配置 -->
    <div v-if="config.mode === 'position'" class="config-group">
      <label class="config-label">目标位置</label>
      <input 
        v-model="config.targetPosition" 
        @input="updateConfig"
        placeholder="输入位置编号（从1开始）"
        class="form-input"
        type="number"
        min="1"
      />
      
      <label class="config-label">插入位置</label>
      <select v-model="config.position" @change="updateConfig" class="form-select">
        <option value="before">在位置前插入</option>
        <option value="after">在位置后插入</option>
      </select>
    </div>

    <!-- 插入内容配置 -->
    <div class="config-group">
      <label class="config-label">插入内容</label>
      <textarea 
        v-model="config.insertText" 
        @input="updateConfig"
        placeholder="输入要插入的文本内容"
        class="form-textarea"
        rows="3"
      ></textarea>
    </div>

    <!-- 插入选项 -->
    <div class="config-group">
      <label class="config-label">插入选项</label>
      <div class="checkbox-group">
        <label class="checkbox-label">
          <input 
            type="checkbox" 
            v-model="config.global" 
            @change="updateConfig"
          />
          全局插入（每行都插入）
        </label>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'

export default {
  name: 'InserterConfig',
  props: {
    modelValue: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const config = ref({
      mode: 'character',
      targetChar: '',
      targetPosition: 1,
      position: 'before',
      insertText: '',
      global: true
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
.inserter-config {
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

.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background: white;
  resize: vertical;
  font-family: inherit;
}

.form-select:focus,
.form-input:focus,
.form-textarea:focus {
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
</style>
