<script setup lang="ts">
import AppNavigation from '@/components/layout/AppNavigation.vue'
</script>

<template>
  <div class="flex min-h-dvh flex-col md:flex-row">
    <AppNavigation />
    <main class="min-w-0 flex-1">
      <RouterView />
    </main>
  </div>
</template>