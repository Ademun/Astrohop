<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, useRoute } from "vue-router";
import type { NavItem } from "@/config/navigation";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const props = defineProps<{
  item: NavItem;
  collapsed?: boolean;
}>();

const emit = defineEmits<{ action: [item: NavItem] }>();

const route = useRoute();

const isActive = computed(() => {
  if (!props.item.to) return false;
  return (
    route.path === props.item.to ||
    route.path.startsWith(`${props.item.to}/`)
  );
});

const classes = computed(() => [
  "group relative flex w-full items-center gap-3 rounded-lg py-2 text-sm font-medium",
  "transition-colors duration-150",
  "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring",
  props.collapsed ? "justify-center px-0" : "px-3",
  isActive.value
    ? "bg-sidebar-accent text-sidebar-accent-foreground"
    : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
]);

const iconClass = computed(() =>
  props.collapsed ? "size-5 shrink-0" : "size-4 shrink-0"
);
</script>

<template>
  <!-- Collapsed: icon-only, tooltip reveal -->
  <Tooltip v-if="collapsed" :delay-duration="200">
    <TooltipTrigger as-child>
      <RouterLink
        v-if="item.to"
        :to="item.to"
        :class="classes"
        :aria-current="isActive ? 'page' : undefined"
      >
        <component :is="item.icon" :class="iconClass" />
        <span class="sr-only">{{ item.label }}</span>
      </RouterLink>
      <button
        v-else
        type="button"
        :class="classes"
        @click="emit('action', item)"
      >
        <component :is="item.icon" :class="iconClass" />
        <span class="sr-only">{{ item.label }}</span>
      </button>
    </TooltipTrigger>
    <TooltipContent side="right" :side-offset="8">
      {{ item.label }}
    </TooltipContent>
  </Tooltip>

  <!-- Expanded -->
  <template v-else>
    <RouterLink
      v-if="item.to"
      :to="item.to"
      :class="classes"
      :aria-current="isActive ? 'page' : undefined"
    >
      <span
        v-if="isActive"
        aria-hidden="true"
        class="pointer-events-none absolute left-0 top-1/2 h-4 w-1 -translate-x-1/2 -translate-y-1/2 rounded-full bg-sidebar-primary"
      />
      <component :is="item.icon" :class="iconClass" />
      <span class="truncate">{{ item.label }}</span>
    </RouterLink>
    <button
      v-else
      type="button"
      :class="classes"
      @click="emit('action', item)"
    >
      <component :is="item.icon" :class="iconClass" />
      <span class="truncate">{{ item.label }}</span>
    </button>
  </template>
</template>