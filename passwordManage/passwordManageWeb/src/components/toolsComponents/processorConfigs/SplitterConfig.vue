<template>
  <div class="splitter-config">
    <div class="config-group">
      <label class="config-label">分割模式</label>
      <select v-model="config.mode" @change="updateConfig" class="form-select">
        <option value="string">按字符串分割</option>
        <option value="position">按位置分割</option>
      </select>
    </div>

    <!-- 按字符串分割配置 -->
    <div v-if="config.mode === 'string'" class="config-group">
      <label class="config-label">分隔符</label>
      <input v-model="config.separator" @input="updateConfig" placeholder="输入分隔符，默认为空格" class="form-input" />

      <label class="config-label">分隔符处理</label>
      <select v-model="config.separatorHandling" @change="updateConfig" class="form-select">
        <option value="keep">保留分隔符</option>
        <option value="remove">删除分隔符</option>
        <option value="replace">替换分隔符</option>
      </select>

      <div v-if="config.separatorHandling === 'replace'" class="config-group">
        <label class="config-label">替换为</label>
        <input v-model="config.replacement" @input="updateConfig" placeholder="替换分隔符的字符串" class="form-input" />
      </div>
    </div>

    <!-- 按位置分割配置 -->
    <div v-if="config.mode === 'position'" class="config-group">
      <label class="config-label">分割位置</label>
      <input v-model="positionInput" @input="updatePositions" placeholder="输入位置，如：1,3,5 或 1-3,5" class="form-input" />
      <small class="help-text">位置从1开始，多个位置用逗号分隔，范围用-表示</small>
    </div>

    <!-- 输出配置 -->
    <div class="config-group">
      <label class="config-label">输出模式</label>
      <select v-model="config.outputMode" @change="updateConfig" class="form-select">
        <option value="all">输出全部字段</option>
        <option value="specific">指定字段</option>
      </select>
    </div>

    <div v-if="config.outputMode === 'specific'" class="config-group">
      <label class="config-label">选择字段</label>
      <input v-model="selectedFieldsInput" @input="updateSelectedFields" placeholder="输入字段编号，如：1,2,5"
        class="form-input" />
      <small class="help-text">字段编号从1开始，多个字段用逗号分隔</small>
    </div>

    <div class="config-group">
      <label class="config-label">输出分隔符</label>
      <input v-model="config.outputSeparator" @input="updateConfig" placeholder="输出时字段间的分隔符，支持\n\r\t等转义字符"
        class="form-input" />
      <small class="help-text">支持的转义字符：\n（换行）、\r（回车）、\t（制表符）</small>
    </div>
  </div>
</template>

<script>
  import { ref, watch } from 'vue'

  export default {
    name: 'SplitterConfig',
    props: {
      modelValue: {
        type: Object,
        default: () => ({})
      }
    },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
      const config = ref({
        mode: 'string',
        separator: ' ',
        separatorHandling: 'keep',
        replacement: '',
        positions: [],
        outputMode: 'all',
        selectedFields: [],
        outputSeparator: ''
      })

      const positionInput = ref('')
      const selectedFieldsInput = ref('')

      const updateConfig = () => {
        emit('update:modelValue', { ...config.value })
      }

      const updateInputs = () => {
        // 更新位置输入 - 只在输入框为空时更新
        if (positionInput.value === '') {
          positionInput.value = config.value.positions && config.value.positions.length > 0
            ? config.value.positions.join(',')
            : ''
        }

        // 更新字段输入 - 只在输入框为空时更新
        if (selectedFieldsInput.value === '') {
          selectedFieldsInput.value = config.value.selectedFields && config.value.selectedFields.length > 0
            ? config.value.selectedFields.join(',')
            : ''
        }
      }

      // 初始化配置
      watch(() => props.modelValue, (newVal) => {
        if (newVal && Object.keys(newVal).length > 0) {
          // 保存当前输入值，防止被覆盖
          const currentSelectedFieldsInput = selectedFieldsInput.value
          const currentPositionInput = positionInput.value

          config.value = { ...config.value, ...newVal }

          // 只在输入框为空时更新，避免覆盖用户正在输入的内容
          if (currentSelectedFieldsInput === '') {
            updateInputs()
          }
        }
      }, { immediate: true })

      const updatePositions = () => {
        const positions = []
        const parts = positionInput.value.split(',')

        for (const part of parts) {
          if (part.includes('-')) {
            const [start, end] = part.split('-').map(Number)
            for (let i = start; i <= end; i++) {
              positions.push(i)
            }
          } else {
            const pos = Number(part)
            if (!isNaN(pos)) {
              positions.push(pos)
            }
          }
        }

        config.value.positions = positions
        updateConfig()
      }

      const updateSelectedFields = () => {
        const fields = []
        const parts = selectedFieldsInput.value.split(',')

        for (const part of parts) {
          const trimmedPart = part.trim()
          if (trimmedPart === '') continue

          if (trimmedPart.includes('-')) {
            // 处理范围格式，如 1-3
            const [startStr, endStr] = trimmedPart.split('-')
            const start = Number(startStr.trim())
            const end = Number(endStr.trim())

            if (!isNaN(start) && !isNaN(end) && start > 0 && end >= start) {
              for (let i = start; i <= end; i++) {
                fields.push(i)
              }
            }
          } else {
            // 处理单个数字格式
            const field = Number(trimmedPart)
            if (!isNaN(field) && field > 0) {
              fields.push(field)
            }
          }
        }

        config.value.selectedFields = fields
        updateConfig()
      }

      return {
        config,
        positionInput,
        selectedFieldsInput,
        updateConfig,
        updatePositions,
        updateSelectedFields,
        updateInputs
      }
    }
  }
</script>

<style scoped>
  .splitter-config {
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
</style>