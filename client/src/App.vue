<script setup lang="ts">
import AppNavigation from "@/components/layout/AppNavigation.vue";
</script>

<template>
  <div class="flex min-h-dvh flex-col md:flex-row">
    <AppNavigation />
    <main class="min-w-0 flex-1">
      <router-view v-slot="{ Component, route }">
        <transition name="page-fade" mode="out-in">
          <component :is="Component" :key="route.path" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style>
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.15s ease-out, transform 0.15s ease-out;
}

.page-fade-leave-to {
  opacity: 0;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}
</style>
