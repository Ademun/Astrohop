<script setup>
import { computed, nextTick, ref } from "vue";
import { X } from "@lucide/vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import SearchBar from "@/components/catalog/SearchBar.vue";
import ObjectList from "@/components/catalog/ObjectList.vue";
import { useCatalogSearch } from "@/composables/useCatalogSearch";

const draft = defineModel({ required: true });

const submitted = ref(false);
const errorEl = ref(null);

const {
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
} = useCatalogSearch();

const objectives = computed(() => draft.value.objectives);
const selectedIds = computed(() => objectives.value.map((o) => o.oid));
const showEmptyError = computed(
    () => submitted.value && objectives.value.length === 0,
);

function toggle(item) {
  draft.value = {
    objectives: selectedIds.value.includes(item.id)
        ? objectives.value.filter((o) => o.oid !== item.id)
        : [
          ...objectives.value,
          { oid: item.id, name: item.common_name ?? item.identifier },
        ],
  };
}

function remove(oid) {
  draft.value = { objectives: objectives.value.filter((o) => o.oid !== oid) };
}

function validate() {
  submitted.value = true;
  const valid = objectives.value.length > 0;
  if (!valid) nextTick(() => errorEl.value?.focus());
  return valid;
}

defineExpose({ validate });
</script>

<template>
  <div class="flex max-w-3xl flex-col gap-6">
    <section
        aria-labelledby="selected-heading"
        class="z-10 max-h-40 overflow-y-auto rounded-lg border border-border bg-card p-4 sm:sticky sm:top-0"
    >
      <h3 id="selected-heading" class="text-sm font-medium">
        Selected targets
        <span class="text-muted-foreground">({{ objectives.length }})</span>
      </h3>

      <ul v-if="objectives.length" class="mt-3 flex flex-wrap gap-2">
        <li v-for="objective in objectives" :key="objective.oid">
          <Badge variant="secondary" class="gap-1 py-1 pl-2.5 pr-1">
            {{ objective.name }}
            <button
                type="button"
                class="rounded-full p-0.5 hover:bg-background/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :aria-label="`Remove ${objective.name}`"
                @click="remove(objective.oid)"
            >
              <X class="size-3" aria-hidden="true" />
            </button>
          </Badge>
        </li>
      </ul>
      <p v-else class="mt-2 text-sm text-muted-foreground">
        Nothing yet. Pick objects from the list below.
      </p>

      <p
          v-if="showEmptyError"
          ref="errorEl"
          role="alert"
          tabindex="-1"
          class="mt-2 text-sm text-destructive outline-none"
      >
        Select at least one target to continue.
      </p>
    </section>

    <section aria-label="Browse the catalog" class="flex flex-col gap-4">
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

      <div class="flex flex-col gap-3">
        <p role="status" class="min-h-5 text-sm text-muted-foreground">
          {{ summary }}
        </p>
        <ObjectList
            multiple
            :items="items"
            :selected-ids="selectedIds"
            :loading="loading"
            :loading-more="loadingMore"
            :has-more="hasMore"
            :error="error"
            :empty-text="emptyText"
            :scoped="hasCollection"
            :suggestions="suggestions"
            @select="toggle"
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
      </div>
    </section>
  </div>
</template>