<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import { Search, Loader2, Check, X } from "@lucide/vue";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/api/client";

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ objectives: [] }),
  },
});
const emit = defineEmits(["update:modelValue", "update:valid"]);

const MIN_QUERY_LENGTH = 3;
const DEBOUNCE_MS = 300;

// Objective[] — seeded from whatever the wizard already has for this step.
const objectives = ref([...(props.modelValue.objectives ?? [])]);

const query = ref("");
const results = ref([]); // AstroObject[]
const isSearching = ref(false);
const searchError = ref("");
const dropdownOpen = ref(false);

let debounceHandle = null;
let requestToken = 0; // guards against an older, slower response overwriting a newer one

function onQueryInput() {
  if (!dropdownOpen.value) {
    dropdownOpen.value = true
  }
  clearTimeout(debounceHandle);
  const trimmed = query.value.trim();

  if (trimmed.length < MIN_QUERY_LENGTH) {
    results.value = [];
    searchError.value = "";
    isSearching.value = false;
    return;
  }

  debounceHandle = setTimeout(() => runSearch(trimmed), DEBOUNCE_MS);
}

async function runSearch(name) {
  const token = ++requestToken;
  isSearching.value = true;
  searchError.value = "";
  try {
    const found = await apiClient.searchObjects(name);
    if (token !== requestToken) return;
    results.value = found;
  } catch (err) {
    if (token !== requestToken) return;
    results.value = [];
    searchError.value = "Search failed — try again.";
  } finally {
    if (token === requestToken) isSearching.value = false;
  }
}

function isSelected(oid) {
  return objectives.value.some((o) => o.oid === oid);
}

function selectObject(obj) {
  // A repeat oid replaces the existing entry instead of duplicating it.
  objectives.value = [
    ...objectives.value.filter((o) => o.oid !== obj.oid),
    { oid: obj.oid, name: obj.name },
  ];
  emitModel();

  query.value = "";
  results.value = [];
  dropdownOpen.value = false;
}

function removeObjective(oid) {
  objectives.value = objectives.value.filter((o) => o.oid !== oid);
  emitModel();
}

function emitModel() {
  emit("update:valid", objectives.value.length > 0);
  emit("update:modelValue", { objectives: objectives.value });
}

function openDropdown() {
  dropdownOpen.value = true;
}

function closeDropdownDeferred() {
  // Deferred so a click on a result (mousedown fires first) still registers
  // before the input's blur would otherwise close the dropdown.
  setTimeout(() => {
    dropdownOpen.value = false;
  }, 120);
}

onMounted(() => {
  emit("update:valid", objectives.value.length > 0);
});

onBeforeUnmount(() => clearTimeout(debounceHandle));
</script>

<template>
  <div class="mx-auto flex max-w-2xl flex-col gap-6">
    <div>
      <h2 class="text-sm font-medium">Objectives</h2>
      <p class="mt-1 text-sm text-muted-foreground">
        Search for the deep-sky objects you want to hop to tonight.
      </p>
    </div>

    <div class="relative">
      <Search
        class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
      />
      <Input
        v-model="query"
        type="text"
        placeholder="Search by name — e.g. Andromeda, M13, Ring Nebula"
        class="pl-9"
        @input="onQueryInput"
        @focus="openDropdown"
        @blur="closeDropdownDeferred"
      />
      <Loader2
        v-if="isSearching"
        class="absolute right-3 top-1/2 size-4 -translate-y-1/2 animate-spin text-muted-foreground"
      />

      <div
        v-if="dropdownOpen && query.trim().length > 0"
        class="absolute z-20 mt-1 w-full overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-md"
      >
        <p
          v-if="query.trim().length < MIN_QUERY_LENGTH"
          class="px-3 py-2 text-sm text-muted-foreground"
        >
          Keep typing — {{ MIN_QUERY_LENGTH }} characters minimum.
        </p>

        <p v-else-if="searchError" class="px-3 py-2 text-sm text-destructive">
          {{ searchError }}
        </p>

        <p
          v-else-if="!isSearching && results.length === 0"
          class="px-3 py-2 text-sm text-muted-foreground"
        >
          No matches for "{{ query.trim() }}".
        </p>

        <ul v-else class="max-h-64 overflow-y-auto py-1">
          <li v-for="obj in results" :key="obj.oid">
            <button
              type="button"
              class="flex w-full items-center justify-between gap-3 px-3 py-2 text-left text-sm hover:bg-accent hover:text-accent-foreground"
              @mousedown.prevent="selectObject(obj)"
            >
              <span class="flex flex-col">
                <span class="font-medium">{{ obj.name }}</span>
                <span class="text-xs text-muted-foreground">{{
                  obj.type
                }}</span>
              </span>
              <Check
                v-if="isSelected(obj.oid)"
                class="size-4 shrink-0 text-primary"
              />
            </button>
          </li>
        </ul>
      </div>
    </div>

    <div>
      <h3 class="text-xs font-medium text-muted-foreground">
        Selected ({{ objectives.length }})
      </h3>
      <p
        v-if="objectives.length === 0"
        class="mt-2 text-sm text-muted-foreground"
      >
        No objectives selected yet.
      </p>
      <div v-else class="mt-2 flex flex-wrap gap-2">
        <Badge
          v-for="obj in objectives"
          :key="obj.oid"
          variant="secondary"
          class="gap-1.5 py-1 pl-2.5 pr-1.5"
        >
          {{ obj.name }}
          <button
            type="button"
            class="rounded-full p-0.5 hover:bg-background/50"
            @click="removeObjective(obj.oid)"
          >
            <X class="size-3" />
            <span class="sr-only">Remove {{ obj.name }}</span>
          </button>
        </Badge>
      </div>
    </div>
  </div>
</template>
