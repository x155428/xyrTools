import { reactive } from 'vue'

export function createLedgerGridState() {
     return reactive({
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
               loading: false, // 是否显示加载中
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
                    isEvery: true, // 每个列都应用过滤
                    remote: false, // 是否远程过滤
                    filterMethod: function (row, column, filterValue) {
                         return row[column.property] === filterValue
                    },
                    showIcon: true, // 是否显示列头过滤图标
               },
               floatingFilterConfig: {  // 浮动过滤配置
                    enabled: true, // 是否启用浮动过滤
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
                              { code: 'insertBefore', name: '在前面插入' },
                              { code: 'insertAfter', name: '在后面插入' },
                              { code: 'remove', name: '删除' }
                         ]
                    },
                    footer: { disabled: true, options: [] }
               },
          },

          /** ========= 2. 列定义 ========= */
          columns: [],

          /** ========= 3. 数据 ========= */
          data: [] // 深度响应式，单元格修改自动触发更新
     })
}