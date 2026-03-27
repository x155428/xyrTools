<template>
  <div class="insert-tool" style="display: flex;overflow: auto;margin: 3px;padding: 3px;">
    <!-- 输入区域 -->
    <div class="section input-section">
      <h3>初始输入文本</h3>
      <textarea v-model="initialInputText" placeholder="请输入需要处理的文本..." rows="5"></textarea>
      <button @click="addProcessingStep">添加处理步骤</button>
    </div>

    <!-- 处理链区域 -->
    <div class="section chain-section">
      <h3>处理链</h3>
      <div v-for="(step, index) in processingSteps" :key="index" class="processing-step">
        <h4>步骤 {{ index + 1 }}</h4>

        <!-- 插入模式 -->
        <label>插入模式:</label>
        <select v-model="step.insertMode">
          <option value="character">按字符插入</option>
          <option value="position">按位置插入</option>
        </select>

        <!-- 按字符插入配置 -->
        <div v-if="step.insertMode === 'character'">
          <label>目标字符:</label>
          <input v-model="step.targetCharacter" type="text" />

          <label>插入位置:</label>
          <select v-model="step.insertLocation">
            <option value="before">目标字符前</option>
            <option value="after">目标字符后</option>
          </select>
        </div>

        <!-- 按位置插入配置 -->
        <div v-if="step.insertMode === 'position'">
          <label>插入位置 (逗号分隔，支持负数):</label>
          <input v-model="step.insertPositions" type="text" />

          <label>插入位置:</label>
          <select v-model="step.insertLocation">
            <option value="before">位置前</option>
            <option value="after">位置后</option>
          </select>
        </div>

        <!-- 表达式输入 -->
        <label>插入内容 (支持简单表达式):</label>
        <el-tooltip class="box-item" effect="dark"
          content="支持的变量: ${index}, ${line}, ${length}, ${date}, ${time} 示例: ${index * 2}, ${length + 5}, ${line.toUpperCase()}"
          placement="top-start">
          <input v-model="step.expression" type="text"
            placeholder="例如: [序号:${index}], ${index * 2}, ${length + 5}, ${line.toUpperCase()}" style="width: 50%;"/>
        </el-tooltip>
        <div>
          <button @click="processStep(index)">处理</button>
          <button @click="removeProcessingStep(index) " class="danger-btn">删除步骤</button>
        </div>


        <!-- 输出结果 -->
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
      insertMode: 'character',
      targetCharacter: '',
      insertPositions: '',
      insertLocation: 'before',
      expression: '',
    });
  };

  // 删除处理步骤
  const removeProcessingStep = (index) => {
    processingSteps.splice(index, 1);
  };

  // 解析表达式，支持动态变量和简单计算逻辑
  const resolveExpression = (expression, context) => {
    try {
      const func = new Function(...Object.keys(context), `return \`${expression}\`;`);
      return func(...Object.values(context));
    } catch (error) {
      console.error('表达式解析错误:', error);
      return expression; // 出现错误时返回原表达式
    }
  };

  // 工具函数
  const parsePositions = (positions) => {
    return positions
      .split(',')
      .map((pos) => parseInt(pos.trim(), 10))
      .filter((pos) => !isNaN(pos));
  };

  // 处理单个步骤
  const processStep = (index) => {
    const step = processingSteps[index];
    const inputText = index === 0 ? initialInputText.value : processingSteps[index - 1].output;

    const lines = inputText.split('\n');
    step.output = lines
      .map((line, lineIndex) => {
        if (!line.trim()) return ''; // 忽略空行

        const context = {
          index: lineIndex + 1, // 当前行索引，从1开始
          line, // 当前行内容
          length: line.length, // 当前行长度
          date: new Date().toLocaleDateString(), // 当前日期
          time: new Date().toLocaleTimeString(), // 当前时间
        };

        let result = line;

        // 应用表达式
        if (step.expression) {
          step.insertValue = resolveExpression(step.expression, context);
        }

        if (step.insertMode === 'position') {
          // 按位置插入
          const positions = parsePositions(step.insertPositions);
          positions.forEach((pos) => {
            if (pos < 0 || pos >= result.length) {
              result += step.insertValue; // 超出位置追加到末尾
            } else {
              result =
                step.insertLocation === 'before'
                  ? result.slice(0, pos) + step.insertValue + result.slice(pos)
                  : result.slice(0, pos + 1) + step.insertValue + result.slice(pos + 1);
            }
          });
        } else if (step.insertMode === 'character') {
          // 按字符插入
          const target = step.targetCharacter;
          if (target) {
            result = result.split(target).join(
              step.insertLocation === 'before'
                ? step.insertValue + target
                : target + step.insertValue
            );
          }
        }

        return result;
      })
      .join('\n');
  };
</script>

<style scoped>
  .insert-tool {
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