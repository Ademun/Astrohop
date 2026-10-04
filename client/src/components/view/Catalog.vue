<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRoute, useRouter, type LocationQueryValue } from "vue-router";
import { useMediaQuery, useWindowScroll } from "@vueuse/core";
import { ArrowUp } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui/sheet";
import { apiClient } from "@/api/client";
import SearchBar from "@/components/catalog/SearchBar.vue";
import ObjectList from "@/components/catalog/ObjectList.vue";
import ObjectDetails from "@/components/catalog/ObjectDetails.vue";
import { useCatalogSearch } from "@/composables/useCatalogSearch";
import { errorMessage } from "@/lib/errors";
import type { CatalogObject, SearchHit } from "@/types/api";

type QueryValue = LocationQueryValue | LocationQueryValue[] | undefined;

const SCROLL_TOP_THRESHOLD_PX = 384;

function firstValue(value: QueryValue): string {
  return (Array.isArray(value) ? value[0] : value) ?? "";
}

function positiveInt(value: QueryValue): number | null {
  const parsed = Number(firstValue(value));
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

const route = useRoute();
const router = useRouter();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const prefersReducedMotion = useMediaQuery("(prefers-reduced-motion: reduce)");
const { y: scrollY } = useWindowScroll();

const {
  query,
  activeCollectionId,
  activeCollection,
  collections,
  collectionsError,
  items,
  hasMore,
  loading: listLoading,
  loadingMore,
  error: listError,
  hasCollection,
  placeholder,
  suggestions,
  summary,
  emptyText,
  canWidenSearch,
  applyFilters,
  onQueryInput,
  toggleCollection,
  loadCollections,
  loadMore,
  retry,
} = useCatalogSearch({
  query: firstValue(route.query.q),
  collection: positiveInt(route.query.collection),
  onFiltersChange: syncRoute,
});

const selectedId = ref(positiveInt(route.query.object));
const selectedLabel = ref<string | null>(null);

const object = ref<CatalogObject | null>(null);
const objectLoading = ref(false);
const objectError = ref<string | null>(null);

let objectController: AbortController | null = null;

const selectedIds = computed(() =>
    selectedId.value === null ? [] : [selectedId.value],
);
const showScrollTop = computed(() => scrollY.value > SCROLL_TOP_THRESHOLD_PX);

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
      q: query.value.trim() || undefined,
      collection: activeCollectionId.value?.toString(),
      object: selectedId.value?.toString(),
    },
  });
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
  if (selectedId.value !== null) void loadObject();
});

onBeforeUnmount(() => objectController?.abort());
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

      <SearchBar
          :query="query"
          :placeholder="placeholder"
          :collections="collections"
          :active-collection-id="activeCollectionId"
          :collections-error="collectionsError"
          :description="activeCollection?.description"
          @query-input="onQueryInput"
          @clear="applyFilters({ query: '' })"
          @toggle-collection="toggleCollection"
          @retry-collections="loadCollections"
      />
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
            :selected-ids="selectedIds"
            :loading="listLoading"
            :loading-more="loadingMore"
            :has-more="hasMore"
            :error="listError"
            :empty-text="emptyText"
            :scoped="hasCollection"
            :suggestions="suggestions"
            @select="selectItem"
            @load-more="loadMore"
            @retry="retry"
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