<script setup lang="ts">
import { Search, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group";
import RetryAlert from "@/components/catalog/RetryAlert.vue";
import type { Collection } from "@/types/api";

defineProps<{
  query: string;
  placeholder: string;
  collections: Collection[];
  activeCollectionId: number | null;
  collectionsError: string | null;
  description?: string | null;
}>();

const emit = defineEmits<{
  queryInput: [value: string | number];
  clear: [];
  toggleCollection: [id: number];
  retryCollections: [];
}>();
</script>

<template>
  <div role="search" class="flex flex-col gap-4">
    <InputGroup>
      <InputGroupAddon>
        <Search aria-hidden="true" />
      </InputGroupAddon>
      <InputGroupInput
        :model-value="query"
        type="text"
        autocomplete="off"
        :placeholder="placeholder"
        aria-label="Search objects"
        @update:model-value="emit('queryInput', $event)"
      />
      <InputGroupAddon v-if="query" align="inline-end">
        <InputGroupButton
          size="icon-xs"
          aria-label="Clear search"
          @click="emit('clear')"
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
          @click="emit('toggleCollection', collection.id)"
        >
          {{ collection.name }}
        </Button>
      </div>
      <RetryAlert
        v-if="collectionsError"
        :message="collectionsError"
        @retry="emit('retryCollections')"
      />
      <p v-if="description" class="text-sm text-muted-foreground">
        {{ description }}
      </p>
    </div>
  </div>
</template>
