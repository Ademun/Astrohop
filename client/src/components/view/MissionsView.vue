<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { CompassIcon, PlusIcon } from '@lucide/vue'
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

const missions = ref<Mission[]>([])
const isLoading = ref(true)
const errorMessage = ref<string | null>(null)

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function formatCreatedAt(iso: string): string {
  return dateFormatter.format(new Date(iso))
}

async function loadMissions(): Promise<void> {
  isLoading.value = true
  errorMessage.value = null

  try {
    const accountKey = await getOrCreateAccountKey()
    apiClient.setToken(accountKey)
    missions.value = await apiClient.listMissions()
  } catch (err) {
    errorMessage.value =
      err instanceof ApiError
        ? err.message
        : 'Could not reach the server. Check your connection and try again.'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadMissions)
</script>

<template>
  <div class="mx-auto w-full max-w-2xl px-4 py-8 sm:px-6 sm:py-12">
    <header class="mb-8">
      <h1 class="font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
        Missions
      </h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Plan a night under the sky, or pick up one already mapped out.
      </p>
    </header>

    <Button
      as-child
      variant="ghost"
      class="h-auto w-full items-center justify-center gap-2 rounded-lg border-2 border-dashed border-primary/40 bg-primary/5 px-6 py-6 text-primary hover:border-primary/70 hover:bg-primary/10 hover:text-primary"
    >
      <RouterLink to="/missions/new">
        <PlusIcon class="size-5" aria-hidden="true" />
        <span class="text-base font-medium">Create a new mission</span>
      </RouterLink>
    </Button>

    <section class="mt-8">
      <div v-if="isLoading" class="space-y-3" aria-busy="true" aria-live="polite">
        <Skeleton class="h-16 w-full rounded-lg" />
        <Skeleton class="h-16 w-full rounded-lg" />
        <Skeleton class="h-16 w-full rounded-lg" />
      </div>

      <Alert v-else-if="errorMessage" variant="destructive">
        <AlertTitle>Missions didn't load</AlertTitle>
        <AlertDescription>
          <p>{{ errorMessage }}</p>
          <Button variant="outline" size="sm" class="mt-3" @click="loadMissions">
            Try again
          </Button>
        </AlertDescription>
      </Alert>

      <Empty v-else-if="missions.length === 0" class="border border-dashed">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <CompassIcon />
          </EmptyMedia>
          <EmptyTitle>No missions yet</EmptyTitle>
          <EmptyDescription>
            Create your first mission to get a star-hopping route for tonight's sky.
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <Button as-child size="sm">
            <RouterLink to="/missions/new">Create a new mission</RouterLink>
          </Button>
        </EmptyContent>
      </Empty>

      <ul v-else class="space-y-3">
        <li v-for="mission in missions" :key="mission.mission_id">
          <RouterLink
            :to="`/missions/${mission.mission_id}`"
            class="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-4 text-card-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <span class="font-medium">Mission {{ formatCreatedAt(mission.created_at) }}</span>
          </RouterLink>
        </li>
      </ul>
    </section>
  </div>
</template>
