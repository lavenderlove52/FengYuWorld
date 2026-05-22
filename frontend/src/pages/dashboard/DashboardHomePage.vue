<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { NCard, NGrid, NGi } from 'naive-ui'
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">欢迎回来</h1>
    <NGrid cols="1 s:2 m:3" x-gap="16" y-gap="16">
      <NGi>
        <RouterLink to="/dashboard/posts/new">
          <NCard title="写文章" hoverable>创建新的 Markdown 文章</NCard>
        </RouterLink>
      </NGi>
      <NGi>
        <RouterLink to="/dashboard/posts">
          <NCard title="我的文章" hoverable>管理已发布与私密文章</NCard>
        </RouterLink>
      </NGi>
      <NGi>
        <RouterLink to="/dashboard/friends">
          <NCard title="好友" hoverable>申请、接受与管理好友</NCard>
        </RouterLink>
      </NGi>
    </NGrid>
  </div>
</template>
