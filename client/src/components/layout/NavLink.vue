<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import type { NavItem } from '@/config/navigation'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

const props = defineProps<{
  item: NavItem
  collapsed?: boolean
}>()

const emit = defineEmits<{ action: [item: NavItem] }>()

const route = useRoute()

const isActive = computed(() => {
  if (!props.item.to) return false
  return route.path === props.item.to || route.path.startsWith(`${props.item.to}/`)
})

const itemClasses = computed(() => [
  'group flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium',
  'transition-colors duration-150',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring',
  isActive.value
    ? 'bg-sidebar-accent text-sidebar-accent-foreground'
    : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground',
])
</script>

<template>
  <TooltipProvider v-if="collapsed" :delay-duration="200">
    <Tooltip>
      <TooltipTrigger as-child>
        <RouterLink
          v-if="item.to"
          :to="item.to"
          :class="[...itemClasses, 'justify-center px-0']"
        >
          <component :is="item.icon" class="size-5 shrink-0" />
        </RouterLink>
        <button
          v-else
          type="button"
          :class="[...itemClasses, 'w-full justify-center px-0']"
          @click="emit('action', item)"
        >
          <component :is="item.icon" class="size-5 shrink-0" />
        </button>
      </TooltipTrigger>
      <TooltipContent side="right">{{ item.label }}</TooltipContent>
    </Tooltip>
  </TooltipProvider>

  <template v-else>
    <RouterLink v-if="item.to" :to="item.to" :class="itemClasses">
      <component :is="item.icon" class="size-5 shrink-0" />
      <span class="truncate">{{ item.label }}</span>
    </RouterLink>
    <button v-else type="button" :class="[...itemClasses, 'w-full']" @click="emit('action', item)">
      <component :is="item.icon" class="size-5 shrink-0" />
      <span class="truncate">{{ item.label }}</span>
    </button>
  </template>
</template>
