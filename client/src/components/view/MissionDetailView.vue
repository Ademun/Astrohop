<script lang="ts" setup>
import { apiClient, ApiError } from "@/api/client";
import { getOrCreateAccountKey } from "@/lib/account";
import { Mission } from "@/types/api";
import { onMounted, onUnmounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import MissionStarMap from "../MissionStarMap.vue";

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
  <div class="mx-auto w-full bg-white">
    <div v-if="isLoading" class="space-y-6" aria-busy="true" aria-live="polite">
      <Skeleton class="h-8 w-2/3 rounded-md" />
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr]">
        <div class="space-y-4">
          <Skeleton class="h-20 w-full rounded-lg" />
          <Skeleton class="h-20 w-full rounded-lg" />
          <Skeleton class="h-20 w-full rounded-lg" />
        </div>
        <Skeleton class="aspect-297/210 w-full rounded-lg" />
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
      <MissionStarMap
        v-if="mission.data && mission.map_data"
        :mission="mission"
      />
    </template>
  </div>
</template>
