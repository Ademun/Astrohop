import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { apiClient } from "@/api/client";
import { errorMessage } from "@/lib/errors";
import type { Collection, SearchHit } from "@/types/api";

interface Filters {
    query?: string;
    collection?: number | null;
}

interface Options {
    query?: string;
    collection?: number | null;
    onFiltersChange?: () => void;
}

const PAGE_SIZE = 50;
const SEARCH_DEBOUNCE_MS = 250;
const SUGGESTIONS = ["Andromeda", "Crab", "Sirius"];

export function useCatalogSearch(options: Options = {}) {
    const query = ref(options.query ?? "");
    const activeCollectionId = ref<number | null>(options.collection ?? null);

    const collections = ref<Collection[]>([]);
    const collectionsError = ref<string | null>(null);

    const items = ref<SearchHit[]>([]);
    const hasMore = ref(false);
    const loading = ref(false);
    const loadingMore = ref(false);
    const error = ref<string | null>(null);

    let debounceTimer: ReturnType<typeof setTimeout> | undefined;
    let controller: AbortController | null = null;

    const term = computed(() => query.value.trim());
    const hasCollection = computed(() => activeCollectionId.value !== null);
    const hasFilters = computed(() => term.value !== "" || hasCollection.value);

    const activeCollection = computed(
        () =>
            collections.value.find((c) => c.id === activeCollectionId.value) ?? null,
    );

    const placeholder = computed(() =>
        activeCollection.value
            ? `Search in ${activeCollection.value.name}`
            : "Search by name, e.g. Andromeda or M31",
    );

    const suggestions = computed(() => (hasFilters.value ? [] : SUGGESTIONS));

    const summary = computed(() => {
        const count = items.value.length;
        if (loading.value || count === 0) return "";
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
            !loading.value &&
            !error.value &&
            items.value.length === 0,
    );

    function resetItems() {
        items.value = [];
        hasMore.value = false;
    }

    async function loadCollections() {
        collectionsError.value = null;
        try {
            collections.value = await apiClient.listCollections();
        } catch (err) {
            collectionsError.value = errorMessage(err);
        }
    }

    async function load(append = false) {
        controller?.abort();
        const current = new AbortController();
        controller = current;

        error.value = null;
        loadingMore.value = append;
        loading.value = !append;

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
                current.signal,
            );
            const page = hits.slice(0, PAGE_SIZE);
            items.value = append ? [...items.value, ...page] : page;
            hasMore.value = hits.length > PAGE_SIZE;
        } catch (err) {
            if (current.signal.aborted) return;
            error.value = errorMessage(err);
            if (!append) resetItems();
        } finally {
            if (!current.signal.aborted) {
                loading.value = false;
                loadingMore.value = false;
            }
        }
    }

    function applyFilters(next: Filters = {}) {
        clearTimeout(debounceTimer);
        if (next.query !== undefined) query.value = next.query;
        if (next.collection !== undefined) activeCollectionId.value = next.collection;
        options.onFiltersChange?.();
        void load();
    }

    function onQueryInput(value: string | number) {
        query.value = String(value);
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(applyFilters, term.value ? SEARCH_DEBOUNCE_MS : 0);
    }

    function toggleCollection(id: number) {
        applyFilters({ collection: activeCollectionId.value === id ? null : id });
    }

    const loadMore = () => load(true);
    const retry = () => load(items.value.length > 0);

    onMounted(() => {
        void loadCollections();
        if (hasFilters.value) void load();
    });

    onBeforeUnmount(() => {
        clearTimeout(debounceTimer);
        controller?.abort();
    });

    return {
        query,
        activeCollectionId,
        activeCollection,
        collections,
        collectionsError,
        items,
        hasMore,
        loading,
        loadingMore,
        error,
        hasCollection,
        hasFilters,
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
    };
}