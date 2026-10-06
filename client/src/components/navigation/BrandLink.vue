<script setup lang="ts">
import { RouterLink } from "vue-router";
import logo from "@/assets/favicon/favicon.svg";
</script>

<template>
  <RouterLink
      to="/"
      class="flex items-center gap-2 rounded-md px-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring group-data-[collapsible=icon]:justify-center"
  >
    <img :src="logo" alt="" class="size-7 shrink-0" />
    <span
        class="font-heading text-lg font-bold tracking-wide group-data-[collapsible=icon]:sr-only"
    >
      ASTROHOP
    </span>
  </RouterLink>
</template>