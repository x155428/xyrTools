import axios from 'axios';
import router from '@/router/index.js';
import { getServerAddress } from '@/js/getServerAddress.js';
import { ElMessage } from 'element-plus';

const serverAddress = getServerAddress();
const http = axios.create({
    baseURL: serverAddress,
    timeout: 10000,
    withCredentials: true,
    headers: {
        'X-Requested-With': 'XMLHttpRequest',
    },
});

// 全局响应拦截器：401、4xx、5xx 统一处理
http.interceptors.response.use(
    response => response,
    async error => {
        const status = error?.response?.status;

        if (status === 401) {
            const { useAuthStore } = await import('@/store/auth.js');
            const authStore = useAuthStore();
            authStore.logout();
            //ElMessage({
            //    message: '会话已过期，请重新登录!',
            //    type: 'error',
            //    grouping: true,
            //});
            router.push('/'); // 重定向到登录页
        } else if (status >= 400 && status < 600) {
            // 提取错误信息
            const msg =
                error.response?.data?.message ||
                `请求失败，状态码：${status}`;

            ElMessage({
                message: msg,
                type: 'error',
                grouping: true,
            });
            //console.error(`[HTTP Error ${status}]`, error.response);
        } else {
            // 其他异常
            const msg = "请检查网络或服务器状态，未知异常：" + error.response?.data?.message || `请求失败，状态码：${status}`;
            ElMessage({
                message: msg,
                type: 'error',
                grouping: true,
            });
        }

        return Promise.reject(error);
    }
);

export default http;
