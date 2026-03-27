<!--
 * @Author: 小鱼
 * @Date: 2025-09-02 13:45:37
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-25 15:19:21
 * @FilePath: \passwordManageWeb\src\components\toolsComponents\ChainedTextProcessor.vue
 * @Description: 
 * 
 * Copyright (c) 2025 by 小鱼, All Rights Reserved. 
-->
<template>
  <div class="chained-text-processor">
    <!-- 输入区域 -->
    <div class="input-section" :class="{ 'collapsed': isInputCollapsed }">
      <div class="section-header">
        <h3>输入文本</h3>
        <button class="toggle-btn" @click="toggleInputSection">
          {{ isInputCollapsed ? '展开' : '收起' }}
        </button>
      </div>

      <div v-if="!isInputCollapsed" class="input-content">
        <div class="input-actions">
          <button class="btn btn-secondary" @click="pasteFromClipboard">粘贴</button>
          <button class="btn btn-secondary" @click="openFile">打开文件</button>
          <button class="btn btn-secondary" @click="clearInput">清空</button>
        </div>
        <textarea v-model="inputText" class="input-textarea" placeholder="在此输入文本，或使用粘贴/打开文件功能"
          @input="processAllChains"></textarea>
        <input ref="fileInput" type="file" style="display: none;" @change="handleFileSelect" />
      </div>
    </div>

    <!-- 链式处理区域 -->
    <div class="processing-section">
      <div class="chain-tabs">
        <div class="tabs-header">
          <div v-for="(chain, index) in processingChains" :key="chain.id" class="tab"
            :class="{ 'active': activeChainIndex === index }" @click="switchToChain(index)">
            <span>链 {{ index + 1 }}</span>
            <button class="close-tab-btn" @click.stop="removeChain(index)">×</button>
          </div>
          <button class="add-chain-btn" @click="addNewChain">+</button>
        </div>
      </div>

      <div class="chain-content">
        <!-- 处理器列表 -->
        <div class="processors-container" :style="{ width: processorsWidth }">
          <div class="processors-header">
            <h3>处理器列表</h3>
            <div class="processor-selector">
              <select v-model="selectedProcessorType" class="form-select">
                <option value="">选择处理器类型</option>
                <option value="splitter">文本分割提取器</option>
                <option value="inserter">文本插入器</option>
                <option value="formatter">文本格式化器</option>
                <option value="sorter">排序器</option>
                <option value="extractor">数据提取器</option>
              </select>
              <button class="btn btn-primary" @click="addProcessor">添加</button>
            </div>
          </div>

          <div class="processors-list">
            <div v-for="(processor, processorIndex) in currentChain.processors" :key="processor.id"
              class="processor-card" :class="{ 'active': focusedProcessorIndex === processorIndex }"
              @click="focusProcessor(processorIndex)">
              <div class="processor-header">
                <span class="processor-index">{{ processorIndex + 1 }}</span>
                <span class="processor-name">{{ getProcessorTypeName(processor.type) }}</span>
                <div class="processor-actions">
                  <button class="btn btn-sm btn-outline-secondary" @click.stop="toggleProcessorConfig(processorIndex)"
                    title="展开/收起配置">
                    {{ isConfigExpanded(processorIndex) ? '▼' : '▶' }}
                  </button>
                  <button class="btn btn-sm btn-outline-secondary" @click.stop="moveProcessor(processorIndex, -1)"
                    :disabled="processorIndex === 0">
                    ↑
                  </button>
                  <button class="btn btn-sm btn-outline-secondary" @click.stop="moveProcessor(processorIndex, 1)"
                    :disabled="processorIndex === currentChain.processors.length - 1">
                    ↓
                  </button>
                  <button class="btn btn-sm btn-danger" @click.stop="removeProcessor(processorIndex)">
                    删除
                  </button>
                </div>
              </div>
              <div class="processor-config" v-if="isConfigExpanded(processorIndex)">
                <component :is="getProcessorConfigComponent(processor.type)" v-model="processor.config"
                  @update:modelValue="onProcessorConfigUpdate(processorIndex, $event)" />
              </div>
            </div>

            <div v-if="currentChain.processors.length === 0" class="empty-state">
              <p>暂无处理器，请添加</p>
            </div>
          </div>
        </div>

        <!-- 可拖动分隔线 -->
        <div class="resizer" @mousedown="startResize" :class="{ 'resizing': isResizing }"></div>

        <!-- 结果展示区域 -->
        <div class="preview-container" :style="{ width: previewWidth }">
          <h3>处理器预览</h3>
          <div v-if="focusedProcessorIndex !== null" class="preview-content">
            <div class="preview-section">
              <h4>输入</h4>
              <div class="preview-text">{{ getProcessorInput(focusedProcessorIndex) }}</div>
            </div>
            <div class="preview-section">
              <div class="section-header-with-action">
                <h4>输出</h4>
                <button v-if="focusedProcessorIndex !== null" class="copy-btn" @click="copyOutputToClipboard($event)"
                  title="复制输出结果">
                  复制
                </button>
              </div>
              <div class="preview-text">{{ getProcessorOutput(focusedProcessorIndex) }}</div>
            </div>
          </div>
          <div v-else class="no-preview">
            <p>请选择一个处理器查看预览</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import { ref, computed, watch } from 'vue'
  import SplitterConfig from './processorConfigs/SplitterConfig.vue'
  import InserterConfig from './processorConfigs/InserterConfig.vue'
  import FormatterConfig from './processorConfigs/FormatterConfig.vue'
  import SorterConfig from './processorConfigs/SorterConfig.vue'
  import ExtractorConfig from './processorConfigs/ExtractorConfig.vue'

  // 处理器类型映射
  const PROCESSOR_TYPES = {
    splitter: { name: '文本分割提取器', component: SplitterConfig },
    inserter: { name: '文本插入器', component: InserterConfig },
    formatter: { name: '文本格式化器', component: FormatterConfig },
    sorter: { name: '排序器', component: SorterConfig },
    extractor: { name: '数据提取器', component: ExtractorConfig }
  }

  export default {
    name: 'ChainedTextProcessor',
    components: {
      SplitterConfig,
      InserterConfig,
      FormatterConfig,
      SorterConfig,
      ExtractorConfig
    },
    setup() {
      // 响应式数据
      const inputText = ref('')
      const isInputCollapsed = ref(false)
      const fileInput = ref(null)
      const processingChains = ref([{
        id: Date.now(),
        processors: []
      }])
      const activeChainIndex = ref(0)
      const selectedProcessorType = ref('')
      const focusedProcessorIndex = ref(null)
      const processorResults = ref({})
      // 用于跟踪每个处理器配置面板的展开/收起状态
      const expandedConfigs = ref({})
      // 用于控制左右板块的宽度
      const processorsWidth = ref('60%')
      const previewWidth = ref('40%')
      const isResizing = ref(false)

      // 计算属性
      const currentChain = computed(() => processingChains.value[activeChainIndex.value])

      // 处理所有链式结果
      const processAllChains = () => {
        processorResults.value = {}

        processingChains.value.forEach((chain, chainIndex) => {
          let currentText = inputText.value

          chain.processors.forEach((processor, processorIndex) => {
            const key = `${chainIndex}_${processorIndex}`
            processorResults.value[key] = {
              input: currentText,
              output: applyProcessor(currentText, processor)
            }
            currentText = processorResults.value[key].output
          })
        })
      }

      // 应用处理器
      const applyProcessor = (text, processor) => {
        try {
          switch (processor.type) {
            case 'splitter':
              return applySplitter(text, processor.config)
            case 'inserter':
              return applyInserter(text, processor.config)
            case 'formatter':
              return applyFormatter(text, processor.config)
            case 'sorter':
              return applySorter(text, processor.config)
            case 'extractor':
              return applyExtractor(text, processor.config)
            default:
              return text
          }
        } catch (error) {
          console.error('处理器执行错误:', error)
          return `处理错误: ${error.message}`
        }
      }

      // 数据提取器实现 - 支持正则匹配提取数据
      const applyExtractor = (text, config) => {
        if (!text || !config.pattern) return text

        const lines = text.split('\n')
        const globalMatch = config.globalMatch !== false
        const ignoreCase = config.ignoreCase || false
        const distinct = config.distinct || false
        const showLineNumbers = config.showLineNumbers || false
        const captureGroup = config.captureGroup || '0'

        let flags = ''
        if (ignoreCase) flags += 'i'
        if (globalMatch) flags += 'g'

        try {
          const regex = new RegExp(config.pattern, flags)
          const results = new Set()
          const extractedLines = []

          lines.forEach((line, lineIndex) => {
            let matches
            if (globalMatch) {
              matches = Array.from(line.matchAll(regex))
              matches.forEach(match => {
                let value
                if (captureGroup === 'all') {
                  // 提取所有捕获组
                  const groups = match.slice(1).filter(Boolean)
                  value = groups.length > 0 ? groups.join(', ') : match[0]
                } else if (captureGroup === '0') {
                  // 完整匹配
                  value = match[0]
                } else {
                  // 特定捕获组
                  const groupIndex = parseInt(captureGroup)
                  value = match[groupIndex] || match[0]
                }

                if (value) {
                  const displayValue = showLineNumbers
                    ? `${lineIndex + 1}: ${value}`
                    : value

                  if (distinct) {
                    results.add(displayValue)
                  } else {
                    extractedLines.push(displayValue)
                  }
                }
              })
            } else {
              // 非全局匹配
              const match = line.match(regex)
              if (match) {
                let value
                if (captureGroup === 'all') {
                  const groups = match.slice(1).filter(Boolean)
                  value = groups.length > 0 ? groups.join(', ') : match[0]
                } else if (captureGroup === '0') {
                  value = match[0]
                } else {
                  const groupIndex = parseInt(captureGroup)
                  value = match[groupIndex] || match[0]
                }

                if (value) {
                  const displayValue = showLineNumbers
                    ? `${lineIndex + 1}: ${value}`
                    : value

                  if (distinct) {
                    results.add(displayValue)
                  } else {
                    extractedLines.push(displayValue)
                  }
                }
              }
            }
          })

          // 组合最终结果
          const finalResults = distinct
            ? Array.from(results)
            : extractedLines

          return finalResults.join('\n')
        } catch (e) {
          console.error('正则表达式错误:', e)
          return `处理错误: ${e.message}`
        }
      }

      // 文本分割提取器实现 - 完整支持所有配置项
      const applySplitter = (text, config) => {
        const lines = text.split('\n')
        const result = []

        lines.forEach(line => {
          if (!line.trim() && config.skipEmptyLines) return

          let parts = []

          if (config.mode === 'position') {
            // 按位置分割
            if (config.positions && config.positions.length > 0) {
              let lastPos = 0
              config.positions.forEach(pos => {
                if (pos <= line.length) {
                  parts.push(line.slice(lastPos, pos))
                  lastPos = pos
                }
              })
              if (lastPos < line.length) {
                parts.push(line.slice(lastPos))
              }
            } else {
              parts = [line]
            }
          } else {
            // 按字符串分割
            if (config.separator !== undefined && config.separator !== '') {
              parts = line.split(config.separator)

              // 处理分隔符
              if (config.separatorHandling === 'keep') {
                // 保留分隔符
                const tempParts = []
                for (let i = 0; i < parts.length; i++) {
                  tempParts.push(parts[i])
                  if (i < parts.length - 1) {
                    tempParts.push(config.separator)
                  }
                }
                parts = tempParts
              } else if (config.separatorHandling === 'replace' && config.replacement !== undefined) {
                // 替换分隔符
                const tempParts = []
                for (let i = 0; i < parts.length; i++) {
                  tempParts.push(parts[i])
                  if (i < parts.length - 1) {
                    tempParts.push(config.replacement)
                  }
                }
                parts = tempParts
              }
              // 'remove' 是默认行为，不需要额外处理
            } else {
              parts = [line]
            }
          }

          // 输出处理
          let outputParts = parts
          if (config.outputMode === 'specific' && config.selectedFields && config.selectedFields.length > 0) {
            outputParts = config.selectedFields
              .map(index => parts[index - 1] || '')
              .filter(Boolean)
          }

          // 解析转义字符
          const parseEscapeCharacters = (str) => {
            return str.replace(/\\n/g, '\n')
              .replace(/\\r/g, '\r')
              .replace(/\\t/g, '\t')
          }

          // 组合输出
          const parsedSeparator = parseEscapeCharacters(config.outputSeparator ?? ' ')
          const outputLine = outputParts.join(parsedSeparator)
          result.push(outputLine)
        })

        return result.join('\n')
      }

      // 文本插入器实现 - 与新的配置结构匹配
      const applyInserter = (text, config) => {
        const lines = text.split('\n')
        const result = []

        lines.forEach((line) => {
          let processedLine = line

          if (config.global) {
            // 全局插入（每行都插入）
            if (config.mode === 'character' && config.targetChar) {
              // 按字符插入
              if (config.position === 'before') {
                // 在字符前插入
                processedLine = processedLine.split(config.targetChar)
                  .join(config.insertText + config.targetChar)
              } else {
                // 在字符后插入
                processedLine = processedLine.split(config.targetChar)
                  .join(config.targetChar + config.insertText)
              }
            } else if (config.mode === 'position') {
              // 按位置插入
              const pos = parseInt(config.targetPosition) - 1 // 转换为0-based索引
              if (!isNaN(pos) && pos >= 0 && pos <= processedLine.length) {
                if (config.position === 'before') {
                  // 在位置前插入
                  processedLine =
                    processedLine.slice(0, pos) +
                    config.insertText +
                    processedLine.slice(pos)
                } else {
                  // 在位置后插入
                  processedLine =
                    processedLine.slice(0, pos + 1) +
                    config.insertText +
                    processedLine.slice(pos + 1)
                }
              }
            }
          }

          result.push(processedLine)
        })

        return result.join('\n')
      }

      // 文本格式化器实现 - 支持所有配置模式
      const applyFormatter = (text, config) => {
        let result = text

        // 大小写转换
        switch (config.caseTransform) {
          case 'allUpper':
            result = result.toUpperCase()
            break
          case 'allLower':
            result = result.toLowerCase()
            break
          case 'word':
            // 单词首字母大写
            result = result.split('\n').map(line => {
              if (!line.trim()) return line
              return line.replace(/\w\S*/g, (txt) => {
                return txt.charAt(0).toUpperCase() + txt.substr(1)
              })
            }).join('\n')
            break
          case 'line':
            // 每行首字母大写
            result = result.split('\n').map(line => {
              if (!line.trim()) return line
              return line.charAt(0).toUpperCase() + line.slice(1)
            }).join('\n')
            break
          default:
            // 不转换
            break
        }

        // 清除全部空格
        if (config.removeAllSpaces) {
          result = result.replace(/\s/g, '')
        } else {
          // 空格处理



          // 合并连续空格（仅在同一行内）
          if (config.normalizeSpaces) {
            result = result.split('\n').map(line => {
              return line.replace(/\s+/g, ' ')
            }).join('\n')
          } else if (config.removeExtraSpaces) {
            // 保留原有清除多余空格功能，作为normalizeSpaces的别名
            result = result.split('\n').map(line => {
              return line.replace(/\s+/g, ' ')
            }).join('\n')
          }

          // 行首尾空格处理
          if (config.trimLines) {
            result = result.split('\n').map(line => {
              switch (config.trimLines) {
                case 'left':
                  return line.trimStart()
                case 'right':
                  return line.trimEnd()
                case 'both':
                  return line.trim()
                default:
                  return line
              }
            }).join('\n')
          }
        }

        // 空行处理
        if (config.removeEmptyLines) {
          const keepCount = config.keepEmptyLinesCount || 0
          const lines = result.split('\n')
          const processedLines = []
          let emptyLineCount = 0

          for (const line of lines) {
            if (line.trim() === '') {
              emptyLineCount++
              if (emptyLineCount <= keepCount) {
                processedLines.push(line)
              }
            } else {
              emptyLineCount = 0
              processedLines.push(line)
            }
          }

          result = processedLines.join('\n')
        }

        // 去重
        if (config.removeDuplicates) {
          const lines = result.split('\n')
          const uniqueLines = []
          const seen = new Set()

          for (const line of lines) {
            if (!seen.has(line)) {
              seen.add(line)
              uniqueLines.push(line)
            }
          }

          result = uniqueLines.join('\n')
        }

        // 去除不可见字符
        if (config.removeInvisibleChars) {
          result = result.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '')
        }

        // 去除换行符
        if (config.removeLineBreaks) {
          if (config.lineBreakReplace === 'empty') {
            // 替换为空格
            result = result.replace(/\r\n|\r|\n/g, '')
          } else {
            // 替换为空字符串
            result = result.replace(/\r\n|\r|\n/g, ' ')
          }
        }

        // 删除字符
        if (config.removeChars && config.matchContent) {
          let flags = ''
          if (config.global !== false) flags += 'g'
          if (config.ignoreCase) flags += 'i'

          let regex
          if (config.matchMode === 'regex') {
            try {
              regex = new RegExp(config.matchContent, flags)
            } catch (e) {
              console.error('正则表达式错误:', e)
              return result
            }
          } else {
            // 转义字符串以用于正则表达式
            const escapeRegExp = (string) => {
              return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
            }
            regex = new RegExp(escapeRegExp(config.matchContent), flags)
          }

          result = result.replace(regex, '')
        }

        return result
      }

      // 排序器实现 - 完整实现所有配置项
      const applySorter = (text, config) => {
        let lines = text.split('\n')

        // 处理空行
        const emptyLines = []
        let contentLines = []

        if (config.emptyHandling === 'ignore') {
          contentLines = lines.filter(line => line.trim())
        } else {
          // 分离空行和内容行
          lines.forEach(line => {
            if (line.trim()) {
              contentLines.push(line)
            } else {
              emptyLines.push(line)
            }
          })
        }

        // 排序内容行
        let sortedContent = [...contentLines]

        // 如果需要稳定排序，先为每行添加原始索引
        if (config.stable) {
          const linesWithIndex = contentLines.map((line, index) => ({ line, index }))

          linesWithIndex.sort((a, b) => {
            let compareResult = 0

            // 根据排序字段进行比较
            if (config.sortFields && config.sortFields.length > 0) {
              for (const fieldIndex of config.sortFields) {
                const aField = getFieldValue(a.line, fieldIndex, config)
                const bField = getFieldValue(b.line, fieldIndex, config)

                if (config.sortMode === 'numeric') {
                  compareResult = Number(aField) - Number(bField)
                } else if (config.sortMode === 'dictionary') {
                  compareResult = config.ignoreCase
                    ? aField.toLowerCase().localeCompare(bField.toLowerCase())
                    : aField.localeCompare(bField)
                } else {
                  // 混合模式
                  compareResult = compareMixed(aField, bField, config.ignoreCase)
                }

                if (compareResult !== 0) break
              }
            }

            // 稳定排序：如果比较结果相同，保持原始顺序
            if (compareResult === 0) {
              compareResult = a.index - b.index
            }

            // 排序方向
            return config.ascending ? compareResult : -compareResult
          })

          sortedContent = linesWithIndex.map(item => item.line)
        } else {
          // 非稳定排序
          sortedContent.sort((a, b) => {
            let compareResult = 0

            // 根据排序字段进行比较
            if (config.sortFields && config.sortFields.length > 0) {
              for (const fieldIndex of config.sortFields) {
                const aField = getFieldValue(a, fieldIndex, config)
                const bField = getFieldValue(b, fieldIndex, config)

                if (config.sortMode === 'numeric') {
                  compareResult = Number(aField) - Number(bField)
                } else if (config.sortMode === 'dictionary') {
                  compareResult = config.ignoreCase
                    ? aField.toLowerCase().localeCompare(bField.toLowerCase())
                    : aField.localeCompare(bField)
                } else {
                  // 混合模式
                  compareResult = compareMixed(aField, bField, config.ignoreCase)
                }

                if (compareResult !== 0) break
              }
            }

            // 排序方向
            return config.ascending ? compareResult : -compareResult
          })
        }

        // 去重 - 改进为使用排序字段进行比较
        if (config.removeDuplicates && sortedContent.length > 0) {
          const uniqueLines = [sortedContent[0]]

          for (let i = 1; i < sortedContent.length; i++) {
            let isDuplicate = false

            // 比较排序字段来判断重复
            if (config.sortFields && config.sortFields.length > 0) {
              const currentLine = sortedContent[i]
              const lastLine = uniqueLines[uniqueLines.length - 1]
              let fieldsMatch = true

              for (const fieldIndex of config.sortFields) {
                const currentField = getFieldValue(currentLine, fieldIndex, config)
                const lastField = getFieldValue(lastLine, fieldIndex, config)

                if (config.ignoreCase) {
                  if (currentField.toLowerCase() !== lastField.toLowerCase()) {
                    fieldsMatch = false
                    break
                  }
                } else {
                  if (currentField !== lastField) {
                    fieldsMatch = false
                    break
                  }
                }
              }

              isDuplicate = fieldsMatch
            } else {
              // 如果没有指定排序字段，比较整行
              isDuplicate = config.ignoreCase
                ? sortedContent[i].toLowerCase() === uniqueLines[uniqueLines.length - 1].toLowerCase()
                : sortedContent[i] === uniqueLines[uniqueLines.length - 1]
            }

            if (!isDuplicate) {
              uniqueLines.push(sortedContent[i])
            }
          }

          sortedContent = uniqueLines
        }

        // 根据配置重新组合空行和内容行
        let finalResult = []

        if (config.emptyHandling === 'first') {
          // 空行在前
          finalResult = [...emptyLines, ...sortedContent]
        } else if (config.emptyHandling === 'last') {
          // 空行在后
          finalResult = [...sortedContent, ...emptyLines]
        } else {
          // ignore或其他情况，只返回排序后的内容行
          finalResult = sortedContent
        }

        return finalResult.join('\n')
      }

      // 辅助函数：获取字段值
      const getFieldValue = (line, fieldIndex, config) => {
        if (config.fieldSeparator && config.fieldSeparator !== '') {
          const fields = line.split(config.fieldSeparator)
          const field = fields[fieldIndex - 1] || ''
          return config.trimFields ? field.trim() : field
        }
        return config.trimFields ? line.trim() : line
      }

      // 辅助函数：混合比较
      const compareMixed = (a, b, ignoreCase) => {
        if (ignoreCase) {
          a = a.toLowerCase()
          b = b.toLowerCase()
        }

        const numA = parseFloat(a)
        const numB = parseFloat(b)

        if (!isNaN(numA) && !isNaN(numB)) {
          return numA - numB
        }

        return a.localeCompare(b)
      }

      // 获取处理器类型名称
      const getProcessorTypeName = (type) => {
        return PROCESSOR_TYPES[type]?.name || '未知处理器'
      }

      // 获取处理器配置组件
      const getProcessorConfigComponent = (type) => {
        return PROCESSOR_TYPES[type]?.component || null
      }

      // 获取处理器输入
      const getProcessorInput = (processorIndex) => {
        if (processorIndex === 0) {
          return inputText.value
        }

        const key = `${activeChainIndex.value}_${processorIndex - 1}`
        return processorResults.value[key]?.output || ''
      }

      // 获取处理器输出
      const getProcessorOutput = (processorIndex) => {
        const key = `${activeChainIndex.value}_${processorIndex}`
        return processorResults.value[key]?.output || ''
      }

      // 切换到指定链
      const switchToChain = (index) => {
        activeChainIndex.value = index
        focusedProcessorIndex.value = null
      }

      // 添加新链
      const addNewChain = () => {
        processingChains.value.push({
          id: Date.now(),
          processors: []
        })
        activeChainIndex.value = processingChains.value.length - 1
        focusedProcessorIndex.value = null
      }

      // 移除链
      const removeChain = (index) => {
        if (processingChains.value.length <= 1) return

        processingChains.value.splice(index, 1)
        if (activeChainIndex.value >= processingChains.value.length) {
          activeChainIndex.value = processingChains.value.length - 1
        }
        focusedProcessorIndex.value = null
      }

      // 添加处理器
      const addProcessor = () => {
        if (!selectedProcessorType.value) return

        const newProcessor = {
          id: Date.now(),
          type: selectedProcessorType.value,
          config: getDefaultConfig(selectedProcessorType.value)
        }

        currentChain.value.processors.push(newProcessor)
        selectedProcessorType.value = ''
        processAllChains()
      }

      // 移除处理器
      const removeProcessor = (index) => {
        currentChain.value.processors.splice(index, 1)
        focusedProcessorIndex.value = null
        processAllChains()
      }

      // 移动处理器
      const moveProcessor = (index, direction) => {
        const newIndex = index + direction
        if (newIndex < 0 || newIndex >= currentChain.value.processors.length) return

        const processor = currentChain.value.processors.splice(index, 1)[0]
        currentChain.value.processors.splice(newIndex, 0, processor)

        if (focusedProcessorIndex.value === index) {
          focusedProcessorIndex.value = newIndex
        }

        processAllChains()
      }

      // 聚焦处理器
      const focusProcessor = (index) => {
        focusedProcessorIndex.value = index
      }

      // 切换处理器配置面板的展开/收起状态
      const toggleProcessorConfig = (processorIndex) => {
        const key = `${activeChainIndex.value}_${processorIndex}`
        expandedConfigs.value[key] = !expandedConfigs.value[key]
      }

      // 检查处理器配置面板是否展开
      const isConfigExpanded = (processorIndex) => {
        const key = `${activeChainIndex.value}_${processorIndex}`
        // 默认展开
        return expandedConfigs.value[key] !== false
      }

      // 处理拖动开始
      const startResize = (e) => {
        isResizing.value = true
        e.preventDefault()

        const handleMouseMove = (e) => {
          if (!isResizing.value) return

          // 获取链内容容器的尺寸
          const chainContent = document.querySelector('.chain-content')
          if (!chainContent) return

          const containerWidth = chainContent.offsetWidth
          const x = e.clientX
          const rect = chainContent.getBoundingClientRect()
          const relativeX = x - rect.left

          // 计算左右面板的宽度百分比，确保最小宽度
          const newLeftWidth = Math.max(30, Math.min(70, (relativeX / containerWidth) * 100))
          processorsWidth.value = `${newLeftWidth}%`
          previewWidth.value = `${100 - newLeftWidth}%`
        }

        const handleMouseUp = () => {
          isResizing.value = false
          document.removeEventListener('mousemove', handleMouseMove)
          document.removeEventListener('mouseup', handleMouseUp)
        }

        document.addEventListener('mousemove', handleMouseMove)
        document.addEventListener('mouseup', handleMouseUp)
      }

      // 处理器配置更新
      const onProcessorConfigUpdate = (processorIndex, newConfig) => {
        if (currentChain.value.processors[processorIndex]) {
          currentChain.value.processors[processorIndex].config = newConfig
          processAllChains()
        }
      }

      // 获取默认配置 - 与处理器组件的默认配置保持一致
      const getDefaultConfig = (type) => {
        switch (type) {
          case 'splitter':
            return {
              mode: 'string',
              separator: ' ',
              separatorHandling: 'keep',
              replacement: '',
              positions: [],
              outputMode: 'all',
              selectedFields: [],
              outputSeparator: ''
            }
          case 'inserter':
            return {
              mode: 'character',
              targetChar: '',
              targetPosition: 1,
              position: 'before',
              insertText: '',
              global: true
            }
          case 'formatter':
            return {
              mode: 'firstUpper',
              scope: 'line',
              preserveWhitespace: true,
              skipEmptyLines: true
            }
          case 'sorter':
            return {
              sortMode: 'dictionary',
              fieldSeparator: ':',
              sortFields: [1],
              ascending: true,
              ignoreCase: true,
              stable: false,
              trimFields: true,
              removeDuplicates: false,
              emptyHandling: 'last'
            }
          default:
            return {}
        }
      }

      // 切换输入区域
      const toggleInputSection = () => {
        isInputCollapsed.value = !isInputCollapsed.value
      }

      // 复制输出结果到剪贴板
      const copyOutputToClipboard = async (event) => {
        if (focusedProcessorIndex.value === null) return

        try {
          const outputText = getProcessorOutput(focusedProcessorIndex.value)
          await navigator.clipboard.writeText(outputText)

          // 简单的提示效果
          const btn = event.target
          const originalText = btn.textContent
          btn.textContent = '已复制!'
          btn.classList.add('copied')

          setTimeout(() => {
            btn.textContent = originalText
            btn.classList.remove('copied')
          }, 2000)
        } catch (error) {
          console.error('复制失败:', error)
          alert('复制失败，请手动复制')
        }
      }

      // 从剪贴板粘贴
      const pasteFromClipboard = async () => {
        try {
          const text = await navigator.clipboard.readText()
          inputText.value = text
          processAllChains()
        } catch (error) {
          console.error('粘贴失败:', error)
          alert('粘贴失败，请手动粘贴')
        }
      }

      // 打开文件
      const openFile = () => {
        if (fileInput.value) {
          fileInput.value.click()
        }
      }

      // 处理文件选择
      const handleFileSelect = (event) => {
        const file = event.target.files[0]
        if (file) {
          const reader = new FileReader()
          reader.onload = (e) => {
            inputText.value = e.target.result
            processAllChains()
          }
          reader.readAsText(file)
          // 重置文件输入，以便可以选择同一个文件
          event.target.value = ''
        }
      }

      // 清空输入
      const clearInput = () => {
        inputText.value = ''
        processAllChains()
      }

      // 监听输入文本变化
      watch(inputText, () => {
        processAllChains()
      })

      // 初始化
      processAllChains()

      return {
        inputText,
        isInputCollapsed,
        fileInput,
        processingChains,
        activeChainIndex,
        currentChain,
        selectedProcessorType,
        focusedProcessorIndex,
        processorResults,
        getProcessorTypeName,
        getProcessorConfigComponent,
        getProcessorInput,
        getProcessorOutput,
        switchToChain,
        addNewChain,
        removeChain,
        addProcessor,
        removeProcessor,
        moveProcessor,
        focusProcessor,
        onProcessorConfigUpdate,
        toggleInputSection,
        pasteFromClipboard,
        openFile,
        handleFileSelect,
        clearInput,
        copyOutputToClipboard,
        toggleProcessorConfig,
        isConfigExpanded,
        processorsWidth,
        previewWidth,
        isResizing,
        startResize
      }
    }
  }
</script>

<style scoped>
  .chained-text-processor {
    display: flex;
    flex-direction: column;
    height: 85vh;
    background: var(--bg-page);
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    overflow: hidden;
  }

  /* 输入区域 */
  .input-section {
    background: var(--bg-container);
    border-bottom: 1px solid var(--border-color-light);
    transition: all 0.3s ease;
  }

  .input-section.collapsed {
    height: 50px;
    overflow: hidden;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 15px 20px;
    background: var(--bg-muted);
    border-bottom: 1px solid var(--border-color-light);
  }

  .section-header h3 {
    margin: 0;
    color: var(--text-main);
    font-size: 16px;
  }

  .toggle-btn {
    background: transparent;
    border: 1px solid var(--border-color-light);
    border-radius: 4px;
    padding: 4px 12px;
    cursor: pointer;
    font-size: 12px;
    color: var(--text-secondary);
    transition: all 0.2s ease;
  }

  .toggle-btn:hover {
    background: var(--bg-hover);
    border-color: var(--border-color);
  }

  .input-content {
    padding: 20px;
  }

  .input-actions {
    display: flex;
    gap: 10px;
    margin-bottom: 15px;
  }

  .input-textarea {
    width: 100%;
    min-height: 120px;
    padding: 15px;
    border: 1px solid var(--border-color-light);
    border-radius: 8px;
    font-family: 'Courier New', monospace;
    font-size: 14px;
    resize: vertical;
    box-sizing: border-box;
  }

  .input-textarea:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.25);
  }

  /* 处理区域 */
  .processing-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* 链标签页 */
  .chain-tabs {
    background: var(--bg-container);
    border-bottom: 1px solid var(--border-color-light);
  }

  .tabs-header {
    display: flex;
    gap: 5px;
    padding: 10px 20px;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border: 1px solid var(--border-color-light);
    border-radius: 6px;
    background: var(--bg-container);
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 14px;
    color: var(--text-secondary);
  }

  .tab:hover {
    background: var(--bg-hover);
  }

  .tab.active {
    background: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
  }

  .close-tab-btn {
    background: transparent;
    border: none;
    font-size: 18px;
    cursor: pointer;
    color: inherit;
    padding: 0 4px;
    line-height: 1;
  }

  .close-tab-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 50%;
  }

  .add-chain-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border: 1px dashed var(--border-color-light);
    border-radius: 6px;
    background: transparent;
    cursor: pointer;
    font-size: 20px;
    color: var(--text-secondary);
    transition: all 0.2s ease;
  }

  .add-chain-btn:hover {
    background: var(--bg-hover);
    border-color: var(--color-primary);
    color: var(--color-primary);
  }

  /* 链内容 */
  .chain-content {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  /* 处理器容器 */
  .processors-container {
    background: var(--bg-container);
    border-right: 1px solid var(--border-color-light);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    transition: width 0.1s ease;
  }

  .processors-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 15px 20px;
    border-bottom: 1px solid var(--border-color-light);
    background: var(--bg-muted);
  }

  .processors-header h3 {
    margin: 0;
    color: var(--text-main);
    font-size: 16px;
  }

  .processor-selector {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .processor-selector .form-select {
    min-width: 150px;
  }

  /* 处理器列表 */
  .processors-list {
    flex: 1;
    overflow-y: auto;
    padding: 15px;
  }

  .processor-card {
    background: var(--bg-muted);
    border: 1px solid var(--border-color-light);
    border-radius: 8px;
    margin-bottom: 15px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .processor-card:hover {
    border-color: var(--color-primary);
    box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
  }

  .processor-card.active {
    border-color: var(--color-primary);
    background: #e3f2fd;
    box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
  }

  .processor-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 15px;
    border-bottom: 1px solid var(--border-color-light);
  }

  .processor-index {
    background: var(--color-primary);
    color: white;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: bold;
    flex-shrink: 0;
  }

  .processor-name {
    flex: 1;
    font-weight: 500;
    color: var(--text-main);
    font-size: 14px;
  }

  .processor-actions {
    display: flex;
    gap: 5px;
  }

  .processor-actions .btn {
    min-width: 32px;
    height: 32px;
    padding: 0;
    font-size: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .processor-config {
    padding: 15px;
    overflow: hidden;
    transition: max-height 0.3s ease, padding 0.3s ease;
  }

  /* 预览容器 */
  .preview-container {
    background: var(--bg-container);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    transition: width 0.1s ease;
  }

  /* 可拖动分隔线 */
  .resizer {
    width: 4px;
    background-color: #e0e0e0;
    cursor: col-resize;
    position: relative;
    user-select: none;
    flex-shrink: 0;
  }

  .resizer:hover {
    background-color: #007bff;
  }

  .resizer.resizing {
    background-color: #007bff;
    cursor: col-resize;
  }

  /* 分隔线悬停效果 */
  .resizer::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 20px;
    height: 60px;
    background-color: rgba(0, 123, 255, 0.2);
    border-radius: 4px;
    transform: translate(-50%, -50%);
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .resizer:hover::before {
    opacity: 1;
  }

  /* 确保链内容区域正确处理子元素 */
  .chain-content {
    display: flex;
    height: 100%;
  }

  .preview-container h3 {
    margin: 0;
    padding: 15px 20px;
    border-bottom: 1px solid var(--border-color-light);
    background: var(--bg-muted);
    color: var(--text-main);
    font-size: 16px;
  }

  .preview-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }

  /* 预览区域样式 */
  .preview-section {
    margin-bottom: 20px;
  }

  .section-header-with-action {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .preview-section h4 {
    margin: 0;
    color: #666;
    font-size: 14px;
  }

  .copy-btn {
    padding: 4px 10px;
    background: var(--color-success);
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .copy-btn:hover {
    background: var(--color-success-hover);
  }

  .copy-btn.copied {
    background: var(--color-info);
  }

  .preview-text {
    background: var(--bg-muted);
    border: 1px solid var(--border-color-light);
    border-radius: 6px;
    padding: 15px;
    font-family: 'Courier New', monospace;
    font-size: 13px;
    line-height: 1.4;
    white-space: pre-wrap;
    word-wrap: break-word;
    max-height: 200px;
    overflow-y: auto;
  }

  .no-preview {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-size: 14px;
  }

  /* 空状态 */
  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: #999;
  }

  .empty-state p {
    margin: 0;
    font-size: 14px;
  }

  /* 按钮样式 */
  .btn {
    padding: 8px 16px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s ease;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
  }

  .btn:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }

  .btn:active {
    transform: scale(0.98) translateY(-0.5px);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  }

  .btn-primary {
    background: var(--color-primary);
    color: white;
  }

  .btn-primary:hover {
    background: var(--color-primary-hover);
  }

  .btn-secondary {
    background: #6c757d;
    color: white;
  }

  .btn-secondary:hover {
    background: #545b62;
  }

  .btn-danger {
    background: var(--color-danger);
    color: white;
  }

  .btn-danger:hover {
    background: var(--color-danger-hover);
  }

  .btn-sm {
    padding: 4px 8px;
    font-size: 12px;
  }

  .btn-outline-secondary {
    background: transparent;
    color: var(--text-secondary);
    border: 1px solid var(--border-color-light);
  }

  .btn-outline-secondary:hover {
    background: var(--bg-hover);
    border-color: var(--border-color);
  }

  /* 表单元素样式 */
  .form-select {
    padding: 8px 12px;
    border: 1px solid var(--border-color-light);
    border-radius: 4px;
    font-size: 14px;
    background: var(--bg-container);
  }

  .form-select:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.25);
  }

  /* 滚动条样式 */
  .processors-list::-webkit-scrollbar,
  .preview-content::-webkit-scrollbar,
  .preview-text::-webkit-scrollbar {
    width: 6px;
  }

  .processors-list::-webkit-scrollbar-track,
  .preview-content::-webkit-scrollbar-track,
  .preview-text::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 3px;
  }

  .processors-list::-webkit-scrollbar-thumb,
  .preview-content::-webkit-scrollbar-thumb,
  .preview-text::-webkit-scrollbar-thumb {
    background: #adb5bd;
    border-radius: 3px;
  }

  .processors-list::-webkit-scrollbar-thumb:hover,
  .preview-content::-webkit-scrollbar-thumb:hover,
  .preview-text::-webkit-scrollbar-thumb:hover {
    background: #6c757d;
  }
</style>