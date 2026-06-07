<template>
  <div id="app">
    <h1>Sentinel 資安監控平台</h1>
    <!-- 頁面路由佔位，後續接入 vue-router -->
    <slot />
  </div>
</template>

<script setup lang="ts">
// 佔位：後續加入 vue-router 與全域狀態管理
</script>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}
body {
  font-family: system-ui, -apple-system, sans-serif;
  background: #0f172a;
  color: #e2e8f0;
}
</style>
