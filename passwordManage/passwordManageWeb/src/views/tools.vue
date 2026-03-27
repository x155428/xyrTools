<template>
  <div class="table-panel">
    <!-- 工具分类导航 -->
    <div v-if="isNavVisible" class="category-header">
      <div v-for="(category, index) in categories" :key="index" @click="setCategory(index)"
        :class="['category-tab', { active: activeCategory === index }]"
        :style="{ background: activeCategory === index ? 'linear-gradient(to bottom, #e6f7ff, #b3d9ff)' : '#f8f8f8' }">
        {{ category.name }}
      </div>
    </div>


    <!-- 分类模块中的功能列表 -->
    <div class="function-panel-container">
      <div v-if="isFunctionPanelVisible" class="function-panel">

        <div class="function-list">
          <label v-for="(func, index) in categories[activeCategory].functions" :key="index">
            <input type="radio" name="functionSelect" :value="func.name" v-model="activeFunction" />
            {{ func.name }}
          </label>
        </div>
      </div>
      <!-- 折叠/展开箭头 -->
      <div @click="toggleFunctionPanel" class="toggle-arrow" :class="{ 'expanded': isFunctionPanelVisible }">
        {{ isFunctionPanelVisible ? '▲' : '▼' }}
      </div>
    </div>

    <!-- 功能区 -->
    <div v-if="activeFunction" class="function-content" style="max-height: 92%;height: 100%;overflow: hidden;">

      <!-- Base64 功能 -->
      <div v-if="activeFunction === 'Base64'">
        <Base64Tool />
      </div>

      <!-- URL 编码解码功能 -->
      <div v-if="activeFunction === 'URL'">
        <URLTool />
      </div>

      <!-- ASCII码表功能 -->
      <div v-if="activeFunction === 'ASCII码表'">
        <AsciiTable />
      </div>

      <!-- 文件对比功能 -->
      <div v-if="activeFunction === '文本对比'">
        <Fie_compare />
      </div>

      <!-- 文件校验 -->
      <div v-if="activeFunction === '文件校验'">
        <HashFile />
      </div>

      <!-- 文件校验值计算 -->
      <div v-if="activeFunction === '校验值计算'">
        <FileChecksum />
      </div>

      <div v-if="activeFunction === 'AES 加密'">
        <AesEncryptFile />
      </div>

      <!-- 链式文本处理功能 -->
      <div v-if="activeFunction === '链式文本处理'">
        <ChainedTextProcessor />
      </div>

      <!-- XML 转换功能 -->
      <div v-if="activeFunction === 'XML 转换'">
        <p>XML 转换功能的内容。</p>
      </div>

      <!-- 进制转换功能 -->
      <div v-if="activeFunction === '进制转换'">
        <EncodingConversion />
      </div>

      <!-- RSA 解密功能 -->
      <div v-if="activeFunction === 'RSA 解密'">
        <p>RSA 解密功能的内容。</p>
      </div>

      <!-- API 管理功能 -->
      <div v-if="activeFunction === 'winAPI管理'">
        <ApiManager />
      </div>

      <!-- JSON 编辑器功能 -->
      <div v-if="activeFunction === 'JSON 转换'">
        <JsonEditor />
      </div>

      <!-- 时间戳互转功能 -->
      <div v-if="activeFunction === '时间戳互转'">
        <TimestampConverter />
      </div>

      <!-- 如果没有选中功能或选中的功能未实现 -->
      <div
        v-if="activeFunction && !['Base64', 'URL' ,'ASCII码表','文本对比','链式文本处理','校验值计算','文件校验','AES 加密','winAPI管理','进制转换','时间戳互转', 'JSON 转换'].includes(activeFunction)">
        <p>功能 "{{ activeFunction }}" 的功能内容未实现。</p>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import Base64Tool from '../components/toolsComponents/Base64Tools.vue'
  import URLTool from '../components/toolsComponents/UrlTools.vue'
  import AsciiTable from '../components/toolsComponents/AsciiTable.vue'
  import Fie_compare from '../components/toolsComponents/File_Compare.vue'
  import HashFile from '../components/toolsComponents/HashFile.vue'
  import FileChecksum from '../components/toolsComponents/FileChecksum.vue'
  import AesEncryptFile from '../components/toolsComponents/AesEncryptFile.vue'
  import EncodingConversion from '../components/toolsComponents/EncodingConversion.vue'
  import StringSplit from '../components/toolsComponents/StringSplit.vue'
  import InsertString from '../components/toolsComponents/InsertString.vue'
  import ApiManager from '../components/toolsComponents/ApiManager.vue'
  import JsonEditor from '../components/toolsComponents/JsonEditor.vue'
  import ChainedTextProcessor from '../components/toolsComponents/ChainedTextProcessor.vue'
  import TimestampConverter from '../components/toolsComponents/TimestampConverter.vue'
  // 定义工具分类及功能
  const categories = ref([
    {
      name: '编码/解码',
      functions: [
        { name: 'Base64' },
        { name: 'URL' },
        { name: 'ASCII码表' }
      ]
    },
    {
      name: '文本处理',
      functions: [
        { name: '链式文本处理' },
        { name: '文本对比' },
      ]
    },
    {
      name: '转换',
      functions: [
        { name: '进制转换' },
        { name: 'JSON 转换' },
        { name: 'XML 转换' },
        { name: '时间戳互转' },

      ]
    },
    {
      name: '校验',
      functions: [
        { name: '校验值计算' },
        { name: '文件校验' }
      ]
    },
    {
      name: '加解密',
      functions: [
        { name: 'AES 加密' },
        { name: 'RSA 解密' }
      ]
    },
    {
      name: 'WinApi管理',
      functions: [
        { name: 'winAPI管理' },
      ]
    }
  ])

  // 当前激活的分类索引
  const activeCategory = ref(0)
  // 当前选中的功能
  const activeFunction = ref(null)
  // 控制导航栏是否可见
  const isNavVisible = ref(true)
  // 控制功能面板是否可见
  const isFunctionPanelVisible = ref(true)

  // 设置当前激活的分类
  const setCategory = (index) => {
    activeCategory.value = index
    // 当切换分类时，设置为该分类下的第一个功能
    activeFunction.value = categories.value[index].functions[0].name
  }

  // 切换功能面板显示状态
  const toggleFunctionPanel = () => {
    isFunctionPanelVisible.value = !isFunctionPanelVisible.value
  }

  // 在组件加载时设置默认选中项
  onMounted(() => {
    // 设置默认分类为第一个分类，默认功能为该分类下的第一个功能
    activeFunction.value = categories.value[activeCategory.value].functions[0].name
  })
</script>

<style scoped>
  .table-panel {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 99%;
  }

  /* 工具分类导航样式 */
  .category-header {
    display: flex;
    border-bottom: 2px solid var(--border-color-light);
    background-color: var(--bg-muted);
  }

  .category-tab {
    padding: 5px 5px;
    cursor: pointer;
    text-align: center;
    background-color: var(--bg-container);
    flex: 1;
    border-bottom: 2px solid transparent;
    transition: all 0.3s ease;
  }

  /* 功能面板容器样式 */
  .function-panel-container {
    position: relative;
    margin-bottom: 1px;
    margin-top: 1px;
  }

  /* 功能面板样式 */
  .function-panel {
    padding: 5px;
    border: 1px solid var(--border-color-light);
    border-top: none;
    background-color: var(--bg-container);
    position: relative;
  }

  /* 折叠/展开箭头样式 */
  .toggle-arrow {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
    padding: 5px 10px;
    background-color: transparent;
    color: var(--color-primary);
    cursor: pointer;
    transition: background-color 0.3s ease;
    font-size: 12px;
    z-index: 1;
    width: 50%;
    height: 13px;
    line-height: 15px;
    text-align: center;
  }

  .toggle-arrow.expanded {
    bottom: 0;
  }

  /* 功能区样式 */
  .function-content {
    padding: 10px;
    border: 1px solid var(--border-color-light);
    margin-top: 1px;
    background-color: var(--bg-container);
  }
</style>