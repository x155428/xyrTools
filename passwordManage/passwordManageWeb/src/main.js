import '@/assets/styles/main.css'

import { createApp } from 'vue'
import App from '@/App.vue'
import router from '@/router'
import 'element-plus/dist/index.css'  // 导入 Element Plus 样式
import { createPinia } from 'pinia'
// 导入 vxe-table 样式
import 'vxe-table/lib/style.css'
import VXETable from 'vxe-table'
import { initTheme } from '@/utils/theme'
import '@/assets/styles/index.css'

initTheme()

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(VXETable)
app.use(router).mount('#app')

