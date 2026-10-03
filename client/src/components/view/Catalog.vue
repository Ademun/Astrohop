<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRoute, useRouter, type LocationQueryValue } from "vue-router";
import { ArrowLeft, Search, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ApiError, apiClient } from "@/api/client";
import ObjectList from "@/components/catalog/ObjectList.vue";
import ObjectDetails from "@/components/catalog/ObjectDetails.vue";
import type {
  CatalogObject,
  Collection,
  CollectionMember,
  SearchHit,
} from "@/types/api";

type ListItem = CollectionMember | SearchHit;
type QueryValue = LocationQueryValue | LocationQueryValue[] | undefined;

const PAGE_SIZE = 50;
const SEARCH_DEBOUNCE_MS = 250;
const DESKTOP_MEDIA_QUERY = "(min-width: 1024px)";
// Примеры запросов для пустого состояния. Подставь то, что точно есть в каталоге.
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

const items = ref<ListItem[]>([]);
const total = ref(0);
const listLoading = ref(false);
const loadingMore = ref(false);
const listError = ref<string | null>(null);

const object = ref<CatalogObject | null>(null);
const objectLoading = ref(false);
const objectError = ref<string | null>(null);

const resultsEl = ref<HTMLElement | null>(null);
const detailsEl = ref<HTMLElement | null>(null);

let debounceTimer: ReturnType<typeof setTimeout> | undefined;
let listController: AbortController | null = null;
let objectController: AbortController | null = null;

const mode = computed(() => {
  if (query.value.trim()) return "search";
  return activeCollectionId.value !== null ? "collection" : "idle";
});

const highlightedCollectionId = computed(() =>
    mode.value === "collection" ? activeCollectionId.value : null,
);

const activeCollection = computed(
    () =>
        collections.value.find((c) => c.id === highlightedCollectionId.value) ??
        null,
);

const hasMore = computed(
    () => mode.value === "collection" && items.value.length < total.value,
);

const summary = computed(() => {
  if (listLoading.value || items.value.length === 0) return null;
  if (mode.value === "search") {
    return `${total.value} ${total.value === 1 ? "match" : "matches"}`;
  }
  return `Showing ${items.value.length.toLocaleString("en-US")} of ${total.value.toLocaleString("en-US")}`;
});

const emptyText = computed(() => {
  if (mode.value === "search") return `No objects match “${query.value.trim()}”.`;
  if (mode.value === "collection") return "This collection is empty.";
  return "Search by name or pick a collection to start browsing.";
});

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
  const collectionId = activeCollectionId.value;

  listError.value = null;
  loadingMore.value = append;
  listLoading.value = !append;
  // Прежние результаты намеренно остаются на экране до прихода новых,
  // чтобы список не мерцал скелетонами на каждый ввод.

  try {
    if (term) {
      const hits = await apiClient.searchCatalog(term, controller.signal);
      items.value = hits;
      total.value = hits.length;
    } else if (collectionId !== null) {
      const page = await apiClient.getCollectionObjects(
          collectionId,
          { limit: PAGE_SIZE, offset: append ? items.value.length : 0 },
          controller.signal,
      );
      items.value = append ? [...items.value, ...page.items] : page.items;
      total.value = page.total;
    } else {
      items.value = [];
      total.value = 0;
    }
  } catch (error) {
    if (!isAbort(error)) {
      listError.value = errorMessage(error);
      if (!append) {
        items.value = [];
        total.value = 0;
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
  activeCollectionId.value = highlightedCollectionId.value === id ? null : id;
  query.value = "";
  applyListChange();
}

function selectItem(item: ListItem) {
  selectedId.value = item.id;
  selectedLabel.value = item.common_name ?? item.identifier;
  syncRoute();
  void loadObject();
  if (!window.matchMedia(DESKTOP_MEDIA_QUERY).matches) {
    detailsEl.value?.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function backToList() {
  const row = resultsEl.value?.querySelector<HTMLElement>(
      '[aria-pressed="true"]',
  );
  (row ?? resultsEl.value)?.scrollIntoView({
    behavior: "smooth",
    block: "center",
  });
  row?.focus({ preventScroll: true });
}

onMounted(() => {
  void loadCollections();
  if (mode.value !== "idle") void loadList();
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
    <!-- Единая панель: заголовок, поиск и коллекции в одном блоке -->
    <div class="flex flex-col gap-4 rounded-xl border border-border bg-card p-4">
      <div class="flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1">
        <h1 class="font-heading text-xl tracking-tight text-foreground">
          Object catalog
        </h1>
        <p class="text-sm text-muted-foreground">
          Stars and deep-sky objects: search by name or browse by collection
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
            placeholder="Search by name, e.g. Andromeda or M31"
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
              :variant="
              collection.id === highlightedCollectionId ? 'default' : 'outline'
            "
              :aria-pressed="collection.id === highlightedCollectionId"
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
      <section
          ref="resultsEl"
          aria-label="Results"
          class="flex flex-col gap-3"
      >
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
            :suggestions="mode === 'idle' ? SUGGESTIONS : []"
            @select="selectItem"
            @load-more="loadList(true)"
            @retry="loadList(items.length > 0)"
            @suggest="applySuggestion"
        />
      </section>

      <!-- Без overflow: иначе подложка карточки обрезается.
           pl-2/pt-2 = запас под её сдвиг на 0.5rem -->
      <section
          ref="detailsEl"
          aria-label="Object details"
          class="scroll-mt-6 pl-2 pt-2 lg:sticky lg:top-6 lg:self-start"
      >
        <Button
            v-if="selectedId !== null"
            variant="ghost"
            size="sm"
            class="mb-2 lg:hidden"
            @click="backToList"
        >
          <ArrowLeft class="size-4" aria-hidden="true" />
          Back to results
        </Button>
        <ObjectDetails
            :object="object"
            :loading="objectLoading"
            :error="objectError"
            :fallback-title="selectedLabel"
            @retry="loadObject"
        />
      </section>
    </div>
  </div>
</template>