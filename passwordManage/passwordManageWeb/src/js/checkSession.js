import { useAuthStore } from '@/store/auth';
import http from '@/js/http.js';
// 检查会话是否有效的函数
export async function checkSession() {
  const authStore = useAuthStore();
  // 发送带 cookie 的请求到检查会话接口
  try {
    const response = await http.get("/checkSession", {
      withCredentials: true
    });

    const data = response.data.data;
    
    // 解析返回的json数据
    if (data.isLogin) {
      authStore.login(data.username,["user.isLogin"]);
      // 返回会话有效
      return true;
    } else {
      authStore.logout();
      // 返回会话无效
      return false;
    }
  } catch (error) {
    if (error.response) {
      authStore.logout();
      console.error('会话验证请求失败:', error.response.status);
      return false;
    }
    authStore.logout();
    return false;
  }
}
