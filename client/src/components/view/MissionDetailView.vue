<script setup lang="ts">
import {computed, nextTick, onBeforeUnmount, ref, watch} from "vue";
import {useRoute} from "vue-router";
import {InfoIcon, PlusIcon} from "@lucide/vue";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Skeleton} from "@/components/ui/skeleton";
import {Textarea} from "@/components/ui/textarea";
import RetryAlert from "@/components/catalog/RetryAlert.vue";
import {useAutosave} from "@/composables/useAutosave";
import {getOrCreateAccountKey} from "@/lib/account";
import {errorMessage} from "@/lib/errors";
import {apiClient} from "@/api/client";
import type {MissionFull} from "@/types/mission";

const NAME_MAX_LENGTH = 80;
const DESCRIPTION_MAX_LENGTH = 280;
const DETAILS_ID = "mission-details";

const dateFormatter = new Intl.DateTimeFormat(undefined, {dateStyle: "medium"});
const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});

function formatCoordinate(value: number, positive: string, negative: string): string {
  return `${Math.abs(value).toFixed(4)}° ${value >= 0 ? positive : negative}`;
}

const route = useRoute();

const mission = ref<MissionFull | null>(null);
const name = ref("");
const description = ref("");
const isLoading = ref(true);
const error = ref<string | null>(null);
const isDescriptionOpen = ref(false);
const isDetailsOpen = ref(true);
const descriptionInput = ref<{ $el: HTMLTextAreaElement } | null>(null);

let requestId = 0;

const {
  status: saveStatus,
  flush,
  reset,
} = useAutosave({
  key: () => mission.value?.mission_id ?? null,
  value: () => ({
    name: name.value.trim() || undefined,
    description: description.value.trim() || undefined,
  }),
  save: async (missionId, information) => {
    await apiClient.updateMissionInformation(missionId, information);
  },
});

const saveLabel = computed(() => {
  switch (saveStatus.value) {
    case "saving":
      return "Saving…";
    case "saved":
      return "Saved";
    case "error":
      return "Couldn't save changes";
    default:
      return "";
  }
});

async function loadMission() {
  void flush();

  const id = ++requestId;
  isLoading.value = true;
  error.value = null;
  mission.value = null;

  try {
    apiClient.setToken(await getOrCreateAccountKey());
    const result = await apiClient.getMission(route.params.id as string);
    if (id !== requestId) return;

    mission.value = result;
    name.value = result.information.name ?? "";
    description.value = result.information.description ?? "";
    isDescriptionOpen.value = description.value !== "";
    reset();
  } catch (err) {
    if (id === requestId) error.value = errorMessage(err);
  } finally {
    if (id === requestId) isLoading.value = false;
  }
}

async function openDescription() {
  isDescriptionOpen.value = true;
  await nextTick();
  descriptionInput.value?.$el.focus();
}

function commit(event: Event) {
  (event.target as HTMLElement).blur();
}

watch(() => route.params.id, loadMission, {immediate: true});

onBeforeUnmount(() => {
  void flush();
});
</script>

<template>
  <div class="mx-auto w-full max-w-3xl px-4 py-6 sm:px-6">
    <RetryAlert
        v-if="error"
        :message="error"
        class="mb-4"
        @retry="loadMission"
    />

    <div
        v-if="isLoading"
        role="status"
        aria-label="Loading mission"
        class="space-y-3"
    >
      <Skeleton class="h-9 w-2/3" aria-hidden="true"/>
      <Skeleton class="h-4 w-1/4" aria-hidden="true"/>
    </div>

    <div v-else-if="mission" class="flex flex-col gap-4">
      <h1 class="sr-only">{{ name.trim() || "Untitled mission" }}</h1>

      <div class="flex items-start gap-2">
        <div class="min-w-0 flex-1">
          <label for="mission-name" class="sr-only">Mission name</label>
          <Input
              id="mission-name"
              v-model="name"
              :maxlength="NAME_MAX_LENGTH"
              placeholder="Untitled mission"
              autocomplete="off"
              class="h-auto rounded-none border-0 border-b bg-transparent px-0 py-1 font-heading text-2xl font-semibold shadow-none transition-colors hover:border-border focus-visible:border-foreground focus-visible:ring-0 md:text-2xl dark:bg-transparent"
              :class="name ? 'border-transparent' : 'border-dashed border-border'"
              @blur="flush"
              @keydown.enter="commit"
          />
          <p class="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
            <span>
              Created
              <time :datetime="mission.created_at">
                {{ dateFormatter.format(new Date(mission.created_at)) }}
              </time>
            </span>
            <span
                role="status"
                class="min-h-4"
                :class="{
                'text-destructive': saveStatus === 'error',
                'text-green-600 dark:text-green-500': saveStatus === 'saved',
              }"
            >{{ saveLabel }}</span>
            <button
                v-if="saveStatus === 'error'"
                type="button"
                class="underline underline-offset-2 hover:text-foreground"
                @click="flush"
            >
              Retry
            </button>
          </p>
        </div>

        <Button
            type="button"
            variant="ghost"
            size="icon"
            class="mt-1 size-8 shrink-0 text-muted-foreground hover:text-foreground aria-expanded:bg-accent aria-expanded:text-foreground"
            :aria-expanded="isDetailsOpen"
            :aria-controls="DETAILS_ID"
            aria-label="Mission details"
            @click="isDetailsOpen = !isDetailsOpen"
        >
          <InfoIcon aria-hidden="true"/>
        </Button>
      </div>

      <div>
        <template v-if="isDescriptionOpen">
          <label for="mission-description" class="sr-only">Mission description</label>
          <Textarea
              id="mission-description"
              ref="descriptionInput"
              v-model="description"
              :maxlength="DESCRIPTION_MAX_LENGTH"
              placeholder="Add a description"
              rows="3"
              class="resize-none rounded-none border-0 border-b border-border bg-transparent px-0 shadow-none focus-visible:border-foreground focus-visible:ring-0 dark:bg-transparent"
              @blur="flush"
          />
        </template>
        <Button
            v-else
            type="button"
            variant="ghost"
            size="sm"
            class="-ml-2 text-muted-foreground"
            @click="openDescription"
        >
          <PlusIcon aria-hidden="true"/>
          Add description
        </Button>
      </div>

      <dl
          v-if="mission.data"
          v-show="isDetailsOpen"
          :id="DETAILS_ID"
          class="grid grid-cols-2 gap-x-6 gap-y-4 rounded-lg border border-border bg-card p-4 sm:grid-cols-3"
      >
        <div>
          <dt class="text-xs text-muted-foreground">Location</dt>
          <dd class="mt-0.5 text-sm font-medium tabular-nums">
            {{ formatCoordinate(mission.data.location.lat, "N", "S") }}<br/>
            {{ formatCoordinate(mission.data.location.long, "E", "W") }}
          </dd>
        </div>
        <div>
          <dt class="text-xs text-muted-foreground">Observation time</dt>
          <dd class="mt-0.5 text-sm font-medium">
            <time :datetime="mission.data.time">
              {{ dateTimeFormatter.format(new Date(mission.data.time)) }}
            </time>
          </dd>
        </div>
        <div>
          <dt class="text-xs text-muted-foreground">Limiting magnitude</dt>
          <dd class="mt-0.5 text-sm font-medium tabular-nums">
            {{ mission.data.conditions.limiting_magnitude.toFixed(1) }}
          </dd>
        </div>
        <div
            v-if="mission.data.objectives.length"
            class="col-span-full flex flex-col gap-2"
        >
          <dt class="text-xs text-muted-foreground">
            Objectives ({{ mission.data.objectives.length }})
          </dt>
          <dd class="flex flex-wrap gap-1.5">
            <Badge
                v-for="objective in mission.data.objectives"
                :key="objective.oid"
                variant="secondary"
            >
              {{ objective.name }}
            </Badge>
          </dd>
        </div>
      </dl>
    </div>
  </div>
</template>