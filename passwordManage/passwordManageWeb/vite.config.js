import { fileURLToPath, URL } from 'node:url'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'fs';
import path from 'path';
import { lazyImport, VxeResolver } from 'vite-plugin-lazy-import'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
    lazyImport({
      resolvers: [
        VxeResolver({
          libraryName: 'vxe-pc-ui'
        }),
        VxeResolver({
          libraryName: 'vxe-table'
        })
      ]
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    https: {
      key: fs.readFileSync(path.resolve(__dirname, 'C:/Users/小鱼/Desktop/ssl证书链/server.key')), // 私钥文件路径
      cert: fs.readFileSync(path.resolve(__dirname, 'C:/Users/小鱼/Desktop/ssl证书链/server.crt')), // 证书文件路径
    },
    host: '0.0.0.0', // 配置为 0.0.0.0 或特定的域名或IP，默认是 localhost
    port: 8889, // 设置端口
    // 如果需要配置代理
    proxy: {
      '/api': 'http://localhost:4000', // 代理请求
    },
  },
  build: {
    target: 'es2017',
    sourcemap: false,
    minify: 'terser',
    chunkSizeWarningLimit: 1500,
    terserOptions: {
      format: {
        comments: false
      },
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    }
  }
})






