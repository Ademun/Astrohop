<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";
import { Menu, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  primaryNavItems,
  footerNavItems,
  type NavItem,
} from "@/config/navigation";
import NavLink from "./NavLink.vue";

const collapsed = ref(false);
const mobileOpen = ref(false);

function toggleCollapsed() {
  collapsed.value = !collapsed.value;
}

function toggleMobile() {
  mobileOpen.value = !mobileOpen.value;
}

function closeMobile() {
  mobileOpen.value = false;
}

function handleAction(item: NavItem) {
  console.info(`[nav] "${item.label}" has no action wired up yet.`);
}
</script>

<template>
  <aside
    :class="[
      'sticky top-0 hidden h-dvh shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground md:flex',
      'transition-[width] duration-300 ease-in-out motion-reduce:transition-none',
      collapsed ? 'w-18' : 'w-52',
    ]"
  >
    <div
      class="flex h-16 shrink-0 items-center px-3"
      :class="collapsed ? 'justify-center' : 'justify-between'"
    >
      <template v-if="!collapsed">
        <RouterLink
          to="/"
          class="flex flex-1 justify-center rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        >
          <span
            class="font-heading text-xl font-bold tracking-wide text-sidebar-foreground"
          >
            ASTROHOP
          </span>
        </RouterLink>
        <Button
          variant="ghost"
          size="icon"
          class="shrink-0 rounded-lg text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground"
          aria-label="Collapse navigation"
          aria-expanded="true"
          @click="toggleCollapsed"
        >
          <X class="size-4" />
        </Button>
      </template>

      <RouterLink
        v-else
        to="/"
        class="rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      >
        <img
          src="/src/assets/favicon/favicon.svg"
          alt="ASTROHOP"
          class="size-8"
        />
      </RouterLink>
    </div>

    <Separator class="bg-sidebar-border" />

    <div v-if="collapsed" class="flex justify-center py-2">
      <Button
        variant="ghost"
        size="icon"
        class="rounded-lg text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground"
        aria-label="Expand navigation"
        aria-expanded="false"
        @click="toggleCollapsed"
      >
        <Menu class="size-5" />
      </Button>
    </div>

    <nav
      class="flex flex-1 flex-col gap-2 overflow-y-auto p-2"
      aria-label="Primary"
    >
      <NavLink
        v-for="item in primaryNavItems"
        :key="item.label"
        :item="item"
        :collapsed="collapsed"
        @action="handleAction"
      />
    </nav>

    <div v-if="!collapsed" class="flex flex-col gap-1 p-2 pb-4">
      <Separator class="mb-2 bg-sidebar-border" />
      <RouterLink
        v-for="link in footerNavItems"
        :key="link.label"
        :to="link.to"
        class="rounded-md px-3 py-1.5 text-sm text-sidebar-foreground/60 transition-colors duration-150 hover:text-sidebar-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      >
        {{ link.label }}
      </RouterLink>
    </div>
  </aside>

  <header
    class="sticky top-0 z-40 flex h-16 items-center border-b border-sidebar-border bg-sidebar px-4 text-sidebar-foreground md:hidden"
  >
    <RouterLink
      to="/"
      class="flex flex-1 justify-center rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      @click="closeMobile"
    >
      <span
        class="font-heading text-2xl font-bold tracking-wide text-sidebar-foreground"
      >
        ASTROHOP
      </span>
    </RouterLink>

    <Button
      variant="ghost"
      size="icon"
      class="shrink-0 rounded-lg text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground"
      :aria-label="mobileOpen ? 'Close navigation' : 'Open navigation'"
      aria-controls="mobile-nav-overlay"
      :aria-expanded="mobileOpen"
      @click="toggleMobile"
    >
      <Transition name="icon-swap" mode="out-in">
        <X v-if="mobileOpen" key="x" class="size-6" />
        <Menu v-else key="menu" class="size-6" />
      </Transition>
    </Button>
  </header>

  <Transition name="overlay">
    <div
      v-if="mobileOpen"
      id="mobile-nav-overlay"
      class="fixed inset-0 top-16 z-30 flex flex-col overflow-y-auto bg-sidebar text-sidebar-foreground md:hidden"
    >
      <nav class="flex flex-1 flex-col gap-2 p-4" aria-label="Primary">
        <NavLink
          v-for="item in primaryNavItems"
          :key="item.label"
          :item="item"
          @action="
            (i) => {
              handleAction(i);
              closeMobile();
            }
          "
        />
      </nav>

      <div class="flex flex-col gap-1 p-4 pt-0">
        <Separator class="mb-2 bg-sidebar-border" />
        <RouterLink
          v-for="link in footerNavItems"
          :key="link.label"
          :to="link.to"
          class="rounded-md px-3 py-1.5 text-sm text-sidebar-foreground/60 transition-colors duration-150 hover:text-sidebar-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
          @click="closeMobile"
        >
          {{ link.label }}
        </RouterLink>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.icon-swap-enter-active,
.icon-swap-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.icon-swap-enter-from,
.icon-swap-leave-to {
  opacity: 0;
  transform: rotate(-45deg);
}

.overlay-enter-active,
.overlay-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}
.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@media (prefers-reduced-motion: reduce) {
  .icon-swap-enter-active,
  .icon-swap-leave-active,
  .overlay-enter-active,
  .overlay-leave-active {
    transition: none;
  }
  .icon-swap-enter-from,
  .icon-swap-leave-to,
  .overlay-enter-from,
  .overlay-leave-to {
    transform: none;
  }
}
</style>
