import { defineStore } from 'pinia'
 import { checkSession } from '@/js/checkSession.js';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    username: sessionStorage.getItem('username') || '', 
    // 用户权限列表
    permissions: JSON.parse(sessionStorage.getItem('user_permissions') || '[]'),
    // 用户头像，不再保存到本地
    avatar: ''
  }),
  actions: {
    // 登录成功后设置凭证
    login(username, permissinos_v) {
      this.username = username
      this.permissions = permissinos_v
      
      // 持久化存储
      sessionStorage.setItem('username', username)
      sessionStorage.setItem('user_permissions', JSON.stringify(permissinos_v))
    },
    
    // 退出登录
    logout() {
      sessionStorage.removeItem('username')
      sessionStorage.removeItem('user_permissions')
      this.username = ''
      this.permissions = []
      this.avatar = ''
    },
    
    // 检查权限方法
    hasPermission(permission) {
      return this.permissions.includes(permission)
    },
    
    // 更新用户头像，不再保存到本地
    updateAvatar(avatarData) {
      this.avatar = avatarData
    },
    
    // 检查会话
    async checkSession(){
        let isLogin = await checkSession();
        if(!isLogin){
            this.logout();
        }
    }
  }
})