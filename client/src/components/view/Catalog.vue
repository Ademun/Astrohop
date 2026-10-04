<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRoute, useRouter, type LocationQueryValue } from "vue-router";
import { useMediaQuery, useWindowScroll } from "@vueuse/core";
import { ArrowUp, Search, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui/sheet";
import { ApiError, apiClient } from "@/api/client";
import ObjectList from "@/components/catalog/ObjectList.vue";
import ObjectDetails from "@/components/catalog/ObjectDetails.vue";
import RetryAlert from "@/components/catalog/RetryAlert.vue";
import type { CatalogObject, Collection, SearchHit } from "@/types/api";

type QueryValue = LocationQueryValue | LocationQueryValue[] | undefined;

interface Filters {
  query?: string;
  collection?: number | null;
}

const PAGE_SIZE = 50;
const SEARCH_DEBOUNCE_MS = 250;
const SCROLL_TOP_THRESHOLD_PX = 384;
const SUGGESTIONS = ["Andromeda", "Crab", "Sirius"];

function firstValue(value: QueryValue): string {
  return (Array.isArray(value) ? value[0] : value) ?? "";
}

function positiveInt(value: QueryValue): number | null {
  const parsed = Number(firstValue(value));
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

function errorMessage(error: unknown): string {
  return error instanceof ApiError
      ? error.message
      : "Something went wrong. Please try again.";
}

const route = useRoute();
const router = useRouter();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const prefersReducedMotion = useMediaQuery("(prefers-reduced-motion: reduce)");
const { y: scrollY } = useWindowScroll();

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

let debounceTimer: ReturnType<typeof setTimeout> | undefined;
let listController: AbortController | null = null;
let objectController: AbortController | null = null;

const term = computed(() => query.value.trim());
const hasCollection = computed(() => activeCollectionId.value !== null);
const hasFilters = computed(() => term.value !== "" || hasCollection.value);
const showScrollTop = computed(() => scrollY.value > SCROLL_TOP_THRESHOLD_PX);

const activeCollection = computed(
    () => collections.value.find((c) => c.id === activeCollectionId.value) ?? null,
);

const placeholder = computed(() =>
    activeCollection.value
        ? `Search in ${activeCollection.value.name}`
        : "Search by name, e.g. Andromeda or M31",
);

const summary = computed(() => {
  const count = items.value.length;
  if (listLoading.value || count === 0) return "";
  const total = `${count.toLocaleString("en-US")}${hasMore.value ? "+" : ""}`;
  return `${total} ${count === 1 ? "result" : "results"}`;
});

const emptyText = computed(() => {
  const name = activeCollection.value?.name;
  const scope = name ? `“${name}”` : "this collection";
  if (term.value && hasCollection.value) {
    return `No objects in ${scope} match “${term.value}”.`;
  }
  if (term.value) return `No objects match “${term.value}”.`;
  if (hasCollection.value) return "This collection is empty.";
  return "Search by name or pick a collection to start browsing.";
});

const canWidenSearch = computed(
    () =>
        term.value !== "" &&
        hasCollection.value &&
        !listLoading.value &&
        !listError.value &&
        items.value.length === 0,
);

const detailsProps = computed(() => ({
  object: object.value,
  loading: objectLoading.value,
  error: objectError.value,
  fallbackTitle: selectedLabel.value,
}));

const sheetOpen = computed({
  get: () => !isDesktop.value && selectedId.value !== null,
  set: (open) => {
    if (!open) clearSelection();
  },
});

function syncRoute() {
  void router.replace({
    query: {
      q: term.value || undefined,
      collection: activeCollectionId.value?.toString(),
      object: selectedId.value?.toString(),
    },
  });
}

function resetItems() {
  items.value = [];
  hasMore.value = false;
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

  listError.value = null;
  loadingMore.value = append;
  listLoading.value = !append;

  try {
    if (!hasFilters.value) {
      resetItems();
      return;
    }

    const hits = await apiClient.searchCatalog(
        {
          query: term.value || undefined,
          collection: activeCollectionId.value ?? undefined,
          limit: PAGE_SIZE + 1,
          offset: append ? items.value.length : 0,
        },
        controller.signal,
    );
    const page = hits.slice(0, PAGE_SIZE);
    items.value = append ? [...items.value, ...page] : page;
    hasMore.value = hits.length > PAGE_SIZE;
  } catch (error) {
    if (controller.signal.aborted) return;
    listError.value = errorMessage(error);
    if (!append) resetItems();
  } finally {
    if (!controller.signal.aborted) {
      listLoading.value = false;
      loadingMore.value = false;
    }
  }
}

async function loadObject() {
  objectController?.abort();
  if (selectedId.value === null) return;

  const controller = new AbortController();
  objectController = controller;
  objectError.value = null;
  objectLoading.value = true;
  object.value = null;

  try {
    object.value = await apiClient.getCatalogObject(
        selectedId.value,
        controller.signal,
    );
  } catch (error) {
    if (!controller.signal.aborted) objectError.value = errorMessage(error);
  } finally {
    if (!controller.signal.aborted) objectLoading.value = false;
  }
}

function applyFilters(next: Filters = {}) {
  clearTimeout(debounceTimer);
  if (next.query !== undefined) query.value = next.query;
  if (next.collection !== undefined) activeCollectionId.value = next.collection;
  syncRoute();
  void loadList();
}

function onQueryInput(value: string | number) {
  query.value = String(value);
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(applyFilters, term.value ? SEARCH_DEBOUNCE_MS : 0);
}

function toggleCollection(id: number) {
  applyFilters({ collection: activeCollectionId.value === id ? null : id });
}

function selectItem(item: SearchHit) {
  selectedId.value = item.id;
  selectedLabel.value = item.common_name ?? item.identifier;
  syncRoute();
  void loadObject();
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

function scrollToTop() {
  window.scrollTo({
    top: 0,
    behavior: prefersReducedMotion.value ? "auto" : "smooth",
  });
}

onMounted(() => {
  void loadCollections();
  if (hasFilters.value) void loadList();
  if (selectedId.value !== null) void loadObject();
});

onBeforeUnmount(() => {
  clearTimeout(debounceTimer);
  listController?.abort();
  objectController?.abort();
});
</script>

<template>
  <div class="mx-auto max-w-6xl px-6 py-8">
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

      <div role="search" class="flex flex-col gap-4">
        <InputGroup>
          <InputGroupAddon>
            <Search aria-hidden="true" />
          </InputGroupAddon>
          <InputGroupInput
              :model-value="query"
              type="text"
              :placeholder="placeholder"
              aria-label="Search objects"
              @update:model-value="onQueryInput"
          />
          <InputGroupAddon v-if="query" align="inline-end">
            <InputGroupButton
                size="icon-xs"
                aria-label="Clear search"
                @click="applyFilters({ query: '' })"
            >
              <X aria-hidden="true" />
            </InputGroupButton>
          </InputGroupAddon>
        </InputGroup>

        <div class="flex flex-col gap-2">
          <div
              role="group"
              aria-label="Collections"
              class="-mx-1 flex gap-2 overflow-x-auto px-1 pb-1 sm:flex-wrap sm:overflow-visible sm:pb-0"
          >
            <Button
                v-for="collection in collections"
                :key="collection.id"
                size="sm"
                class="shrink-0"
                :variant="collection.id === activeCollectionId ? 'default' : 'outline'"
                :aria-pressed="collection.id === activeCollectionId"
                @click="toggleCollection(collection.id)"
            >
              {{ collection.name }}
            </Button>
          </div>
          <RetryAlert
              v-if="collectionsError"
              :message="collectionsError"
              @retry="loadCollections"
          />
          <p
              v-if="activeCollection?.description"
              class="text-sm text-muted-foreground"
          >
            {{ activeCollection.description }}
          </p>
        </div>
      </div>
    </div>

    <div
        class="mt-6 grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]"
    >
      <section aria-label="Results" class="flex flex-col gap-3 pb-20">
        <p role="status" class="min-h-5 text-sm text-muted-foreground">
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
            @suggest="applyFilters({ query: $event })"
        />
        <Button
            v-if="canWidenSearch"
            variant="outline"
            size="sm"
            class="self-center"
            @click="applyFilters({ collection: null })"
        >
          Search all collections
        </Button>
      </section>

      <section
          v-if="isDesktop"
          aria-label="Object details"
          class="sticky top-6 self-start pl-2 pt-2"
      >
        <ObjectDetails v-bind="detailsProps" @retry="loadObject" />
      </section>
    </div>

    <Sheet v-model:open="sheetOpen">
      <SheetContent
          side="bottom"
          class="max-h-[90dvh] overflow-y-auto rounded-t-2xl px-4 pb-6 pt-12"
      >
        <SheetTitle class="sr-only">Object details</SheetTitle>
        <SheetDescription class="sr-only">
          Catalog data for the selected object.
        </SheetDescription>
        <ObjectDetails v-bind="detailsProps" @retry="loadObject" />
      </SheetContent>
    </Sheet>

    <Transition
        enter-active-class="transition duration-200 ease-out motion-reduce:transition-none"
        enter-from-class="translate-y-2 opacity-0"
        leave-active-class="transition duration-150 ease-in motion-reduce:transition-none"
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