<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  ChevronRightIcon,
  CompassIcon,
  PlusIcon,
  RefreshCwIcon,
} from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { getOrCreateAccountKey } from '@/lib/account'
import { apiClient, ApiError } from '@/api/client'
import type { Mission } from '@/types/api'

/* ── state ──────────────────────────────────────────────────────────── */

const missions = ref<Mission[]>([])
const isInitialLoading = ref(true)
const isRefreshing = ref(false)
const errorMessage = ref<string | null>(null)

const hasMissions = computed(() => missions.value.length > 0)

/* Request guard so a slow first fetch can't clobber a fast second one. */
let requestId = 0

/* ── formatting ─────────────────────────────────────────────────────── */

const absoluteFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})
const relativeFormatter = new Intl.RelativeTimeFormat(undefined, {
  numeric: 'auto',
})

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ['year', 60 * 60 * 24 * 365],
  ['month', 60 * 60 * 24 * 30],
  ['day', 60 * 60 * 24],
  ['hour', 60 * 60],
  ['minute', 60],
]

function formatAbsolute(iso: string) {
  return absoluteFormatter.format(new Date(iso))
}

function formatRelative(iso: string) {
  const diffSec = Math.round((new Date(iso).getTime() - Date.now()) / 1000)
  for (const [unit, seconds] of UNITS) {
    if (Math.abs(diffSec) >= seconds) {
      return relativeFormatter.format(Math.round(diffSec / seconds), unit)
    }
  }
  return relativeFormatter.format(diffSec, 'second')
}

/* ── data loading ───────────────────────────────────────────────────── */

async function loadMissions({ silent = false } = {}) {
  const id = ++requestId

  if (silent) isRefreshing.value = true
  else isInitialLoading.value = true
  errorMessage.value = null

  try {
    const accountKey = await getOrCreateAccountKey()
    apiClient.setToken(accountKey)
    const result = await apiClient.listMissions()
    if (id !== requestId) return // superseded
    missions.value = result
  } catch (err) {
    if (id !== requestId) return
    errorMessage.value =
      err instanceof ApiError
        ? err.message
        : 'Could not reach the server. Check your connection and try again.'
  } finally {
    if (id === requestId) {
      isInitialLoading.value = false
      isRefreshing.value = false
    }
  }
}

onMounted(() => loadMissions())
</script>

<template>
  <div class="mx-auto w-full max-w-3xl px-4 py-8 sm:px-6 sm:py-12">
    <!-- ── Header ───────────────────────────────────────────────── -->
    <header
      class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
    >
      <div>
        <h1
          class="font-heading text-2xl font-semibold tracking-tight sm:text-3xl"
        >
          Missions
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Plan a night under the sky, or pick up one already mapped out.
        </p>
        <p
          v-if="hasMissions"
          class="mt-2 text-xs text-muted-foreground/80"
          aria-live="polite"
        >
          {{ missions.length }}
          {{ missions.length === 1 ? 'mission' : 'missions' }}
        </p>
      </div>

      <div class="flex items-center gap-2 sm:pt-1">
        <Button
          variant="ghost"
          size="icon"
          class="size-8 text-muted-foreground hover:text-foreground"
          :disabled="isRefreshing || isInitialLoading"
          aria-label="Refresh missions"
          @click="loadMissions({ silent: true })"
        >
          <RefreshCwIcon
            class="size-4"
            :class="isRefreshing && 'animate-spin'"
            aria-hidden="true"
          />
        </Button>
        <Button as-child size="sm">
          <RouterLink to="/missions/new">
            <PlusIcon class="size-4" aria-hidden="true" />
            New mission
          </RouterLink>
        </Button>
      </div>
    </header>

    <!-- ── Inline error during refresh (stale data still visible) ── -->
    <Alert
      v-if="errorMessage && hasMissions"
      variant="destructive"
      class="mb-4"
    >
      <AlertTitle>Couldn't refresh</AlertTitle>
      <AlertDescription>
        {{ errorMessage }}
        <Button
          variant="outline"
          size="sm"
          class="mt-2"
          @click="loadMissions({ silent: true })"
        >
          Try again
        </Button>
      </AlertDescription>
    </Alert>

    <!-- ── Initial loading skeletons (match card shape) ──────────── -->
    <div
      v-if="isInitialLoading"
      class="space-y-2"
      aria-busy="true"
      aria-live="polite"
    >
      <div
        v-for="n in 4"
        :key="n"
        class="flex items-center gap-4 rounded-lg border border-border bg-card px-4 py-3.5"
      >
        <div class="flex-1 space-y-2">
          <Skeleton class="h-4 w-2/5" />
          <Skeleton class="h-3 w-1/4" />
        </div>
        <Skeleton class="size-4 rounded-sm" />
      </div>
    </div>

    <!-- ── Hard error, nothing to show ───────────────────────────── -->
    <Alert v-else-if="errorMessage && !hasMissions" variant="destructive">
      <AlertTitle>Missions didn't load</AlertTitle>
      <AlertDescription>
        <p>{{ errorMessage }}</p>
        <Button
          variant="outline"
          size="sm"
          class="mt-3"
          @click="loadMissions()"
        >
          Try again
        </Button>
      </AlertDescription>
    </Alert>

    <!-- ── Empty state ───────────────────────────────────────────── -->
    <Empty v-else-if="!hasMissions" class="border border-dashed">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <CompassIcon />
        </EmptyMedia>
        <EmptyTitle>No missions yet</EmptyTitle>
        <EmptyDescription>
          Create your first mission to get a star-hopping route for tonight's
          sky.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button as-child size="sm">
          <RouterLink to="/missions/new">Create a new mission</RouterLink>
        </Button>
      </EmptyContent>
    </Empty>

    <!-- ── Mission list ──────────────────────────────────────────── -->
    <ul v-else class="space-y-2">
      <li v-for="mission in missions" :key="mission.mission_id">
        <RouterLink
          :to="`/missions/${mission.mission_id}`"
          class="group flex items-center gap-4 rounded-lg border border-border bg-card px-4 py-3.5 text-card-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium">
              {{ formatAbsolute(mission.created_at) }}
            </p>
            <p class="mt-0.5 truncate text-xs text-muted-foreground">
              Created {{ formatRelative(mission.created_at) }}
            </p>
          </div>
          <ChevronRightIcon
            class="size-4 shrink-0 text-muted-foreground transition-transform duration-150 group-hover:translate-x-0.5"
            aria-hidden="true"
          />
        </RouterLink>
      </li>
    </ul>
  </div>
</template>