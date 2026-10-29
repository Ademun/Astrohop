<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  ChevronRightIcon,
  CompassIcon,
  GlobeIcon,
  PlusIcon,
  RefreshCwIcon,
} from "@lucide/vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import RetryAlert from "@/components/catalog/RetryAlert.vue";
import { getOrCreateAccountKey } from "@/lib/account";
import { errorMessage } from "@/lib/errors";
import { apiClient } from "@/api/client";
import type { MissionShort } from "@/types/mission";

const dateFormatter = new Intl.DateTimeFormat(undefined, { dateStyle: "medium" });
const relativeFormatter = new Intl.RelativeTimeFormat(undefined, {
  numeric: "auto",
});

const RELATIVE_UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 60 * 60 * 24 * 365],
  ["month", 60 * 60 * 24 * 30],
  ["day", 60 * 60 * 24],
  ["hour", 60 * 60],
  ["minute", 60],
];

const missions = ref<MissionShort[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);

let requestId = 0;

const hasMissions = computed(() => missions.value.length > 0);
const showSkeleton = computed(() => isLoading.value && !hasMissions.value);
const sortedMissions = computed(() =>
    [...missions.value].sort(
        (a, b) => Date.parse(b.created_at) - Date.parse(a.created_at),
    ),
);

function title(mission: MissionShort): string {
  return mission.information.name?.trim() || "Untitled mission";
}

function formatRelative(iso: string): string {
  const diffSeconds = Math.round((Date.parse(iso) - Date.now()) / 1000);
  for (const [unit, seconds] of RELATIVE_UNITS) {
    if (Math.abs(diffSeconds) >= seconds) {
      return relativeFormatter.format(Math.round(diffSeconds / seconds), unit);
    }
  }
  return relativeFormatter.format(diffSeconds, "second");
}

async function loadMissions() {
  const id = ++requestId;
  isLoading.value = true;
  error.value = null;

  try {
    apiClient.setToken(await getOrCreateAccountKey());
    const result = await apiClient.listMissions();
    if (id === requestId) missions.value = result;
  } catch (err) {
    if (id === requestId) error.value = errorMessage(err);
  } finally {
    if (id === requestId) isLoading.value = false;
  }
}

onMounted(loadMissions);
</script>

<template>
  <div class="mx-auto w-full max-w-3xl px-4 py-8 sm:px-6 sm:py-12">
    <header
        class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
    >
      <div>
        <h1 class="font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
          My missions
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Plan a night under the sky, or pick up one already mapped out.
        </p>
      </div>

      <div class="flex items-center gap-2 sm:pt-1">
        <Button
            variant="ghost"
            size="icon"
            class="size-8 text-muted-foreground hover:text-foreground"
            :disabled="isLoading"
            aria-label="Refresh missions"
            @click="loadMissions"
        >
          <RefreshCwIcon
              :class="isLoading && 'animate-spin motion-reduce:animate-none'"
              aria-hidden="true"
          />
        </Button>
        <Button as-child size="sm">
          <RouterLink to="/missions/new">
            <PlusIcon aria-hidden="true" />
            New mission
          </RouterLink>
        </Button>
      </div>
    </header>

    <RetryAlert
        v-if="error"
        :message="error"
        class="mb-4"
        @retry="loadMissions"
    />

    <div
        v-if="showSkeleton"
        role="status"
        aria-label="Loading missions"
        class="space-y-2"
    >
      <div
          v-for="n in 4"
          :key="n"
          aria-hidden="true"
          class="flex items-center gap-4 rounded-lg border border-border bg-card px-4 py-3.5"
      >
        <div class="flex-1 space-y-2">
          <Skeleton class="h-4 w-2/5" />
          <Skeleton class="h-3 w-3/5" />
          <Skeleton class="h-3 w-1/4" />
        </div>
        <Skeleton class="size-4 rounded-sm" />
      </div>
    </div>

    <Empty v-else-if="!hasMissions && !error" class="border border-dashed">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <CompassIcon aria-hidden="true" />
        </EmptyMedia>
        <EmptyTitle>No missions yet</EmptyTitle>
        <EmptyDescription>
          Create your first mission to get a star-hopping route for tonight's
          sky.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button as-child size="sm">
          <RouterLink to="/missions/new">New mission</RouterLink>
        </Button>
      </EmptyContent>
    </Empty>

    <ul v-else-if="hasMissions" class="space-y-2">
      <li v-for="mission in sortedMissions" :key="mission.mission_id">
        <RouterLink
            :to="`/missions/${mission.mission_id}`"
            class="group flex items-center gap-4 rounded-lg border border-border bg-card px-4 py-3.5 text-card-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <p class="truncate text-sm font-medium">{{ title(mission) }}</p>
              <Badge v-if="mission.is_public" variant="secondary" class="shrink-0 gap-1">
                <GlobeIcon class="size-3" aria-hidden="true" />
                Public
              </Badge>
            </div>
            <p
                v-if="mission.information.description"
                class="mt-0.5 line-clamp-1 text-sm text-muted-foreground"
            >
              {{ mission.information.description }}
            </p>
            <p class="mt-1 truncate text-xs text-muted-foreground">
              Created
              <time :datetime="mission.created_at">
                {{ dateFormatter.format(new Date(mission.created_at)) }}
              </time>
              · {{ formatRelative(mission.created_at) }}
            </p>
          </div>
          <ChevronRightIcon
              class="size-4 shrink-0 text-muted-foreground transition-transform duration-150 group-hover:translate-x-0.5 motion-reduce:transition-none"
              aria-hidden="true"
          />
        </RouterLink>
      </li>
    </ul>
  </div>
</template>