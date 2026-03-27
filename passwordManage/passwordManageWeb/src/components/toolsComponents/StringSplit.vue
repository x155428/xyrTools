<template>
  <div class="awk-tool" style="display: flex;overflow: auto;margin: 3px;padding: 3px;">
    <!-- 输入区域 -->
    <div class="section input-section">
      <h3>初始输入文本</h3>
      <textarea v-model="initialInputText" placeholder="请输入需要处理的文本..." rows="5"></textarea>
      <button @click="addProcessingStep" class="primary-btn">添加处理步骤</button>
    </div>

    <!-- 处理链区域 -->
    <div class="section chain-section">
      <h3>处理链</h3>
      <div v-for="(step, index) in processingSteps" :key="index" class="processing-step">
        <h4>步骤 {{ index + 1 }}</h4>
        <label for="splitMode">分隔模式:</label>
        <select v-model="step.splitMode">
          <option value="character">按字符分隔</option>
          <option value="position">按位置分隔</option>
        </select>

        <div v-if="step.splitMode === 'character'" class="character-config">
          <label for="fieldSeparator">字段分隔符:</label>
          <input v-model="step.fieldSeparator" type="text" />

          <label for="selectedFields">选择字段 (逗号分隔，默认全字段输出):</label>
          <input v-model="step.selectedFields" type="text" />


          <div>
            <label>分隔符操作:</label>
            <label>
              <input type="radio" v-model="step.separatorAction" value="keep" />
              保留分隔符
            </label>
            <label>
              <input type="radio" v-model="step.separatorAction" value="remove" />
              删除分隔符
            </label>
            <label>
              <input type="radio" v-model="step.separatorAction" value="replace" />
              替换分隔符
            </label>
          </div>

          <div v-if="step.separatorAction === 'replace'">
            <label for="replacement">替换为:</label>
            <input v-model="step.replacement" type="text" placeholder="替换字符" />
          </div>
        </div>

        <div v-if="step.splitMode === 'position'" class="position-config">
          <label for="cutPositions">切割位置 (逗号分隔，支持负数):</label>
          <input v-model="step.cutPositions" type="text" />

          <label for="selectedParts">选择部分 (逗号分隔):</label>
          <input v-model="step.selectedParts" type="text" />

          <label for="positionSeparator">分隔符 (可选):</label>
          <input v-model="step.positionSeparator" type="text" placeholder="默认无分隔符" />
        </div>

        <div class="button-group">
          <button @click="processStep(index)" class="secondary-btn">处理</button>
          <button @click="removeProcessingStep(index)" class="danger-btn">删除步骤</button>
        </div>

        <div v-if="step.output">
          <h5>输出结果</h5>
          <textarea :value="step.output" readonly rows="3"></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue';

  const initialInputText = ref(''); // 初始输入文本
  const processingSteps = reactive([]); // 处理步骤链

  // 添加处理步骤
  const addProcessingStep = () => {
    processingSteps.push({
      inputText: processingSteps.length === 0 ? initialInputText.value : '',
      output: '',
      splitMode: 'character',
      fieldSeparator: ',',
      selectedFields: '',
      cutPositions: '',
      selectedParts: '1',
      separatorAction: 'remove', // 分隔符操作，默认为删除
      replacement: '', // 替换字符
      positionSeparator: '', // 按位置分割时的分隔符
    });
  };

  // 删除处理步骤
  const removeProcessingStep = (index) => {
    processingSteps.splice(index, 1);
  };

  const escapeRegExp = (str) => {
    return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); // 转义正则特殊字符
  };
  // 处理单个步骤
  const processStep = (index) => {
    const step = processingSteps[index];
    const inputText = index === 0 ? initialInputText.value : processingSteps[index - 1].output;

    const lines = inputText.split('\n');

    step.output = lines
      .map((line) => {
        if (!line.trim()) return ''; // 忽略空行

        if (step.splitMode === 'position') {
          // 按位置切割并提取部分
          const cutPositionsArray = parseCutPositions(step.cutPositions);
          const selectedPartsArray = parseSelectedParts(step.selectedParts);
          const extracted = extractPartsByPosition(line, cutPositionsArray, selectedPartsArray);
          if (step.positionSeparator) {
            return extracted.join(step.positionSeparator);
          }
          return extracted.join('');
        } else {
          // 按字符分隔处理
          const fieldIndexes = parseFieldIndexes(step.selectedFields, line, step.fieldSeparator);

          const escapedSeparator = escapeRegExp(step.fieldSeparator);
          // 使用正则表达式捕获分隔符并分隔字符串
          const regex = new RegExp(`(${escapedSeparator})`, 'g');
          const parts = line.split(regex);

          // 构造结果，仅提取字段
          let result = fieldIndexes
            .map((index) => {
              const field = parts[index * 2]; // 索引对应字段位置，跳过分隔符
              return field !== undefined ? field : '';
            })
            .join(''); // 默认行为

          // 根据分隔符操作调整结果
          if (step.separatorAction === 'keep') {
            result = fieldIndexes
              .map((index) => {
                const field = parts[index * 2];
                const separator = parts[index * 2 + 1] || '';
                return field + separator;
              })
              .join('');
          } else if (step.separatorAction === 'replace') {
            result = fieldIndexes
              .map((index) => {
                const field = parts[index * 2];
                return field !== undefined ? field : '';
              })
              .join(step.replacement || '');
          } else if (step.separatorAction === 'remove') {
            result = fieldIndexes
              .map((index) => {
                const field = parts[index * 2];
                return field !== undefined ? field : '';
              })
              .join('');
          }

          return result;
        }
      })
      .join('\n');
  };

  // 工具函数
  const parseFieldIndexes = (fields, line, separator) => {
    const escapedSeparator = escapeRegExp(separator); // 转义
    const regex = new RegExp(`(${escapedSeparator})`, 'g'); // 使用转义后的分隔符
    if (!fields.trim()) {
      // 如果 `fields` 为空，返回所有字段索引
      return Array.from({ length: line.split(regex).length / 2 }, (_, i) => i);
    }
    return fields
      .split(',')
      .map((index) => parseInt(index.trim(), 10) - 1)
      .filter((index) => !isNaN(index));
  };

  const parseCutPositions = (positions) => {
    return positions
      .split(',')
      .map((pos) => parseInt(pos.trim(), 10))
      .filter((pos) => !isNaN(pos));
  };

  const parseSelectedParts = (parts) => {
    return parts
      .split(',')
      .map((index) => parseInt(index.trim(), 10) - 1)
      .filter((index) => !isNaN(index));
  };

  const extractPartsByPosition = (text, positions, selectedParts) => {
    const segments = [];
    positions = [0, ...positions, text.length]; // 确保包含起点和终点
    for (let i = 0; i < positions.length - 1; i++) {
      const start = positions[i] >= 0 ? positions[i] : text.length + positions[i];
      const end = positions[i + 1] >= 0 ? positions[i + 1] : text.length + positions[i + 1];
      segments.push(text.slice(Math.max(0, start), Math.min(text.length, end)));
    }

    // 确保选择的部分索引在有效范围内
    return selectedParts
      .filter((index) => index >= 0 && index < segments.length) // 过滤掉无效索引
      .map((index) => segments[index] || ''); // 提取有效段
  };

</script>

<style scoped>
  .awk-tool {
    flex-direction: column;
    gap: 20px;
    max-width: 100%;
    max-height: calc(100vh - 150px);
    margin: auto;
    padding: 10px;
    box-sizing: border-box;
    border: 1px solid #ddd;
    border-radius: 8px;
    background-color: #f9f9f9;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .preprocessing-options label {
    font-weight: normal;
    color: #555;
  }

  .processing-step {
    border: 1px solid #ccc;
    border-radius: 4px;
    padding: 15px;
    background-color: #fff;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    margin-bottom: 20px;
  }

  h3 {
    font-size: 20px;
    color: #333;
    margin-bottom: 10px;
    border-bottom: 2px solid #007bff;
    padding-bottom: 5px;
  }

  h4 {
    font-size: 18px;
    color: #555;
    margin-bottom: 10px;
  }

  textarea {
    width: 100%;
    padding: 10px;
    font-size: 16px;
    border: 1px solid #ccc;
    border-radius: 4px;
    resize: vertical;
    background-color: #fff;
    line-height: 1.6;
    box-sizing: border-box;
  }

  label {
    font-weight: bold;
    color: #555;
  }

  input,
  select {
    padding: 10px;
    font-size: 16px;
    border: 1px solid #ccc;
    border-radius: 4px;
    background-color: #fff;
    box-sizing: border-box;
  }

  button {
    padding: 12px 25px;
    font-size: 16px;
    border: none;
    background-color: #007bff;
    color: white;
    cursor: pointer;
    border-radius: 4px;
    transition: background-color 0.3s ease, transform 0.2s;
  }

  button:hover {
    background-color: #0056b3;
    transform: translateY(-2px);
  }

  .chain-section {
    display: flex;
    flex-direction: column;
    gap: 20px;
    overflow-x: auto;
    /* 确保横向可滚动 */
  }

  /* 按钮组样式 */
  .button-group {
    display: flex;
    justify-content: flex-start;
    gap: 10px;
    /* 按钮之间的间距 */
    margin-top: 10px;
  }

  .primary-btn {
    background-color: #007bff;
  }

  .secondary-btn {
    background-color: #28a745;
  }

  .danger-btn {
    background-color: #dc3545;
  }
</style>