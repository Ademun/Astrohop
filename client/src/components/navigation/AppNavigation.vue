<script setup lang="ts">
import { computed, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { ChevronLeft, ChevronRight, Plus } from "@lucide/vue";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
  useSidebar,
} from "@/components/ui/sidebar";
import { footerNavItems, primaryNavItems, type NavItem } from "@/config/navigation";
import BrandLink from "./BrandLink.vue";
import NavMenuItem from "./NavMenuItem.vue";

const emit = defineEmits<{ action: [item: NavItem] }>();

const route = useRoute();
const { state, isMobile, toggleSidebar, setOpenMobile } = useSidebar();

const isExpanded = computed(() => state.value === "expanded");
const isCollapsed = computed(() => state.value === "collapsed" && !isMobile.value);

watch(
    () => route.fullPath,
    () => setOpenMobile(false),
);
</script>

<template>
  <Sidebar collapsible="icon">
    <button
        v-if="!isMobile"
        type="button"
        aria-label="Toggle navigation"
        :aria-expanded="isExpanded"
        class="absolute -right-3 top-5 z-20 flex size-6 items-center justify-center rounded-full border border-sidebar-border bg-sidebar text-sidebar-foreground/60 shadow-sm transition-colors hover:text-sidebar-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        @click="toggleSidebar"
    >
      <ChevronLeft v-if="isExpanded" class="size-3.5" aria-hidden="true" />
      <ChevronRight v-else class="size-3.5" aria-hidden="true" />
    </button>

    <SidebarHeader class="h-16 justify-center border-b border-sidebar-border">
      <BrandLink />
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
                as-child
                tooltip="New mission"
                class="bg-sidebar-primary text-sidebar-primary-foreground hover:bg-sidebar-primary/90 hover:text-sidebar-primary-foreground active:bg-sidebar-primary/90 active:text-sidebar-primary-foreground"
            >
              <RouterLink to="/missions/new">
                <Plus aria-hidden="true" />
                <span>New mission</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>

      <SidebarGroup>
        <nav aria-label="Primary">
          <SidebarMenu>
            <NavMenuItem
                v-for="item in primaryNavItems"
                :key="item.label"
                :item="item"
                @action="emit('action', $event)"
            />
          </SidebarMenu>
        </nav>
      </SidebarGroup>

      <SidebarSeparator />
    </SidebarContent>

    <SidebarFooter class="border-t border-sidebar-border">
      <nav aria-label="Secondary">
        <SidebarMenu>
          <NavMenuItem
              v-for="link in footerNavItems"
              :key="link.label"
              :item="link"
              size="sm"
              @action="emit('action', $event)"
          />
        </SidebarMenu>
      </nav>
    </SidebarFooter>
  </Sidebar>
</template>