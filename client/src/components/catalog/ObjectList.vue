<script setup lang="ts">
import {computed, ref} from "vue";
import {Sparkles} from "@lucide/vue";
import {Button} from "@/components/ui/button";
import {Skeleton} from "@/components/ui/skeleton";
import type {CollectionMember, SearchHit} from "@/types/api";

type Item = CollectionMember | SearchHit;

const props = defineProps<{
  items: Item[];
  selectedId: number | null;
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  error: string | null;
  emptyText: string;
  suggestions?: string[];
}>();

const emit = defineEmits<{
  select: [item: Item];
  loadMore: [];
  retry: [];
  suggest: [term: string];
}>();

const listEl = ref<HTMLElement | null>(null);
const focusIndex = ref<number | null>(null);

const showSkeleton = computed(() => props.loading && props.items.length === 0);
const refreshing = computed(() => props.loading && props.items.length > 0);

const tabbableIndex = computed(() => {
  const last = props.items.length - 1;
  if (last < 0) return -1;
  const selected = props.items.findIndex((i) => i.id === props.selectedId);
  return Math.min(focusIndex.value ?? (selected >= 0 ? selected : 0), last);
});

function collectionOf(item: Item): string | null {
  return "collection" in item ? item.collection : null;
}

function title(item: Item): string {
  return item.common_name ?? item.identifier;
}

function subtitle(item: Item): string {
  return [item.identifier !== title(item) ? item.identifier : null, collectionOf(item)]
      .filter(Boolean)
      .join(" · ");
}

function rowKey(item: Item): string {
  return `${item.id}:${item.identifier}:${collectionOf(item) ?? ""}`;
}

function isSelected(item: Item): boolean {
  return item.id === props.selectedId;
}

function moveFocus(event: KeyboardEvent, index: number) {
  const last = props.items.length - 1;
  let next: number;
  switch (event.key) {
    case "ArrowDown":
      next = Math.min(index + 1, last);
      break;
    case "ArrowUp":
      next = Math.max(index - 1, 0);
      break;
    case "Home":
      next = 0;
      break;
    case "End":
      next = last;
      break;
    default:
      return;
  }
  event.preventDefault();
  listEl.value
      ?.querySelectorAll<HTMLButtonElement>("button[data-row]")[next]
      ?.focus();
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <div v-if="showSkeleton" class="flex flex-col gap-1" aria-hidden="true">
      <Skeleton v-for="n in 6" :key="n" class="h-14 w-full rounded-lg"/>
    </div>

    <template v-else>
      <ul
          v-if="items.length > 0"
          ref="listEl"
          class="flex flex-col gap-1 transition-opacity"
          :class="{'opacity-60': refreshing}"
          :aria-busy="refreshing"
      >
        <li v-for="(item, index) in items" :key="rowKey(item)">
          <button
              type="button"
              data-row
              :data-class="item.class"
              :tabindex="index === tabbableIndex ? 0 : -1"
              class="relative flex min-h-14 w-full items-center justify-between gap-3 rounded-lg px-4 py-2 text-left transition-colors hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :class="{'bg-accent': isSelected(item)}"
              :aria-pressed="isSelected(item)"
              @click="emit('select', item)"
              @focus="focusIndex = index"
              @keydown="moveFocus($event, index)"
          >
            <span
                aria-hidden="true"
                class="absolute inset-y-2 left-0 w-1 rounded-full bg-(--class-color) transition-opacity"
                :class="isSelected(item) ? 'opacity-100' : 'opacity-0'"
            />
            <span class="flex min-w-0 items-center gap-3">
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
            </span>
            <!-- Текстовый класс нужен только там, где выдача смешанная (поиск) -->
            <span
                v-if="collectionOf(item) !== null"
                class="shrink-0 text-xs uppercase tracking-wide text-muted-foreground"
            >
              {{ item.class }}
            </span>
          </button>
        </li>
      </ul>

      <div
          v-else-if="!error"
          class="flex flex-col items-center gap-4 py-12 text-center"
      >
        <Sparkles
            class="size-8 text-muted-foreground"
            :stroke-width="1.5"
            aria-hidden="true"
        />
        <p class="max-w-xs text-sm text-muted-foreground">{{ emptyText }}</p>
        <div
            v-if="suggestions?.length"
            class="flex flex-wrap justify-center gap-2"
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
        </div>
      </div>

      <div
          v-if="error"
          class="flex flex-col items-start gap-3 rounded-lg border border-border p-4"
      >
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" @click="emit('retry')">
          Try again
        </Button>
      </div>
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