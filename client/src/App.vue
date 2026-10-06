<script setup lang="ts">
import { useStorage } from "@vueuse/core";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppNavigation from "@/components/navigation/AppNavigation.vue";
import BrandLink from "@/components/navigation/BrandLink.vue";

const open = useStorage("sidebar:open", true);
</script>

<template>
  <SidebarProvider v-model:open="open" style="--sidebar-width: 14rem">
    <AppNavigation />
    <SidebarInset class="min-w-0">
      <header
          class="sticky top-0 z-30 flex h-14 shrink-0 items-center justify-between gap-2 border-b border-sidebar-border bg-sidebar px-4 text-sidebar-foreground md:hidden"
      >
        <BrandLink />
        <SidebarTrigger aria-label="Open navigation" />
      </header>

      <div class="min-w-0 flex-1">
        <router-view v-slot="{ Component, route }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" :key="route.path" />
          </transition>
        </router-view>
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>

<style>
.page-fade-enter-active,
.page-fade-leave-active {
  transition:
      opacity 0.15s ease-out,
      transform 0.15s ease-out;
}

.page-fade-leave-to {
  opacity: 0;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

@media (prefers-reduced-motion: reduce) {
  .page-fade-enter-active,
  .page-fade-leave-active {
    transition: none;
  }

  .page-fade-enter-from {
    transform: none;
  }
}
</style>