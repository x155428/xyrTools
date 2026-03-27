<!--
 * @Author: 小鱼
 * @Date: 2025-07-03 15:27:13
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-17 16:17:28
 * @FilePath: \passwordManageWeb\src\App.vue
 * @Description: 
 * 
 * Copyright (c) 2025 by 小鱼, All Rights Reserved. 
-->
<template>
  <div id="app" style="margin: 0;padding: 0;width: 100%;height: 100%; max-width: 100%;">
    <router-view></router-view>
  </div>
</template>

<script>
  export default {
    name: 'App',
  };
</script>
<style>
  body {
    margin: 0;
    padding: 0;
    width: 100%;
    min-height: 100vh;
    max-width: 100%;
    place-items: normal;
    background-color: var(--bg-primary);
    background-image:
      radial-gradient(circle at 20% 30%, rgba(37, 99, 235, 0.1) 0%, transparent 40%),
      radial-gradient(circle at 80% 70%, rgba(16, 185, 129, 0.1) 0%, transparent 40%),
      linear-gradient(135deg, rgba(37, 99, 235, 0.05) 0%, rgba(16, 185, 129, 0.05) 100%);
    background-size: cover;
    background-attachment: fixed;
    background-blend-mode: multiply;
  }

  #app {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    width: 100%;
    max-width: 100%;
    margin: 0;
    padding: 20px;
    box-sizing: border-box;
  }
</style>