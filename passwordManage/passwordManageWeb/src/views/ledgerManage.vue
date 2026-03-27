<!--
 * @Author: 小鱼
 * @Date: 2025-09-03 17:23:30
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-04 17:05:45
 * @FilePath: \passwordManageWeb\src\views\ledgerManage.vue
 * @Description: 台账管理页面，支持目录结构管理和表格数据编辑
 * 
 * Copyright (c) 2025 by 小鱼, All Rights Reserved. 
-->
<template>
  <div class="ledger-container">
    <!-- 头部区域 -->
    <div class="ledger-header">
      <h2>台账管理</h2>
      <div class="header-actions">
        <el-dropdown placement="bottom">
          <el-button type="primary">
            新建 <el-icon class="el-icon--right">
              <ArrowDown />
            </el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleNewProject">新建项目</el-dropdown-item>
              <el-dropdown-item @click="handleNewGroup">新建分组</el-dropdown-item>
              <el-dropdown-item @click="handleNewLedger">新建台账</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button type="default" @click="handleImportLedger">导入</el-button>
        <el-button type="default" @click="handleExportLedger"
          :disabled="!selectedNode || selectedNode.isFolder">导出</el-button>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="ledger-content">
      <!-- 左侧台账目录区域 -->
      <div class="ledger-directory-wrapper">
        <div class="ledger-directory" :class="{ 'collapsed': isDirectoryCollapsed }">
          <div class="directory-header">
            <div class="search-box">
              <el-input v-model="searchKeyword" placeholder="搜索台账" :prefix-icon="Search" size="small"
                @input="handleSearch" />
            </div>
            <div class="directory-actions">
              <el-button type="text" @click="handleRefreshTree" size="small" :icon="Refresh" title="刷新台账" />
              <el-button type="text" @click="toggleDirectory" size="small"
                :icon="isDirectoryCollapsed ? ArrowRight : ArrowLeft" title="收起目录" />
            </div>
          </div>

          <div class="directory-tree">
            <!-- 台账树状结构 -->
            <VTree ref="treeRef" @click="handleNodeClick" @node-right-click="handleNodeRightClick" keyField="id"
              titleField="name" childrenField="children" :isLeaf="node => !node.isFolder" :draggable="true"
              :droppable="true" :beforeDropMethod="handleBeforeDrop" :nodeClassName="setNodeClassName"
              :expandedKeys="expandedKeys">
            </VTree>
          </div>
        </div>
        <!-- 展开按钮 -->
        <div v-if="isDirectoryCollapsed" class="expand-button-container">
          <el-button type="text" @click="toggleDirectory" size="small" :icon="ArrowRight" title="展开目录" />
        </div>
      </div>

      <!-- 右侧台账数据区域 -->
      <div class="ledger-data">
        <div v-if="selectedNode && !selectedNode.isFolder&&gridOptions.data.length>0" class="data-header">
          <h3>{{ selectedNode.name }}</h3>
          <div class="data-actions">
            <el-button type="primary" @click="handleAddRecord" size="small" :disabled="!isEdit">新增记录</el-button>
            <el-button type="primary" @click="handleEditLedger(gridOptions.schema.ledgerId)" size="small"
              :icon="Edit">编辑台账</el-button>
            <el-button type="default" @click="handleBatchImport" size="small">批量导入</el-button>
            <el-button type="default" @click="handleBatchExport" size="small">批量导出</el-button>
            <el-button :type="isEdit ? 'warning' : 'success'" @click="toggleEditMode" size="small">
              {{ isEdit ? '退出编辑' : '进入编辑' }}
            </el-button>
            <!-- 编辑模式下显示保存和取消按钮 -->
            <template v-if="isEdit">
              <el-button type="primary" @click="handleSaveChanges" size="small" :icon="Check">保存</el-button>
              <el-button type="default" @click="handleCancelChanges" size="small" :icon="Close">取消</el-button>
            </template>
          </div>
        </div>

        <!-- 编辑模式提示 -->
        <div v-if="isEdit" class="edit-warning"
          style="padding: 8px 16px; background-color: #fff3cd; color: #856404; border-radius: 4px; margin: 0 16px 16px;">
          ⚠️ 注意：当前处于编辑模式，请及时保存您的修改！
        </div>

        <!-- 台账数据表格 -->
        <div class="account-content">
          <template v-if="selectedNode && !selectedNode.isFolder">
            <vxe-grid ref="gridRef"
              v-bind="{...gridOptions.options,columns:gridOptions.columns,data:gridOptions.data || []}"
              @cell-click="onCellClick" @menu-click="handleMenuClick">
              <template #toolbarButtons>
                <vxe-button @click="handleAddRecord" type="primary" size="small">新增记录</vxe-button>
              </template>
              <template #operation="{ row }">
                <vxe-button @click="handleEditRow(row)" type="text" size="small" status="primary">编辑</vxe-button>
                <vxe-button @click="handleDeleteRow(row)" type="text" size="small" status="danger">删除</vxe-button>
              </template>
            </vxe-grid>
          </template>
          <template v-else>
            <div class="account-section">
              <el-empty description="请选择一个台账查看数据"></el-empty>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- 树结构右键菜单 -->
    <div v-show="contextMenuVisible" :style="{left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px'}"
      class="custom-context-menu">
      <div v-if="selectedNodeRightClick && selectedNodeRightClick._level > 0" class="context-menu-item"
        @click="handleRenameSelected">
        重命名
      </div>
      <div v-if="selectedNodeRightClick && selectedNodeRightClick._level === 0" class="context-menu-item"
        @click="handleNewProjectRight">
        新建项目
      </div>
      <div v-if="selectedNodeRightClick && selectedNodeRightClick.isFolder" class="context-menu-item"
        @click="handleNewGroupRight">
        新建分组
      </div>
      <div v-if="selectedNodeRightClick && selectedNodeRightClick.isFolder" class="context-menu-item"
        @click="handleNewLedger">
        新建台账
      </div>
      <div v-else-if="selectedNodeRightClick && selectedNodeRightClick.type === 'ledger'" class="context-menu-item"
        @click="handleEditLedger(selectedNodeRightClick.id)">
        编辑台账
      </div>
      <div v-if="selectedNodeRightClick && selectedNodeRightClick._level > 0" class="context-menu-item"
        @click="handleDeleteSelected">
        删除
      </div>
    </div>
  </div>

  <!-- 台账设置对话框 -->
  <el-dialog v-model="ledgerSetupDialogVisible" :title="isEditMode ? '编辑台账设置' : '新建台账设置'" width="800px"
    :destroy-on-close="true">
    <div class="ledger-setup-content">
      <!-- 台账基本信息 -->
      <div class="setup-section">
        <h4>基本信息</h4>
        <el-form label-width="80px">
          <el-form-item label="台账名称">
            <el-input v-model="ledgerName" placeholder="请输入台账名称" />
          </el-form-item>
          <el-form-item label="初始行数">
            <el-input-number v-model="rowCount" :min="1" :max="100" :step="1" />
          </el-form-item>
        </el-form>
      </div>

      <!-- 表格设置 -->
      <div class="setup-section">
        <h4>表格设置</h4>
        <el-form label-width="120px" size="small">
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="斑马纹">
                <el-switch v-model="tableSettings.stripe" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="内容超出省略">
                <el-switch v-model="tableSettings.showOverflow" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="表头超出省略">
                <el-switch v-model="tableSettings.showHeaderOverflow" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="列宽可调整">
                <el-switch v-model="tableSettings.columnResizable" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="列可移动">
                <el-switch v-model="tableSettings.columnDraggable" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="行悬停效果">
                <el-switch v-model="tableSettings.rowHover" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="高亮当前行">
                <el-switch v-model="tableSettings.highlightCurrentRow" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="显示边框">
                <el-switch v-model="tableSettings.border" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="显示工具栏">
                <el-switch v-model="tableSettings.showToolbar" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="显示表尾">
                <el-switch v-model="tableSettings.showFooter" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="编辑触发方式">
                <el-select v-model="tableSettings.editTrigger" placeholder="请选择">
                  <el-option v-for="trigger in editTriggers" :key="trigger.value" :label="trigger.label"
                    :value="trigger.value" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="编辑模式">
                <el-select v-model="tableSettings.editMode" placeholder="请选择">
                  <el-option v-for="mode in editModes" :key="mode.value" :label="mode.label" :value="mode.value" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </div>

      <!-- 列配置 -->
      <div class="setup-section">
        <div class="section-header">
          <h4>列配置</h4>
          <el-button @click="addColumn" type="primary" size="small">
            <el-icon>
              <Plus />
            </el-icon> 添加列
          </el-button>
        </div>

        <div class="columns-config">
          <template v-for="(column, index) in columnsConfig" :key="index">
            <div class="column-item">
              <div class="column-header">
                <span>列 {{ index + 1 }}</span>
                <div class="column-actions">
                  <el-button @click="insertColumnBefore(index)" type="text" size="small" title="在前面插入">
                    <el-icon>
                      <ArrowLeft />
                    </el-icon> 前面插入
                  </el-button>
                  <el-button @click="insertColumnAfter(index)" type="text" size="small" title="在后面插入">
                    <el-icon>
                      <ArrowRight />
                    </el-icon> 后面插入
                  </el-button>
                  <el-button @click="removeColumn(index)" type="text" size="small" danger
                    v-if="columnsConfig.length > 1">
                    <el-icon>
                      <Delete />
                    </el-icon> 删除
                  </el-button>
                </div>
              </div>

              <el-form label-width="80px">
                <el-form-item label="列名">
                  <el-input v-model="column.title" placeholder="请输入列名" />
                </el-form-item>
                <el-form-item label="字段名">
                  <el-input v-model="column.field" placeholder="请输入字段名" />
                </el-form-item>
                <el-form-item label="数据类型">
                  <el-select v-model="column.type" placeholder="请选择数据类型">
                    <el-option v-for="type in dataTypes" :key="type.value" :label="type.label" :value="type.value" />
                  </el-select>
                  <el-button v-if="column.type === 'select'" @click="generateSelectOptions(index)" type="primary"
                    size="small" style="margin-left: 10px">
                    设置选项
                  </el-button>
                </el-form-item>
                <el-form-item label="列宽">
                  <el-input-number v-model="column.width" :min="40" :max="500" :step="10" />
                </el-form-item>
                <el-form-item label="是否必填">
                  <el-switch v-model="column.required" />
                </el-form-item>
              </el-form>
            </div>
            <div v-if="index < columnsConfig.length - 1" class="divider"></div>
          </template>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="ledgerSetupDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="confirmCreateLedger">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import utils from '@/js/utils.js'
  import { ref, reactive, computed, onMounted, onUnmounted, h } from 'vue'
  import { ElButton, ElInput, ElDropdown, ElDropdownItem, ElMessageBox, ElMessage, ElEmpty } from 'element-plus'
  import { Search, Plus, Download, Upload, Edit, Delete, ArrowDown, Check, Close, ArrowLeft, ArrowRight, Refresh } from '@element-plus/icons-vue'
  import 'element-plus/dist/index.css'
  import VXETable from 'vxe-table'
  import 'vxe-table/lib/style.css'
  import { VxeGrid } from 'vxe-table'
  import XEUtils from 'xe-utils'
  import VTree from '@wsfe/vue-tree'
  import '@wsfe/vue-tree/style.css'
  import http from '@/js/http.js'
  import httpTest from '@/js/httpApi.js'

  const treeRef = ref(null) // 树状结构引用
  const expandedKeys = ref(['root']); // 默认展开
  const gridRef = ref(null) // 台账表格引用
  const tableSetting = ref() // 表格配置
  const columnsConfig = ref([]) // 列配置
  const searchKeyword = ref('') // 搜索关键词
  const selectedNode = ref(null) // 当前选中的节点
  const selectedNodeRightClick = ref(null) // 当前选中的右键节点
  const tableHeight = ref('400px') // 表格高度

  const contextMenuVisible = ref(false) // 上下文菜单是否可见
  const contextMenuPosition = reactive({ x: 0, y: 0 }) // 上下文菜单位置
  const isDirectoryCollapsed = ref(false) // 左侧目录板块收起/展开
  const currentCell = ref(null) // 当前点击的单元格数据

  // 行右键菜单相关
  const selectedRow = ref(null) // 当前选中的行数据
  const selectedRowIndex = ref(-1) // 当前选中的行索引
  const rowContextMenuPosition = reactive({ x: 0, y: 0 }) // 行右键菜单位置

  // 编辑模式控制
  const isEdit = ref(false) // 默认非编辑模式

  // 表格选项配置
  const gridOptions = reactive({
    /** ========= 1. 表格配置 ========= */
    options: {
      // 基础外观
      height: '100%',
      minHeight: 400,
      maxHeight: 800,
      stripe: false,
      border: true,
      round: false,
      size: 'medium', // 表格尺寸 medium,small,mini
      loading: true, // 是否显示加载中
      align: 'center', // 所有列对齐方式 center,left,right
      headerAlign: 'center', // 表头对齐方式 center,left,right
      footerAlign: 'center', // 表尾对齐方式 center,left,right
      showHeader: false, // 是否显示表头
      showFooter: false, // 是否显示表尾
      footerData: [], // 表尾数据
      showOverflow: true, // 所有内容过长显示省略号
      showHeaderOverflow: true, // 表头内容过长显示省略号
      showFooterOverflow: true,
      columnConfig: {
        isCurrent: true, // 点击列头高亮当前列
        isHover: true, // 鼠标悬停高亮
        width: 'auto', // 列宽度
        minWidth: 100, // 最小宽度
        maxFixedSize: 2,// 冻结列允许设置的最大数量
        drag: true, // 列拖动
      },
      currentColumnConfig: {
        trigger: 'click', // 点击触发
      },
      cellConfig: {  // 单元格配置
        padding: true, // 是否显示单元格间距
        height: 50, // 单元格高度
        verticalAlign: 'center', // 垂直对齐方式 top,center
      },
      headerCellConfig: {  // 表头单元格配置
        padding: true, // 是否显示单元格间距
        height: 50, // 单元格高度
      },
      footerCellConfig: {  // 表尾单元格配置
        padding: true, // 是否显示单元格间距
        height: 50, // 单元格高度
      },
      rowConfig: {  // 行配置
        isHover: true, // 鼠标悬停高亮
        isCurrent: true, // 选中行高亮
        drag: false, // 行拖动
      },
      currentRowConfig: {  // 当前行配置
        trigger: 'default', // 触发 default/row/manual
      },
      resizableConfig: {  // 列宽调整配置
        dragMode: 'auto', // 拖拽模式 auto/fixed
        isDblclickAutoWidth: true, // 双击自动调整宽度
        isDblclickAutoHeight: true, // 双击自动调整高度
      },
      sortConfig: {  // 排序配置
        ascTitle: '升序', // 升序标题
        descTitle: '降序', // 降序标题
        defaultSort: {
          field: 'id',
          order: 'asc'
        },
        remote: false, // 是否远程排序
        allowClear: true, // 一致时取消排序 只对allowBth有效
        trigger: 'default', // 触发方式 default/row/manual
        allowBth: true, //点击排序图标进行操作
        showIcon: true, // 是否显示列头排序图标
        iconLayout: 'vertical', // 排序图标布局 vertical/horizontal
      },
      rowDragConfig: {  // 行拖动配置
        trigger: 'row', // 触发方式 default/row/manual
        showIcon: false, // 是否显示行头拖动图标
        animation: true, // 是否启用动画
      },
      columnDragConfig: {  // 列拖动配置
        trigger: 'default', // 触发方式 default/cell
        showIcon: false, // 是否显示列头拖动图标
        animation: true, // 是否启用动画
      },
      filterConfig: {  // 过滤配置
        isEvery: false, // 每个列都应用过滤
        remote: false, // 是否远程过滤
        filterMethod: function (row, column, filterValue) {
          return row[column.property] === filterValue
        },
        showIcon: true, // 是否显示列头过滤图标
      },
      floatingFilterConfig: {  // 浮动过滤配置
        enabled: false, // 是否启用浮动过滤
        floatingFilters: false, // 是否显示浮动过滤
      },
      // 编辑配置
      editConfig: {
        enabled: true, // 是否启用编辑
        mode: 'cell', // 支持单元格编辑 cell/row
        trigger: 'dblclick', // 触发方式 manual/click/dblclick
        showStatus: true,
        autoClear: true
      },

      // 分页配置
      pagerConfig: {
        enabled: true,
        pageSize: 20,
        pageSizes: [10, 20, 50, 100],
        layouts: ['PrevPage', 'JumpNumber', 'NextPage', 'Sizes', 'Total']
      },

      // 工具栏
      toolbarConfig: {
        enabled: true,
        refresh: true,
        custom: true,
        export: false,
        print: false,
        slots: {}
      },

      // 右键菜单
      menuConfig: {
        enabled: true, // 是否启用右键菜单
        trigger: 'default', // 触发方式 default/cell
        header: { disabled: true, options: [] }, // 表头右键菜单
        body: { // 表格体右键菜单
          disabled: false, // 是否禁用
          options: [
            [
              { code: 'insertBefore', name: '在前面插入' },
              { code: 'insertAfter', name: '在后面插入' },
              { code: 'remove', name: '删除' }]
          ]
        },
        footer: { disabled: true, options: [] }
      },
    },

    /** ========= 2. 列定义 ========= */
    columns: [],

    /** ========= 3. 数据 ========= */
    data: [], // 深度响应式，单元格修改自动触发更新
    schema: {
      ledgerId: '',
    }
  })

  // ########################################## 树相关###################################
  // 在组件挂载时从API获取树形数据
  onMounted(async () => {
    try {
      // 调用API获取台账树数据
      const response = await httpTest.get('/getLedgerTree')
      // 使用树组件的 setData 方法更新数据
      treeRef.value.setData(response.data || [])
    } catch (error) {
      console.error('获取台账树数据失败:', error)
      ElMessage({
        message: '获取台账树数据失败',
        type: 'error',
        grouping: true
      })
    }
  })

  // 处理节点点击事件，加载数据
  const handleNodeClick = async (data) => {
    selectedNode.value = data
    // 切换台账时关闭编辑模式
    isEdit.value = false

    // 如果点击的是文件夹，不显示数据
    if (data.isFolder) {
      if (gridRef.value) {
        gridRef.value.clearSelection()
      }
      return
    }
    // 加载台账数据
    gridOptions.options.loading = true
    try {
      const [tableCfgRes, columnRes, dataRes] = await Promise.all([
        getTableSetting(data.id),
        getColumnSetting(data.id),
        getLedgerData(data.id)
      ])
      if (tableCfgRes) {
        Object.assign(gridOptions.options, tableCfgRes)
      }
      gridOptions.columns = columnRes.columns || []
      gridOptions.data = dataRes.data || []
      gridOptions.schema.ledgerId = data.ledgerId
    } finally {
      gridOptions.options.loading = false
    }
    // 重新渲染表格
    if (gridRef.value) {
      gridRef.value.refreshColumn()
    }
  }

  // 获取表格配置
  const getTableSetting = async (id) => {
    try {
      const res = await httpTest.get(`/getLedgerTableSets?ledgerId=${id}`)
      return res.data
    } catch (error) {
      console.error('获取台账表格配置失败:', error)
      ElMessage({
        type: 'error',
        message: '获取台账表格配置失败'
      })
      return null  // 出错返回 null，保证调用方安全
    }
  }

  const getColumnSetting = async (id) => {
    try {
      const res = await httpTest.get(`/getLedgerColumnSets?ledgerId=${id}`)
      return res.data
    } catch (err) {
      console.error(err)
      return { columns: [] }
    }
  }

  const getLedgerData = async (id) => {
    try {
      const res = await httpTest.get(`/getLedgerData?ledgerId=${id}`)
      return res.data
    } catch (err) {
      console.error(err)
      return { data: [] }
    }
  }

  const updateTableHeight = () => {
    tableHeight.value = '100%'
  }

  // 刷新台账树数据
  const handleRefreshTree = async () => {
    try {
      // 调用API获取台账树数据
      const response = await httpTest.get('/getLedgerTree')
      // 使用树组件的 setData 方法更新数据
      treeRef.value.setData(response.data || [])
      ElMessage({
        message: '台账树刷新成功',
        type: 'success',
        grouping: true
      })
    } catch (error) {
      console.error('获取台账树数据失败:', error)
      ElMessage({
        message: '获取台账树数据失败',
        type: 'error',
        grouping: true
      })
    }
  }

  // 监听窗口大小变化（保留，以便在需要时恢复计算逻辑）
  onMounted(() => {
    updateTableHeight()
    window.addEventListener('resize', updateTableHeight)
    document.addEventListener('click', handleGlobalClick)
  })

  // 组件卸载时移除事件监听器
  onUnmounted(() => {
    window.removeEventListener('resize', updateTableHeight)
    document.removeEventListener('click', handleGlobalClick)
  })

  // 处理搜索
  const handleSearch = () => {
  }

  // 切换目录展开/收起状态
  const toggleDirectory = () => {
    isDirectoryCollapsed.value = !isDirectoryCollapsed.value
  }

  // 处理节点右键菜单
  const handleNodeRightClick = (data, event) => {
    // 设置选中节点
    selectedNodeRightClick.value = data
    // 阻止浏览器默认右键菜单
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault()
    }
    // 使用事件对象中的位置信息
    if (event) {
      contextMenuPosition.x = event.clientX
      contextMenuPosition.y = event.clientY
      contextMenuVisible.value = true
    }
  }

  //////////////////////////////////////////// 表格相关/////////////////////////////////////////
  // 加载台账数据
  const loadLedgerData = async (ledger) => {
    // 加载台账数据
    gridOptions.options.loading = true
    try {
      // 调用API获取数据
      const [tableCfgRes, columnRes, dataRes] = await Promise.all([
        getTableSetting(ledger.id),
        getColumnSetting(ledger.id),
        getLedgerData(ledger.id)
      ])

      // 应用表格配置
      if (tableCfgRes) {
        Object.assign(gridOptions.options, tableCfgRes)
      }

      // 处理列配置，添加序号列和操作列
      const columns = columnRes.columns || []

      // 为列添加筛选渲染配置
      const columnsWithFilter = columns.map(column => ({
        ...column,
        // 筛选渲染配置
        filterRender: {
          name: 'VxeInput',
          props: {
            clearable: true
          }
        }
      }))

      // 添加序号列和操作列
      const finalColumns = [
        { field: 'seq', type: 'seq', width: 60, align: 'center' },
        ...columnsWithFilter,
        {
          field: 'operation',
          title: '操作',
          width: 120,
          fixed: 'right',
          slots: {
            default: 'operation'
          }
        }
      ]

      // 更新表格配置
      gridOptions.columns = finalColumns
      gridOptions.data = dataRes.data || []
    } catch (error) {
      console.error('加载台账数据失败:', error)
      ElMessage({
        type: 'error',
        message: '加载台账数据失败'
      })
    } finally {
      gridOptions.options.loading = false
    }

    // 重新渲染表格
    if (gridRef.value) {
      gridRef.value.refreshColumn()
    }
  }


  // 处理重命名选中项
  const handleRenameSelected = () => {
    if (!selectedNodeRightClick.value) {
      ElMessage({
        message: '请先选择要重命名的项目、分组或台账',
        type: 'warning',
        grouping: true,
      })
      contextMenuVisible.value = false
      return
    }

    let title = '重命名'
    let promptText = '请输入新名称'

    // 根据节点类型显示不同的提示文本
    if (selectedNodeRightClick.value.type === 'project') {
      promptText = '请输入项目新名称'
      title = '重命名项目'
    } else if (selectedNodeRightClick.value.type === 'group') {
      promptText = '请输入分组新名称'
      title = '重命名分组'
    } else if (selectedNodeRightClick.value.type === 'ledger') {
      promptText = '请输入台账新名称'
      title = '重命名台账'
    }

    ElMessageBox.prompt(promptText, title, {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^\S+$/,
      inputErrorMessage: '名称不能为空',
      inputValue: selectedNodeRightClick.value.name
    }).then(({ value }) => {
      if (selectedNodeRightClick.value._level > 0) {
        if (utils.isDuplicateInSameLevel(selectedNodeRightClick.value._parent, value)) {
          ElMessage({
            message: '名称重复',
            type: 'error',
            grouping: true,
          })
          return
        }
        // 发送请求重命名更改，暂时只更新本地树
        treeRef.value.updateNode(selectedNodeRightClick.value.id, { name: value })
        ElMessage({
          message: '重命名成功',
          type: 'success',
          grouping: true,
        })

      }
    }).catch(() => {
      // 取消操作
    }).finally(() => {
      // 关闭右键菜单
      contextMenuVisible.value = false
    })
  }

  // 添加全局点击事件来关闭右键菜单
  const handleGlobalClick = () => {
    contextMenuVisible.value = false
  }

  // 处理台账右键菜单事件
  const handleMenuClick = ({ code, row, rowIndex }) => {
    selectedRow.value = row
    selectedRowIndex.value = rowIndex

    console.log('rowIndex:', rowIndex)

    if (code === 'remove') {
      dialogVisible.value = true
    }
  }

  // 在选中行前面插入记录
  const handleInsertRowBefore = () => {
    // 检查当前台账和选中行是否存在
    if (!selectedNode.value || selectedRowIndex.value === -1) return

    const newRecord = {
      id: Date.now()
    }

    // 为每个列添加默认值
    gridOptions.columns.forEach(column => {
      // 跳过序号列和操作列
      if (column.field !== 'seq' && column.field !== 'operation') {
        newRecord[column.field] = ''
      }
    })

    gridOptions.data.splice(selectedRowIndex.value, 0, newRecord)

    ElMessage({
      message: '插入记录成功',
      type: 'success',
      grouping: true,
    })
  }

  // 在选中行后面插入记录
  const handleInsertRowAfter = () => {
    // 检查当前台账和选中行是否存在
    if (!selectedNode.value || selectedRowIndex.value === -1) return

    const newRecord = {
      id: Date.now()
    }

    // 为每个列添加默认值
    gridOptions.columns.forEach(column => {
      // 跳过序号列和操作列
      if (column.field !== 'seq' && column.field !== 'operation') {
        newRecord[column.field] = ''
      }
    })

    gridOptions.data.splice(selectedRowIndex.value + 1, 0, newRecord)

    ElMessage({
      message: '插入记录成功',
      type: 'success',
      grouping: true,
    })
  }

  // 处理新建项目
  const handleNewProject = () => {
    ElMessageBox.prompt('请输入项目名称', '新建项目', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^\S+$/,
      inputErrorMessage: '项目名称不能为空'
    }).then(({ value }) => {
      const rootNode = treeRef.value.getNode("root")
      // 检查项目名称是否重复
      if (utils.isDuplicateInSameLevel(rootNode, value)) {
        ElMessage({
          message: '项目名称已存在',
          type: 'warning',
          grouping: true,
        })
        return
      }
      const newProject = {
        id: Date.now().toString(),
        name: value,
        isFolder: true,
        type: 'project', // 项目类型
        children: []
      }
      //获取项目节点数
      const nodeCount = treeRef.value.getNodesCount()
      if (nodeCount === 0) {
        ElMessage({
          message: '未加载到数据，检查网络情况！',
          type: 'warning',
          grouping: true,
        })
        return
      }
      treeRef.value.prepend(newProject, rootNode.id)
      ElMessage({
        message: '新建项目成功',
        type: 'success',
        grouping: true,
      })
    }).catch(() => {
      // 取消操作
    })
  }

  const handleNewProjectRight = () => {
    handleNewProject()
  }
  // 处理新建分组
  const handleNewGroup = () => {
    // 检查是否有选中的节点且是文件夹
    if (!selectedNode.value || !selectedNode.value.isFolder) {
      ElMessage({
        message: '请先选择一个项目或分组',
        type: 'warning',
        grouping: true,
      })
      return
    }

    ElMessageBox.prompt('请输入分组名称', '新建分组', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^\S+/,
      inputErrorMessage: '分组名称不能为空'
    }).then(({ value }) => {
      // 检查分组名称是否重复
      if (utils.isDuplicateInSameLevel(selectedNode.value, value)) {
        ElMessage({
          message: '分组名称已存在',
          type: 'warning',
          grouping: true,
        })
        return
      }
      const newGroup = {
        id: Date.now().toString(),
        name: value,
        isFolder: true,
        type: 'group', // 分组类型
        children: []
      }

      // 添加到选中的文件夹下
      treeRef.value.prepend(newGroup, selectedNode.value.id)
      ElMessage({
        message: '新建分组成功',
        type: 'success',
        grouping: true,
      })
    }).catch(() => {
      // 取消操作
    })
  }
  const handleNewGroupRight = () => {
    // 检查是否有选中的节点且是文件夹
    if (!selectedNodeRightClick.value || !selectedNodeRightClick.value.isFolder) {
      ElMessage({
        message: '请先选择一个项目或分组',
        type: 'warning',
        grouping: true,
      })
      return
    }

    ElMessageBox.prompt('请输入分组名称', '新建分组', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^\S+/,
      inputErrorMessage: '分组名称不能为空'
    }).then(({ value }) => {
      if (utils.isDuplicateInSameLevel(selectedNodeRightClick.value, value)) {
        ElMessage({
          message: '分组名称已存在',
          type: 'warning',
          grouping: true,
        })
        return
      }
      const newGroup = {
        id: Date.now().toString(),
        name: value,
        isFolder: true,
        type: 'group', // 分组类型
        children: []
      }

      // 添加到选中的文件夹下
      treeRef.value.prepend(newGroup, selectedNodeRightClick.value.id)
      ElMessage({
        message: '新建分组成功',
        type: 'success',
        grouping: true,
      })
    }).catch(() => {
      // 取消操作
    })
  }
  // 台账设置相关状态
  const ledgerSetupDialogVisible = ref(false)
  const isEditMode = ref(false) // 编辑模式标识
  const editingLedgerId = ref('') // 当前编辑的台账ID
  const ledgerName = ref('')
  const columnCount = ref(3)
  const rowCount = ref(10)
  const dataTypes = [
    { label: '文本', value: 'text' },
    { label: '数字', value: 'number' },
    { label: '日期', value: 'date' },
    { label: '下拉选择', value: 'select' }
  ]

  // 表格设置
  const tableSettings = ref({
    stripe: false,          // 斑马纹
    showOverflow: true,     // 内容超出显示省略号
    showHeaderOverflow: true, // 表头超出显示省略号
    columnResizable: true,  // 列宽可调整
    columnDraggable: true,  // 列可移动
    rowHover: true,         // 行悬停效果
    editTrigger: 'click',   // 编辑触发方式
    editMode: 'row',        // 编辑模式
    border: true,           // 边框显示
    showToolbar: true,      // 显示工具栏
    showFooter: false,      // 显示表尾
    highlightCurrentRow: true // 高亮当前行
  })

  const editTriggers = [
    { label: '点击', value: 'click' },
    { label: '双击', value: 'dblclick' },
    { label: '回车', value: 'enter' }
  ]

  const editModes = [
    { label: '行编辑', value: 'row' },
    { label: '单元格编辑', value: 'cell' }
  ]

  // 处理新建台账
  const handleNewLedger = () => {
    // 检查是否有选中的文件夹节点
    if (!selectedNode.value || !selectedNode.value.isFolder) {
      ElMessage({
        message: '请先选择一个项目或分组',
        type: 'warning',
        grouping: true,
      })
      return
    }

    // 重置设置
    isEditMode.value = false
    editingLedgerId.value = ''
    ledgerName.value = ''
    columnCount.value = 3
    rowCount.value = 10
    columnsConfig.value = [
      { field: 'id', title: 'ID', type: 'text', width: 60, required: true },
      { field: 'name', title: '名称', type: 'text', width: 120, required: true },
      { field: 'value', title: '值', type: 'text', width: 200, required: false }
    ]

    // 打开设置对话框
    ledgerSetupDialogVisible.value = true
  }

  // 处理编辑台账
  const handleEditLedger = (ledgerId) => {
    if (!gridOptions.data || gridOptions.data.length === 0 || selectedNode.value.type !== 'ledger') {
      ElMessage({
        message: '请先选择一个台账',
        type: 'warning',
        grouping: true,
      })
      return
    }

    // 设置编辑模式
    isEditMode.value = true

    // 加载现有配置
    ledgerName.value = selectedNode.value.name
    columnsConfig.value = JSON.parse(JSON.stringify(gridOptions.columns))

    // 设置表格设置
    if (gridOptions.tableSettings) {
      tableSettings.value = JSON.parse(JSON.stringify(gridOptions.tableSettings))
    }

    // 打开设置对话框
    ledgerSetupDialogVisible.value = true
  }

  // 添加新列
  const addColumn = () => {
    const newField = 'field' + (columnsConfig.value.length + 1)
    const newTitle = '列' + (columnsConfig.value.length + 1)
    columnsConfig.value.push({
      field: newField,
      title: newTitle,
      type: 'text',
      width: 120,
      required: false
    })
  }

  // 移除列
  const removeColumn = (index) => {
    if (columnsConfig.value.length <= 1) {
      ElMessage({
        message: '至少保留一列',
        type: 'warning',
        grouping: true,
      })
      return
    }
    columnsConfig.value.splice(index, 1)
  }

  // 在指定列前面插入新列
  const insertColumnBefore = (index) => {
    const newField = 'field' + (columnsConfig.value.length + 1)
    const newTitle = '列' + (columnsConfig.value.length + 1)
    const newColumn = {
      field: newField,
      title: newTitle,
      type: 'text',
      width: 120,
      required: false
    }
    columnsConfig.value.splice(index, 0, newColumn)
  }

  // 在指定列后面插入新列
  const insertColumnAfter = (index) => {
    const newField = 'field' + (columnsConfig.value.length + 1)
    const newTitle = '列' + (columnsConfig.value.length + 1)
    const newColumn = {
      field: newField,
      title: newTitle,
      type: 'text',
      width: 120,
      required: false
    }
    columnsConfig.value.splice(index + 1, 0, newColumn)
  }

  // 生成下拉选项
  const generateSelectOptions = (columnIndex) => {
    const optionsStr = ElMessageBox.prompt('请输入下拉选项，用逗号分隔', '设置下拉选项', {
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    }).then(({ value }) => {
      if (value) {
        const options = value.split(',').map(option => ({
          label: option.trim(),
          value: option.trim()
        }))
        columnsConfig.value[columnIndex].options = options
      }
    })
  }

  // 查找台账节点的递归函数
  const findLedgerNode = (nodes, id) => {
    for (let i = 0; i < nodes.length; i++) {
      if (nodes[i].id === id) {
        return nodes[i]
      }
      if (nodes[i].children && nodes[i].children.length > 0) {
        const found = findLedgerNode(nodes[i].children, id)
        if (found) {
          return found
        }
      }
    }
    return null
  }

  // 确认创建或编辑台账
  const confirmCreateLedger = () => {
    if (!ledgerName.value.trim()) {
      ElMessage({
        message: '请输入台账名称',
        type: 'warning',
        grouping: true,
      })
      return
    }

    // 准备台账配置
    const ledgerConfig = {
      name: ledgerName.value,
      columns: columnsConfig.value,
      tableSettings: {
        stripe: tableSettings.value.stripe,
        showOverflow: tableSettings.value.showOverflow,
        showHeaderOverflow: tableSettings.value.showHeaderOverflow,
        columnResizable: tableSettings.value.columnResizable,
        columnDraggable: tableSettings.value.columnDraggable,
        rowHover: tableSettings.value.rowHover,
        editTrigger: tableSettings.value.editTrigger,
        editMode: tableSettings.value.editMode,
        border: tableSettings.value.border,
        showToolbar: tableSettings.value.showToolbar,
        showFooter: tableSettings.value.showFooter,
        highlightCurrentRow: tableSettings.value.highlightCurrentRow
      }
    }

    if (isEditMode.value && gridOptions.schema.ledgerId) {
      // 有id 编辑模式：更新现有台账
      const existingLedger = findLedgerNode(ledgersTree.value, editingLedgerId.value)
      if (existingLedger) {
        // 更新台账配置
        existingLedger.name = ledgerConfig.name
        existingLedger.columns = ledgerConfig.columns
        existingLedger.tableSettings = ledgerConfig.tableSettings

        // 如果当前查看的就是这个台账，重新加载数据和配置
        if (selectedNode.value && selectedNode.value.id === editingLedgerId.value) {
          // 重新加载台账数据
          loadLedgerData(existingLedger)
        }

        ledgerSetupDialogVisible.value = false
        ElMessage({
          message: '编辑台账成功',
          type: 'success',
          grouping: true,
        })
      } else {
        ElMessage({
          message: '未找到要编辑的台账',
          type: 'error',
          grouping: true,
        })
      }
    } else {
      // 创建模式：新建台账
      const newLedger = {
        id: Date.now().toString(),
        name: ledgerConfig.name,
        isFolder: false,
        type: 'ledger', // 台账类型
        columns: ledgerConfig.columns,
        tableSettings: ledgerConfig.tableSettings
      }

      // 添加到选中的文件夹下
      if (!selectedNode.value.children) {
        selectedNode.value.children = []
      }
      selectedNode.value.children.push(newLedger)
      ledgerSetupDialogVisible.value = false
      ElMessage({
        message: '新建台账成功',
        type: 'success',
        grouping: true,
      })
    }
  }

  // 处理删除选中项
  const handleDeleteSelected = () => {
    if (!selectedNodeRightClick.value) return

    // 检查是否有子节点
    let confirmMessage = `确定要删除「${selectedNodeRightClick.value.name}」吗？`
    if (selectedNodeRightClick.value.children && selectedNodeRightClick.value.children.length > 0) {
      confirmMessage += ' 此操作将同时删除其所有子项。'
    }

    ElMessageBox.confirm(
      confirmMessage,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      // 遍历获取所有子节点ID
      const getAllChildIds = (node) => {
        let ids = []
        if (node.children) {
          node.children.forEach(child => {
            ids.push(child.id)
            if (child.children) {
              ids = ids.concat(getAllChildIds(child))
            }
          })
        }
        return ids
      }
      const childIds = getAllChildIds(selectedNodeRightClick.value)
      // 删除所有子节点
      childIds.forEach(id => {
        treeRef.value.remove(id)
      })
      treeRef.value.remove(selectedNodeRightClick.value.id)
      selectedNode.value = null
      selectedNodeRightClick.value = null
      ElMessage({
        message: '删除成功',
        type: 'success',
        grouping: true,
      })
    }).catch(() => {
      // 取消操作
    })
  }

  // 处理导入台账
  const handleImportLedger = () => {
    ElMessage({
      message: '导入功能待实现',
      type: 'info',
      grouping: true,
    })
  }

  // 处理导出台账
  const handleExportLedger = () => {
    if (!selectedNode.value) return

    ElMessage({
      message: '导出功能待实现',
      type: 'info',
      grouping: true,
    })
  }

  // 处理添加记录
  const handleAddRecord = () => {
    if (!selectedNode.value) return

    const newRecord = {
      id: Date.now()
    }

    // 为每个列添加默认值
    gridOptions.columns.forEach(column => {
      // 跳过序号列和操作列
      if (column.field !== 'seq' && column.field !== 'operation') {
        newRecord[column.field] = ''
      }
    })

    gridOptions.data.push(newRecord)

    // 刷新表格数据
    if (gridRef.value) {
      gridRef.value.refreshColumn()
    }

    ElMessage({
      message: '添加记录成功',
      type: 'success',
      grouping: true,
    })
  }

  // 处理编辑行
  const handleEditRow = (row) => {
    // 编辑功能由vxe-table内置实现
  }

  // 处理删除行
  const handleDeleteRow = (row) => {
    ElMessageBox.confirm(
      '确定要删除这条记录吗？',
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      const index = gridOptions.data.findIndex(item => item.id === row.id)
      if (index > -1) {
        gridOptions.data.splice(index, 1)
      }
    }).catch(() => {
      // 取消操作
    })
  }

  // 处理批量导入
  const handleBatchImport = () => {
    ElMessage({
      message: '批量导入功能待实现',
      type: 'info',
      grouping: true,
    })
  }

  // 处理批量导出
  const handleBatchExport = () => {
    ElMessage({
      message: '批量导出功能待实现',
      type: 'info',
      grouping: true,
    })
  }

  // 切换编辑模式
  const toggleEditMode = () => {
    // 如果从编辑模式退出，提示是否保存
    if (isEdit.value) {
      ElMessageBox.confirm(
        '当前处于编辑模式，退出将丢失未保存的修改，是否继续？',
        '退出编辑确认',
        {
          confirmButtonText: '退出',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(() => {
        isEdit.value = false
        ElMessage({
          message: '已退出编辑模式',
          type: 'success',
          grouping: true
        })
      }).catch(() => {
        // 取消退出，保持编辑模式
      })
    } else {
      // 进入编辑模式
      isEdit.value = true
      ElMessage({
        message: '已进入编辑模式，可进行数据修改',
        type: 'info',
        grouping: true
      })
    }
  }

  // 保存修改
  const handleSaveChanges = () => {
    if (!selectedNode.value) return

    // 调用API保存数据
    httpTest.post(`/saveLedgerData`, {
      ledgerId: selectedNode.value.id,
      data: gridOptions.schema.data
    })
      .then(response => {
        ElMessage({
          message: '保存成功',
          type: 'success',
          grouping: true
        })
        // 保存成功后退出编辑模式
        isEdit.value = false
      })
      .catch(error => {
        console.error('保存台账数据失败:', error)
        ElMessage({
          type: 'error',
          message: '保存台账数据失败'
        })
      })
  }

  // 取消修改
  const handleCancelChanges = () => {
    if (!selectedNode.value) return

    // 重新加载数据，放弃当前修改
    httpTest.get(`/getLedgerData?ledgerId=${selectedNode.value.id}`)
      .then(response => {
        gridOptions.schema.data = response.data.data || []
        // 取消后退出编辑模式
        isEdit.value = false
        ElMessage({
          message: '已取消修改',
          type: 'info',
          grouping: true
        })
      })
      .catch(error => {
        console.error('获取台账数据失败:', error)
        ElMessage({
          type: 'error',
          message: '获取台账数据失败'
        })
      })
  }

  // 点击单元格时保存单元格信息
  function onCellClick(params) {
    currentCell.value = params
  }

  // 树节点拖拽约束
  const handleBeforeDrop = (dragKey, dropKey, hoverPart) => {
    // 获取当前拖拽的节点信息
    const dragNode = treeRef.value.getNode(dragKey);
    // 获取目标节点信息
    const dropNode = treeRef.value.getNode(dropKey);
    switch (dragNode.type) {
      case 'ledger':
        // 台账  前后都可放置
        if (dropNode.type === 'ledger' && (hoverPart === 'before' || hoverPart === 'after')) {
          return true;
        }
        // 分组，任何位置都可放置
        if (dropNode.type === 'group') {
          return true;
        }
        // 项目层级下也可放置
        if (dropNode.type === 'project' && hoverPart === 'body') {
          return true;
        }
        break;
      case 'group':
        // 分组只能放置在项目或其他分组下，且只能放置在目标节点前面
        if (dropNode.type === 'group') {
          return true;
        }
        if (dropNode.type === 'project' && hoverPart === 'body') {
          return true;
        }
        break;
      case 'project':
        // 项目只能放置在其他项目的前面，不能放置在分组或台账下
        if (dropNode.type === 'project' && (hoverPart === 'before' || hoverPart === 'after')) {
          return true;
        }
        break;
      default:
        return false;
    }
    return false;
  }

  // 树节点类名设置
  const setNodeClassName = (node) => {
    return {
      'tree-node-root': node.type === 'root',
      'tree-node-project': node.type === 'project',
      'tree-node-group': node.type === 'group',
      'tree-node-ledger': node.type === 'ledger',
      'tree-node-default': true,
      'isSelected': node.id === selectedNode.value?.id
    }
  }
</script>

<style scoped>
  /* 台账管理页面样式 */
  .ledger-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 16px;
    background-color: #f5f7fa;
  }

  .ledger-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding: 16px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .ledger-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #333;
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }

  .ledger-content {
    display: flex;
    flex: 1;
    gap: 16px;
    overflow: hidden;
  }

  .ledger-directory-wrapper {
    position: relative;
    display: flex;
    height: 100%;
  }

  .ledger-directory {
    width: 300px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
    height: 100%;
    transition: width 0.3s ease;
    overflow: hidden;
  }

  /* 目录折叠样式 - 完全收起 */
  .ledger-directory.collapsed {
    width: 0;
    padding: 0;
    border: none;
    box-shadow: none;
  }

  /* 折叠状态下隐藏所有内容 */
  .ledger-directory.collapsed * {
    display: none;
  }

  /* 展开按钮样式 */
  .expand-button-container {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    box-shadow: none;
    padding: 0;
    border: none;
    height: auto;
    width: auto;
    cursor: pointer;
    transition: all 0.3s ease;
  }

  /* 调整展开按钮的位置，确保在目录收起时显示在正确位置 */
  .ledger-directory.collapsed+.expand-button-container {
    left: 0;
  }

  /* 为按钮本身添加样式 */
  .expand-button-container .el-button {
    background-color: #fff;
    border-radius: 0 4px 4px 0;
    box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1);
    padding: 4px 8px;
    border: none;
    outline: none;
  }

  .expand-button-container .el-button:hover {
    background-color: #f5f7fa;
    color: #409eff;
  }

  /* 右侧数据区域自动填充剩余空间 */
  .ledger-data {
    flex: 1;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .directory-header {
    flex: 0 0 auto;
    padding: 12px;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
  }

  .search-box {
    flex: 1;
    margin-bottom: 0;
  }

  .directory-actions {
    display: flex;
    gap: 4px;
  }

  .directory-tree {
    flex: 1 1 0;
    overflow: auto;
    padding: 8px;
  }

  /* 自定义滚动条 */
  .directory-tree-tree::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  .directory-tree-tree::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 3px;
  }

  .directory-tree::-webkit-scrollbar-thumb {
    background: #c0c4cc;
    border-radius: 4px;
  }

  .directory-tree-tree::-webkit-scrollbar-thumb:hover {
    background: #909399;
  }

  .data-header {
    padding: 12px 16px;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .data-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 500;
    color: #333;
  }

  .data-actions {
    display: flex;
    gap: 8px;
  }

  .account-content {
    flex: 1;
    overflow: hidden;
    padding: 16px;
    display: flex;
    flex-direction: column;
  }

  .account-section {
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  /* 台账设置对话框样式 */
  .ledger-setup-content {
    max-height: 500px;
    overflow-y: auto;
  }

  .setup-section {
    margin-bottom: 20px;
  }

  .setup-section h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 500;
    color: #333;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .columns-config {
    border: 1px solid #e6e6e6;
    border-radius: 4px;
    overflow: hidden;
  }

  .column-item {
    padding: 16px;
    background-color: #fafafa;
  }

  .column-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    font-weight: 500;
    color: #666;
  }

  .column-actions {
    display: flex;
    gap: 8px;
  }

  .column-actions .el-button {
    padding: 2px 8px;
    font-size: 12px;
  }

  .divider {
    height: 1px;
    background-color: #e6e6e6;
  }

  .column-item .el-form {
    margin-bottom: 0;
  }

  .column-item .el-form-item {
    margin-bottom: 12px;
  }

  .column-item .el-form-item:last-child {
    margin-bottom: 0;
  }

  /* 自定义树形节点样式 */
  .custom-tree-node {
    display: inline-block;
  }

  /* 自定义右键菜单样式 */
  .custom-context-menu {
    position: fixed;
    z-index: 2000;
    background-color: #fff;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    padding: 2px 0;
    min-width: 120px;
  }

  .context-menu-item {
    padding: 6px 16px;
    font-size: 14px;
    color: #333;
    cursor: pointer;
    white-space: nowrap;
  }

  .context-menu-item:hover {
    background-color: #f5f7fa;
    color: #1890ff;
  }

  /* vxe-table 样式优化 */
  :deep(.vxe-table) {
    border-radius: 8px;
    overflow: hidden;
  }

  :deep(.vxe-table--header-wrapper) {
    background-color: #fafafa;
  }

  :deep(.vxe-table--header-wrapper th) {
    font-weight: 600;
    color: #333;
    border-bottom: 1px solid #e0e0e0;
  }

  :deep(.vxe-table--header-wrapper th:hover) {
    background-color: #f0f0f0;
  }

  :deep(.vxe-table--body-wrapper) {
    border: 1px solid #e0e0e0;
    border-top: none;
  }

  :deep(.vxe-table--body-wrapper .vxe-row:hover) {
    background-color: #f5f7fa;
  }

  :deep(.vxe-table--body-wrapper .vxe-row.vxe-row--current) {
    background-color: #e6f7ff;
  }

  :deep(.vxe-table--body-wrapper .vxe-cell) {
    border-right: 1px solid #f0f0f0;
  }

  :deep(.vxe-table--body-wrapper .vxe-cell:last-child) {
    border-right: none;
  }

  :deep(.vxe-table--body-wrapper .vxe-row) {
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.vxe-table--body-wrapper .vxe-row:last-child) {
    border-bottom: none;
  }

  :deep(.vxe-button) {
    border-radius: 4px;
    transition: all 0.3s;
  }

  :deep(.vxe-button:hover) {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  :deep(.vxe-toolbar) {
    padding: 10px 0;
    border-bottom: none;
  }

  :deep(.vxe-grid) {
    border: none;
  }

  /* 筛选样式优化 */
  :deep(.vxe-table--filter-trigger) {
    opacity: 0.6;
    transition: opacity 0.2s;
  }

  :deep(.vxe-table--header-wrapper th:hover .vxe-table--filter-trigger) {
    opacity: 1;
  }

  /* 空状态样式 */
  :deep(.vxe-table--empty-wrapper) {
    padding: 60px 0;
    color: #909399;
  }

  :deep(.vxe-table--empty-wrapper .vxe-table--empty-block) {
    font-size: 14px;
  }

  /* 响应式布局 */
  @media (max-width: 1024px) {
    .ledger-directory {
      width: 250px;
    }
  }

  @media (max-width: 768px) {
    .ledger-content {
      flex-direction: column;
    }

    .ledger-directory {
      width: 100%;
      height: 200px;
    }

    .header-actions {
      flex-wrap: wrap;
    }
  }


  /* ---- 树形节点样式 ---- */
  /* ---- 通用节点 ---- */
  :deep(.tree-node-default) {
    padding: 6px 12px;
    border-radius: 6px;
    position: relative;
    transition: all 0.2s ease;
    font-size: 13px;
    color: #333;
    background: transparent;
  }

  /* ---- 悬停状态 ---- */
  :deep(.tree-node-default:hover) {
    background: linear-gradient(90deg,
        rgba(22, 119, 255, 0.08) 0%,
        rgba(22, 119, 255, 0.02) 100%);
  }

  /* ---- 节点类型字体 & 基础颜色 ---- */
  :deep(.tree-node-root) {
    font-weight: 700;
    font-size: 14px;
    color: #606266;
  }

  :deep(.tree-node-project) {
    font-weight: 600;
    font-size: 13.5px;
    color: #409eff;
  }

  :deep(.tree-node-group) {
    font-weight: 500;
    font-size: 13px;
    color: #67c23a;
  }

  :deep(.tree-node-ledger) {
    font-weight: 400;
    font-size: 12.5px;
    color: #e6a23c;
  }

  /* ---- 选中状态 ---- */
  :deep(.isSelected) {
    /* 背景淡蓝 + 阴影 */
    background: linear-gradient(90deg,
        rgba(22, 119, 255, 0.12) 0%,
        rgba(22, 119, 255, 0.05) 100%);
    font-weight: 600;
    box-shadow: 0 2px 8px rgba(22, 119, 255, 0.15);

    /* 左侧高亮条 */
    border-left: 3px solid #1677ff;
  }

  /* ---- 左侧类型标识（未选中）---- */
  :deep(.tree-node-root):not(.isSelected),
  :deep(.tree-node-project):not(.isSelected),
  :deep(.tree-node-group):not(.isSelected),
  :deep(.tree-node-ledger):not(.isSelected) {
    border-left-width: 3px;
    border-left-style: solid;
    border-left-color: transparent;
  }
</style>