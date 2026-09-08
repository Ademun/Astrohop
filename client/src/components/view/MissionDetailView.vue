<script lang="ts" setup>
import { apiClient, ApiError } from "@/api/client";
import { getOrCreateAccountKey } from "@/lib/account";
import { Mission } from "@/types/api";
import { onMounted, onUnmounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();
const missionId = route.params.id as string;

const mission = ref<Mission | null>(null);
const isLoading = ref(true);
const errorMessage = ref<string | null>(null);

const isBuildingRoute = ref(false);
const routeError = ref<string | null>(null);
let progressController: AbortController | null = null;

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});

function startRouteStream(): void {
  isBuildingRoute.value = true;
  routeError.value = null;

  progressController = apiClient.streamMissionProgress(missionId, {
    onEvent(event) {
      if (event.progress === "done") {
        if (mission.value && event.payload) {
          mission.value = { ...mission.value, map_data: event.payload };
        }
        isBuildingRoute.value = false;
      } else if (event.progress === "failed") {
        routeError.value = event.error ?? "Building the route failed.";
        isBuildingRoute.value = false;
      }
    },
    onError() {
      routeError.value = "Lost connection while building the route.";
      isBuildingRoute.value = false;
    },
  });
}

async function loadMission(): Promise<void> {
  isLoading.value = true;
  errorMessage.value = null;

  try {
    const accountKey = await getOrCreateAccountKey();
    apiClient.setToken(accountKey);
    mission.value = await apiClient.getMission(missionId);

    if (!mission.value.map_data) {
      startRouteStream();
    }
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      router.replace("/404");
      return;
    }
    errorMessage.value =
      err instanceof ApiError
        ? err.message
        : "Could not reach the server. Check your connection and try again.";
  } finally {
    isLoading.value = false;
  }
}

onMounted(loadMission);
onUnmounted(() => progressController?.abort());
</script>

<template>
  <div class="mx-auto w-full max-w-9/10 px-4 py-8 sm:px-6 sm:py-12">
    <div v-if="isLoading" class="space-y-6" aria-busy="true" aria-live="polite">
      <Skeleton class="h-8 w-2/3 rounded-md" />
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr]">
        <div class="space-y-4">
          <Skeleton class="h-20 w-full rounded-lg" />
          <Skeleton class="h-20 w-full rounded-lg" />
          <Skeleton class="h-20 w-full rounded-lg" />
        </div>
        <Skeleton class="aspect-[297/210] w-full rounded-lg" />
      </div>
    </div>

    <Alert v-else-if="errorMessage" variant="destructive">
      <AlertTitle>Mission didn't load</AlertTitle>
      <AlertDescription>
        <p>{{ errorMessage }}</p>
        <Button variant="outline" size="sm" class="mt-3" @click="loadMission">
          Try again
        </Button>
      </AlertDescription>
    </Alert>

    <template v-else-if="mission">
      <header class="mb-8 font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
          Mission {{ dateFormatter.format(new Date(mission.created_at)) }}
      </header>

      <div
        class="grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr] lg:items-start"
      >
        <aside v-if="mission.data" class="space-y-4">
          <div class="rounded-lg border border-border bg-card p-4">
            <h2 class="text-sm font-medium text-muted-foreground">Location</h2>
            <p class="mt-1">
              {{ mission.data.location.lat }}, {{ mission.data.location.long }}
            </p>
          </div>

          <div class="rounded-lg border border-border bg-card p-4">
            <h2 class="text-sm font-medium text-muted-foreground">Time</h2>
            <p class="mt-1">
              {{ dateFormatter.format(new Date(mission.data.time)) }}
            </p>
          </div>

          <div class="rounded-lg border border-border bg-card p-4">
            <h2 class="text-sm font-medium text-muted-foreground">
              Conditions
            </h2>
            <p class="mt-1">
              Limiting magnitude
              {{ mission.data.conditions.limiting_magnitude }}
            </p>
          </div>

          <div class="rounded-lg border border-border bg-card p-4">
            <h2 class="text-sm font-medium text-muted-foreground">
              Objectives ({{ mission.data.objectives.length }})
            </h2>
            <ul
              v-if="mission.data.objectives.length"
              class="mt-2 max-h-64 space-y-1 overflow-y-auto"
            >
              <li
                v-for="objective in mission.data.objectives"
                :key="objective.oid"
              >
                {{ objective.name }}
              </li>
            </ul>
            <p v-else class="mt-1 text-muted-foreground">No objectives set.</p>
          </div>
        </aside>
        <p v-else class="text-muted-foreground">
          This mission has no data yet.
        </p>

        <div>
          <MissionStarMap
            v-if="mission.data && mission.map_data"
            :mission-data="mission.data"
            :map-data="mission.map_data"
          />

          <div
            v-else-if="isBuildingRoute"
            class="flex items-center gap-2 rounded-lg border border-border bg-card p-4 text-muted-foreground"
          >
            <Spinner class="size-4" />
            <span>Building the route…</span>
          </div>

          <Alert v-else-if="routeError" variant="destructive">
            <AlertTitle>Route didn't build</AlertTitle>
            <AlertDescription>
              <p>{{ routeError }}</p>
              <Button
                variant="outline"
                size="sm"
                class="mt-3"
                @click="startRouteStream"
              >
                Try again
              </Button>
            </AlertDescription>
          </Alert>
        </div>
      </div>
    </template>
  </div>
</template>
