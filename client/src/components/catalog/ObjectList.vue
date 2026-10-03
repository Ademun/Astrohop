<script setup lang="ts">
import { computed, ref } from "vue";
import { Sparkles } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import RetryAlert from "@/components/catalog/RetryAlert.vue";
import type { SearchHit } from "@/types/api";

const props = defineProps<{
  items: SearchHit[];
  selectedId: number | null;
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  error: string | null;
  emptyText: string;
  suggestions: string[];
  scoped?: boolean;
}>();

const emit = defineEmits<{
  select: [item: SearchHit];
  loadMore: [];
  retry: [];
  suggest: [term: string];
}>();

const listEl = ref<HTMLElement | null>(null);
const focusIndex = ref<number | null>(null);

const showSkeleton = computed(() => props.loading && props.items.length === 0);
const refreshing = computed(() => props.loading && props.items.length > 0);

const tabbableIndex = computed(() => {
  const selected = props.items.findIndex((item) => item.id === props.selectedId);
  return Math.min(
      focusIndex.value ?? Math.max(selected, 0),
      props.items.length - 1,
  );
});

function title(item: SearchHit): string {
  return item.common_name ?? item.identifier;
}

function subtitle(item: SearchHit): string {
  return [
    item.identifier !== title(item) ? item.identifier : null,
    props.scoped ? null : item.collection,
  ]
      .filter(Boolean)
      .join(" · ");
}

function rowKey(item: SearchHit): string {
  return `${item.id}:${item.identifier}:${item.collection ?? ""}`;
}

function moveFocus(event: KeyboardEvent, index: number) {
  const last = props.items.length - 1;
  const targets: Record<string, number> = {
    ArrowDown: index + 1,
    ArrowUp: index - 1,
    Home: 0,
    End: last,
  };
  const target = targets[event.key];
  if (target === undefined) return;

  event.preventDefault();
  listEl.value
      ?.querySelectorAll("button")
      [Math.min(Math.max(target, 0), last)]?.focus();
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <div
        v-if="showSkeleton"
        role="status"
        aria-label="Loading results"
        class="flex flex-col gap-1"
    >
      <Skeleton
          v-for="n in 6"
          :key="n"
          aria-hidden="true"
          class="h-14 w-full rounded-lg"
      />
    </div>

    <template v-else>
      <ul
          v-if="items.length"
          ref="listEl"
          class="flex flex-col gap-1 transition-opacity"
          :class="{ 'opacity-60': refreshing }"
          :aria-busy="refreshing"
      >
        <li v-for="(item, index) in items" :key="rowKey(item)">
          <button
              type="button"
              :data-class="item.class"
              :tabindex="index === tabbableIndex ? 0 : -1"
              :aria-pressed="item.id === selectedId"
              class="group relative flex min-h-14 w-full items-center justify-between gap-3 rounded-lg px-4 py-2 text-left transition-colors hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring aria-pressed:bg-accent"
              @click="emit('select', item)"
              @focus="focusIndex = index"
              @keydown="moveFocus($event, index)"
          >
            <span
                aria-hidden="true"
                class="absolute inset-y-2 left-0 w-1 rounded-full bg-(--class-color) opacity-0 transition-opacity group-aria-pressed:opacity-100"
            />
            <span class="flex min-w-0 flex-col">
              <span class="truncate text-sm font-medium text-foreground">
                {{ title(item) }}
              </span>
              <span
                  v-if="subtitle(item)"
                  class="truncate text-xs text-muted-foreground"
              >
                {{ subtitle(item) }}
              </span>
            </span>
            <span
                v-if="!scoped"
                class="shrink-0 text-xs uppercase tracking-wide text-muted-foreground"
            >
              {{ item.class }}
            </span>
          </button>
        </li>
      </ul>

      <Empty v-else-if="!error" class="border-0">
        <EmptyHeader>
          <div class="flex items-center justify-center">
            <Sparkles class="size-8 text-muted-foreground" :stroke-width="1.5" aria-hidden="true" />
          </div>
          <EmptyDescription>{{ emptyText }}</EmptyDescription>
        </EmptyHeader>
        <EmptyContent
            v-if="suggestions.length"
            class="flex-row flex-wrap justify-center"
        >
          <Button
              v-for="term in suggestions"
              :key="term"
              variant="outline"
              size="sm"
              @click="emit('suggest', term)"
          >
            {{ term }}
          </Button>
        </EmptyContent>
      </Empty>

      <RetryAlert v-if="error" :message="error" @retry="emit('retry')" />
      <Button
          v-else-if="hasMore && !loading"
          variant="ghost"
          size="sm"
          class="self-center"
          :disabled="loadingMore"
          @click="emit('loadMore')"
      >
        {{ loadingMore ? "Loading…" : "Load more" }}
      </Button>
    </template>
  </div>
</template>