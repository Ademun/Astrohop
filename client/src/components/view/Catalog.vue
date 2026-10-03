<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import { useRoute, useRouter, type LocationQueryValue } from "vue-router";
import { ArrowUp, Search, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ApiError, apiClient } from "@/api/client";
import ObjectList from "@/components/catalog/ObjectList.vue";
import ObjectDetails from "@/components/catalog/ObjectDetails.vue";
import type { CatalogObject, Collection, SearchHit } from "@/types/api";

type QueryValue = LocationQueryValue | LocationQueryValue[] | undefined;

const PAGE_SIZE = 50;
const SEARCH_DEBOUNCE_MS = 250;
const DESKTOP_MEDIA_QUERY = "(min-width: 1024px)";
const DEFAULT_PLACEHOLDER = "Search by name, e.g. Andromeda or M31";
const SUGGESTIONS = ["Andromeda", "Crab", "Sirius"];

function firstValue(value: QueryValue): string {
  return (Array.isArray(value) ? value[0] : value) ?? "";
}

function positiveInt(value: QueryValue): number | null {
  const parsed = Number(firstValue(value));
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

function errorMessage(error: unknown): string {
  return error instanceof ApiError
      ? error.message
      : "Something went wrong. Please try again.";
}

const route = useRoute();
const router = useRouter();

const query = ref(firstValue(route.query.q));
const activeCollectionId = ref(positiveInt(route.query.collection));
const selectedId = ref(positiveInt(route.query.object));
const selectedLabel = ref<string | null>(null);

const collections = ref<Collection[]>([]);
const collectionsError = ref<string | null>(null);

const items = ref<SearchHit[]>([]);
const hasMore = ref(false);
const listLoading = ref(false);
const loadingMore = ref(false);
const listError = ref<string | null>(null);

const object = ref<CatalogObject | null>(null);
const objectLoading = ref(false);
const objectError = ref<string | null>(null);

const topSentinel = ref<HTMLElement | null>(null);
const showScrollTop = ref(false);

const desktopQuery = window.matchMedia(DESKTOP_MEDIA_QUERY);
const isDesktop = ref(desktopQuery.matches);
const sheetEl = ref<HTMLDialogElement | null>(null);
const sheetOpen = ref(false);

let topObserver: IntersectionObserver | null = null;
let debounceTimer: ReturnType<typeof setTimeout> | undefined;
let listController: AbortController | null = null;
let objectController: AbortController | null = null;

const hasQuery = computed(() => query.value.trim() !== "");
const hasCollection = computed(() => activeCollectionId.value !== null);
const hasFilters = computed(() => hasQuery.value || hasCollection.value);

const activeCollection = computed(
    () =>
        collections.value.find((c) => c.id === activeCollectionId.value) ?? null,
);

const placeholder = computed(() =>
    activeCollection.value
        ? `Search in ${activeCollection.value.name}`
        : DEFAULT_PLACEHOLDER,
);

const summary = computed(() => {
  if (listLoading.value || items.value.length === 0) return null;
  const count = `${items.value.length.toLocaleString("en-US")}${hasMore.value ? "+" : ""}`;
  return `${count} ${items.value.length === 1 && !hasMore.value ? "result" : "results"}`;
});

const emptyText = computed(() => {
  const term = query.value.trim();
  const scope = activeCollection.value
      ? `“${activeCollection.value.name}”`
      : "this collection";
  if (term && hasCollection.value) {
    return `No objects in ${scope} match “${term}”.`;
  }
  if (term) return `No objects match “${term}”.`;
  if (hasCollection.value) return "This collection is empty.";
  return "Search by name or pick a collection to start browsing.";
});

const canWidenSearch = computed(
    () =>
        hasQuery.value &&
        hasCollection.value &&
        !listLoading.value &&
        !listError.value &&
        items.value.length === 0,
);

function syncRoute() {
  const next: Record<string, string> = {};
  const term = query.value.trim();
  if (term) next.q = term;
  if (activeCollectionId.value !== null) {
    next.collection = String(activeCollectionId.value);
  }
  if (selectedId.value !== null) next.object = String(selectedId.value);
  void router.replace({ query: next });
}

async function loadCollections() {
  collectionsError.value = null;
  try {
    collections.value = await apiClient.listCollections();
  } catch (error) {
    collectionsError.value = errorMessage(error);
  }
}

async function loadList(append = false) {
  listController?.abort();
  const controller = new AbortController();
  listController = controller;

  const term = query.value.trim();
  const collection = activeCollectionId.value;

  listError.value = null;
  loadingMore.value = append;
  listLoading.value = !append;

  try {
    if (term || collection !== null) {
      const hits = await apiClient.searchCatalog(
          {
            query: term || undefined,
            collection: collection ?? undefined,
            limit: PAGE_SIZE + 1,
            offset: append ? items.value.length : 0,
          },
          controller.signal,
      );
      const page = hits.slice(0, PAGE_SIZE);
      items.value = append ? [...items.value, ...page] : page;
      hasMore.value = hits.length > PAGE_SIZE;
    } else {
      items.value = [];
      hasMore.value = false;
    }
  } catch (error) {
    if (!isAbort(error)) {
      listError.value = errorMessage(error);
      if (!append) {
        items.value = [];
        hasMore.value = false;
      }
    }
  } finally {
    if (listController === controller) {
      listLoading.value = false;
      loadingMore.value = false;
    }
  }
}

async function loadObject() {
  objectController?.abort();
  const id = selectedId.value;
  if (id === null) return;

  const controller = new AbortController();
  objectController = controller;
  objectError.value = null;
  objectLoading.value = true;
  object.value = null;

  try {
    object.value = await apiClient.getCatalogObject(id, controller.signal);
  } catch (error) {
    if (!isAbort(error)) objectError.value = errorMessage(error);
  } finally {
    if (objectController === controller) objectLoading.value = false;
  }
}

function applyListChange() {
  syncRoute();
  void loadList();
}

function onQueryInput(value: string | number) {
  query.value = String(value);
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(
      applyListChange,
      query.value.trim() ? SEARCH_DEBOUNCE_MS : 0,
  );
}

function applySuggestion(term: string) {
  clearTimeout(debounceTimer);
  query.value = term;
  applyListChange();
}

function clearQuery() {
  clearTimeout(debounceTimer);
  query.value = "";
  applyListChange();
}

function selectCollection(id: number) {
  clearTimeout(debounceTimer);
  activeCollectionId.value = activeCollectionId.value === id ? null : id;
  applyListChange();
}

function searchAllCollections() {
  clearTimeout(debounceTimer);
  activeCollectionId.value = null;
  applyListChange();
}

function clearSelection() {
  objectController?.abort();
  selectedId.value = null;
  selectedLabel.value = null;
  object.value = null;
  objectError.value = null;
  objectLoading.value = false;
  syncRoute();
}

async function openSheet() {
  sheetOpen.value = true;
  await nextTick();
  if (sheetEl.value && !sheetEl.value.open) sheetEl.value.showModal();
}

function closeSheet() {
  sheetEl.value?.close();
}

function onSheetClose() {
  sheetOpen.value = false;
  clearSelection();
}

function selectItem(item: SearchHit) {
  selectedId.value = item.id;
  selectedLabel.value = item.common_name ?? item.identifier;
  syncRoute();
  void loadObject();
  if (!isDesktop.value) void openSheet();
}

function scrollToTop() {
  const reduceMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
  ).matches;
  window.scrollTo({ top: 0, behavior: reduceMotion ? "auto" : "smooth" });
}

function onDesktopChange(event: MediaQueryListEvent) {
  isDesktop.value = event.matches;
}

watch(isDesktop, (desktop) => {
  if (desktop) sheetOpen.value = false;
});

onMounted(() => {
  desktopQuery.addEventListener("change", onDesktopChange);
  void loadCollections();
  if (hasFilters.value) void loadList();
  if (selectedId.value !== null) {
    void loadObject();
    if (!isDesktop.value) void openSheet();
  }

  if (topSentinel.value) {
    topObserver = new IntersectionObserver(([entry]) => {
      showScrollTop.value = !entry.isIntersecting;
    });
    topObserver.observe(topSentinel.value);
  }
});

onBeforeUnmount(() => {
  desktopQuery.removeEventListener("change", onDesktopChange);
  topObserver?.disconnect();
  clearTimeout(debounceTimer);
  listController?.abort();
  objectController?.abort();
});
</script>

<template>
  <div class="relative mx-auto max-w-6xl px-6 py-8">
    <div
        ref="topSentinel"
        aria-hidden="true"
        class="pointer-events-none absolute inset-x-0 top-0 h-96"
    />

    <div class="flex flex-col gap-4 rounded-xl border border-border bg-card p-4">
      <div class="flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1">
        <h1 class="font-heading text-xl tracking-tight text-foreground">
          Object catalog
        </h1>
        <p class="text-sm text-muted-foreground">
          Stars and deep-sky objects: search by name, optionally within a
          collection
        </p>
      </div>

      <div class="relative">
        <Search
            class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
            aria-hidden="true"
        />
        <Input
            :model-value="query"
            type="text"
            :placeholder="placeholder"
            aria-label="Search objects"
            class="pl-9 pr-9"
            @update:model-value="onQueryInput"
        />
        <Button
            v-if="query"
            variant="ghost"
            size="icon"
            class="absolute right-1 top-1/2 size-7 -translate-y-1/2"
            aria-label="Clear search"
            @click="clearQuery"
        >
          <X class="size-4" aria-hidden="true" />
        </Button>
      </div>

      <div class="flex flex-col gap-2">
        <div
            class="-mx-1 flex gap-2 overflow-x-auto px-1 pb-1 sm:flex-wrap sm:overflow-visible sm:pb-0"
        >
          <Button
              v-for="collection in collections"
              :key="collection.id"
              size="sm"
              class="shrink-0"
              :variant="collection.id === activeCollectionId ? 'default' : 'outline'"
              :aria-pressed="collection.id === activeCollectionId"
              @click="selectCollection(collection.id)"
          >
            {{ collection.name }}
          </Button>
          <p v-if="collectionsError" class="text-sm text-destructive">
            {{ collectionsError }}
            <button
                type="button"
                class="underline underline-offset-2"
                @click="loadCollections"
            >
              Retry
            </button>
          </p>
        </div>
        <p
            v-if="activeCollection?.description"
            class="text-sm text-muted-foreground"
        >
          {{ activeCollection.description }}
        </p>
      </div>
    </div>

    <div
        class="mt-6 grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]"
    >
      <section aria-label="Results" class="flex flex-col gap-3 pb-20">
        <p v-if="summary" class="text-sm text-muted-foreground">
          {{ summary }}
        </p>
        <ObjectList
            :items="items"
            :selected-id="selectedId"
            :loading="listLoading"
            :loading-more="loadingMore"
            :has-more="hasMore"
            :error="listError"
            :empty-text="emptyText"
            :scoped="hasCollection"
            :suggestions="hasFilters ? [] : SUGGESTIONS"
            @select="selectItem"
            @load-more="loadList(true)"
            @retry="loadList(items.length > 0)"
            @suggest="applySuggestion"
        />
        <Button
            v-if="canWidenSearch"
            variant="outline"
            size="sm"
            class="self-center"
            @click="searchAllCollections"
        >
          Search all collections
        </Button>
      </section>

      <section
          v-if="isDesktop"
          aria-label="Object details"
          class="sticky top-6 self-start pl-2 pt-2"
      >
        <ObjectDetails
            :object="object"
            :loading="objectLoading"
            :error="objectError"
            :fallback-title="selectedLabel"
            @retry="loadObject"
        />
      </section>
    </div>

    <dialog
        v-if="!isDesktop"
        ref="sheetEl"
        aria-label="Object details"
        class="object-sheet inset-x-0 bottom-0 top-auto m-0 w-full max-w-none overflow-hidden rounded-t-2xl border-0 bg-background p-0 text-foreground backdrop:bg-black/70"
        @close="onSheetClose"
        @click.self="closeSheet"
    >
      <div v-if="sheetOpen" class="flex max-h-[90dvh] flex-col">
        <div class="flex justify-end px-3 pt-3">
          <Button
              variant="ghost"
              size="icon"
              class="size-8"
              aria-label="Close"
              @click="closeSheet"
          >
            <X class="size-4" aria-hidden="true" />
          </Button>
        </div>
        <div class="overflow-y-auto overscroll-contain px-4 pb-6 pt-3">
          <ObjectDetails
              :object="object"
              :loading="objectLoading"
              :error="objectError"
              :fallback-title="selectedLabel"
              @retry="loadObject"
          />
        </div>
      </div>
    </dialog>

    <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-y-2 opacity-0"
        leave-active-class="transition duration-150 ease-in"
        leave-to-class="translate-y-2 opacity-0"
    >
      <Button
          v-if="showScrollTop"
          variant="outline"
          size="icon"
          class="fixed inset-x-0 bottom-6 z-40 mx-auto size-11 rounded-full bg-card shadow-lg hover:bg-accent"
          aria-label="Back to top"
          @click="scrollToTop"
      >
        <ArrowUp class="size-5" aria-hidden="true" />
      </Button>
    </Transition>
  </div>
</template>

<style>
body:has(dialog.object-sheet[open]) {
  overflow: hidden;
}

@keyframes object-sheet-in {
  from {
    transform: translateY(100%);
  }
}

dialog.object-sheet[open] {
  animation: object-sheet-in 0.22s ease-out;
}

@media (prefers-reduced-motion: reduce) {
  dialog.object-sheet[open] {
    animation: none;
  }
}
</style>