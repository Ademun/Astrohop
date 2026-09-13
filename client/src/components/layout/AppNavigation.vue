<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { ChevronLeft, ChevronRight, Menu, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  primaryNavItems,
  footerNavItems,
  type NavItem,
} from "@/config/navigation";
import NavList from "./NavList.vue";
import LinkDevicesAction from "./LinkDevicesAction.vue";

const collapsed = ref(false);
const mobileOpen = ref(false);
const drawerRef = ref<HTMLElement | null>(null);

function handleAction(item: NavItem) {
  console.info(`[nav] "${item.label}" has no action wired up yet.`);
}

function closeMobile() {
  mobileOpen.value = false;
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && mobileOpen.value) closeMobile();
}

onMounted(() => document.addEventListener("keydown", onKeydown));

onBeforeUnmount(() => {
  document.removeEventListener("keydown", onKeydown);
  document.body.style.overflow = "";
});

watch(mobileOpen, async (open) => {
  document.body.style.overflow = open ? "hidden" : "";
  if (!open) return;
  await nextTick();
  const firstFocusable = drawerRef.value?.querySelector<HTMLElement>(
    'a, button, [tabindex]:not([tabindex="-1"])'
  );
  (firstFocusable ?? drawerRef.value)?.focus();
});
</script>

<template>
  <!-- ═══════════ Desktop sidebar ═══════════ -->
  <aside
    :class="[
      'sticky top-0 hidden h-dvh shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground md:flex',
      'transition-[width] duration-200 ease-out motion-reduce:transition-none',
      collapsed ? 'w-16' : 'w-56',
    ]"
  >
    <!-- Floating collapse toggle: always in the same spot -->
    <button
      type="button"
      :aria-label="collapsed ? 'Expand navigation' : 'Collapse navigation'"
      :aria-expanded="!collapsed"
      class="absolute -right-3 top-5 z-20 hidden size-6 items-center justify-center rounded-full border border-sidebar-border bg-sidebar text-sidebar-foreground/50 shadow-sm transition-colors duration-150 hover:text-sidebar-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring md:flex"
      @click="collapsed = !collapsed"
    >
      <ChevronRight v-if="collapsed" class="size-3.5" />
      <ChevronLeft v-else class="size-3.5" />
    </button>

    <!-- Header -->
    <div class="flex h-16 shrink-0 items-center px-3">
      <RouterLink
        to="/"
        class="flex flex-1 items-center gap-2 rounded-md px-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        :class="collapsed && 'justify-center'"
      >
        <img
          src="/src/assets/favicon/favicon.svg"
          alt=""
          class="size-7 shrink-0"
        />
        <span
          v-if="!collapsed"
          class="font-heading text-lg font-bold tracking-wide"
        >
          ASTROHOP
        </span>
        <span v-else class="sr-only">ASTROHOP</span>
      </RouterLink>
    </div>

    <Separator class="bg-sidebar-border" />

    <!-- Primary nav + actions -->
    <nav
      class="flex flex-1 flex-col gap-1 overflow-y-auto p-2"
      aria-label="Primary"
    >
      <NavList
        :items="primaryNavItems"
        :collapsed="collapsed"
        @action="handleAction"
      />

      <Separator class="my-2 bg-sidebar-border" />

      <LinkDevicesAction :collapsed="collapsed" />
    </nav>

    <!-- Footer -->
    <div v-if="!collapsed" class="flex flex-col gap-1 p-2 pb-3">
      <Separator class="mb-1 bg-sidebar-border" />
      <RouterLink
        v-for="link in footerNavItems"
        :key="link.label"
        :to="link.to"
        class="rounded-md px-3 py-1.5 text-xs text-sidebar-foreground/50 transition-colors duration-150 hover:bg-sidebar-accent/40 hover:text-sidebar-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      >
        {{ link.label }}
      </RouterLink>
    </div>
  </aside>

  <!-- ═══════════ Mobile header ═══════════ -->
  <header
    class="sticky top-0 z-50 flex h-16 items-center border-b border-sidebar-border bg-sidebar px-4 text-sidebar-foreground md:hidden"
  >
    <RouterLink
      to="/"
      class="flex flex-1 items-center gap-2 rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      @click="closeMobile"
    >
      <img
        src="/src/assets/favicon/favicon.svg"
        alt=""
        class="size-7 shrink-0"
      />
      <span class="font-heading text-lg font-bold tracking-wide">
        ASTROHOP
      </span>
    </RouterLink>

    <Button
      variant="ghost"
      size="icon"
      class="shrink-0 rounded-lg text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground"
      :aria-label="mobileOpen ? 'Close navigation' : 'Open navigation'"
      aria-controls="mobile-nav"
      :aria-expanded="mobileOpen"
      @click="mobileOpen = !mobileOpen"
    >
      <Transition name="icon-swap" mode="out-in">
        <X v-if="mobileOpen" key="x" class="size-5" />
        <Menu v-else key="menu" class="size-5" />
      </Transition>
    </Button>
  </header>

  <!-- Backdrop -->
  <Transition name="fade">
    <div
      v-if="mobileOpen"
      aria-hidden="true"
      class="fixed inset-0 top-16 z-30 bg-black/40 backdrop-blur-[2px] md:hidden"
      @click="closeMobile"
    />
  </Transition>

  <!-- ═══════════ Mobile drawer ═══════════ -->
  <Transition name="drawer">
    <div
      v-if="mobileOpen"
      id="mobile-nav"
      ref="drawerRef"
      role="dialog"
      aria-modal="true"
      aria-label="Navigation"
      tabindex="-1"
      class="fixed bottom-0 left-0 top-16 z-40 flex w-72 max-w-[85vw] flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground shadow-xl outline-none md:hidden"
    >
      <nav
        class="flex flex-1 flex-col gap-1 overflow-y-auto p-3"
        aria-label="Primary"
      >
        <NavList
          :items="primaryNavItems"
          @action="
            (i) => {
              handleAction(i);
              closeMobile();
            }
          "
        />

        <Separator class="my-2 bg-sidebar-border" />

        <LinkDevicesAction />
      </nav>

      <div class="flex flex-col gap-1 border-t border-sidebar-border p-3">
        <RouterLink
          v-for="link in footerNavItems"
          :key="link.label"
          :to="link.to"
          class="rounded-md px-3 py-1.5 text-xs text-sidebar-foreground/50 transition-colors duration-150 hover:bg-sidebar-accent/40 hover:text-sidebar-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
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

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.drawer-enter-active,
.drawer-leave-active {
  transition: transform 0.25s cubic-bezier(0.32, 0.72, 0, 1);
}
.drawer-enter-from,
.drawer-leave-to {
  transform: translateX(-100%);
}

@media (prefers-reduced-motion: reduce) {
  .icon-swap-enter-active,
  .icon-swap-leave-active,
  .fade-enter-active,
  .fade-leave-active,
  .drawer-enter-active,
  .drawer-leave-active {
    transition: none;
  }
  .icon-swap-enter-from,
  .icon-swap-leave-to,
  .drawer-enter-from,
  .drawer-leave-to {
    transform: none;
  }
}
</style>