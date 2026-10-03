<script setup lang="ts">
import { type Component, computed, ref } from "vue";
import {
  Expand,
  Image as ImageIcon,
  Ruler,
  RulerDimensionLine,
  Sun,
  Telescope,
  Thermometer,
} from "@lucide/vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import RetryAlert from "@/components/catalog/RetryAlert.vue";
import {
  formatAngularSize,
  formatDeclination,
  formatDistance,
  formatNumber,
  formatRightAscension,
} from "@/lib/format";
import type { CatalogObject } from "@/types/api";

interface Row {
  label: string;
  value: string;
}

interface Tile extends Row {
  icon: Component;
  wide?: boolean;
}

const props = defineProps<{
  object: CatalogObject | null;
  loading: boolean;
  error: string | null;
  fallbackTitle: string | null;
}>();

const emit = defineEmits<{
  retry: [];
}>();

const imageOpen = ref(false);

const title = computed(
    () =>
        props.object?.metadata.common_name ??
        props.fallbackTitle ??
        `Object #${props.object?.id}`,
);

const tiles = computed<Tile[]>(() => {
  const object = props.object;
  if (!object) return [];

  const { star, dso } = object;
  const magnitude = star?.visual_mag ?? dso?.visual_mag;
  const result: Tile[] = [];

  if (magnitude != null) {
    result.push({
      label: "Visual magnitude",
      value: formatNumber(magnitude, 2),
      icon: Sun,
    });
  }
  if (star?.spectral_class) {
    result.push({
      label: "Spectral class",
      value: star.spectral_class,
      icon: Thermometer,
    });
  }
  if (dso && (dso.major_axis != null || dso.minor_axis != null)) {
    result.push({
      label: "Apparent size",
      value: formatAngularSize(dso.major_axis, dso.minor_axis),
      icon: RulerDimensionLine,
    });
  }
  if (object.distance_pc != null) {
    result.push({
      label: "Distance",
      value: formatDistance(object.distance_pc),
      icon: Ruler,
      wide: true,
    });
  }
  return result;
});

const reference = computed<Row[]>(() => {
  const object = props.object;
  if (!object) return [];

  const rows: Row[] = [
    {
      label: "RA / Dec",
      value: `${formatRightAscension(object.position.ra)} ${formatDeclination(object.position.dec)}`,
    },
  ];
  if (object.dso?.pos_angle != null) {
    rows.push({
      label: "Position angle",
      value: formatNumber(object.dso.pos_angle, 0, "°"),
    });
  }
  return rows;
});
</script>

<template>
  <div
      :data-class="object?.class"
      class="relative isolate before:absolute before:inset-0 before:-z-10 before:-translate-x-2 before:-translate-y-2 before:rounded-xl before:bg-(--class-color)"
  >
    <div class="@container rounded-xl border border-border bg-card p-6">
      <div
          v-if="loading"
          role="status"
          aria-label="Loading object details"
          class="flex flex-col gap-4"
      >
        <Skeleton class="h-5 w-1/3" />
        <Skeleton class="h-7 w-2/3" />
        <Skeleton class="h-4 w-full" />
        <div class="grid grid-cols-2 gap-3">
          <Skeleton class="h-16" />
          <Skeleton class="h-16" />
        </div>
      </div>

      <RetryAlert v-else-if="error" :message="error" @retry="emit('retry')" />

      <div v-else-if="object" class="flex flex-col gap-6">
        <header class="flex flex-col gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <Badge
                variant="secondary"
                class="rounded-full bg-(--class-color)/15 uppercase text-(--class-color)"
            >
              {{ object.class }}
            </Badge>
            <span
                v-if="object.metadata.type_name"
                class="text-sm text-muted-foreground"
            >
              {{ object.metadata.type_name }}
            </span>
          </div>
          <h2 class="font-heading text-2xl tracking-tight text-foreground">
            {{ title }}
          </h2>
          <p
              v-if="object.metadata.description"
              class="text-sm leading-relaxed text-muted-foreground"
          >
            {{ object.metadata.description }}
          </p>
          <Button
              v-if="object.metadata.image_url"
              variant="outline"
              size="sm"
              class="self-start @md:hidden"
              @click="imageOpen = true"
          >
            <ImageIcon aria-hidden="true" />
            View image
          </Button>
        </header>

        <div class="flex flex-col gap-6 @md:flex-row @md:items-start">
          <div class="flex min-w-0 flex-1 flex-col gap-6">
            <dl v-if="tiles.length" class="grid grid-cols-2 gap-3">
              <div
                  v-for="tile in tiles"
                  :key="tile.label"
                  class="flex flex-col gap-1.5 rounded-lg bg-muted/40 px-3 py-2.5"
                  :class="{ 'col-span-2': tile.wide }"
              >
                <dt
                    class="flex items-center gap-1.5 text-xs uppercase tracking-wide text-muted-foreground"
                >
                  <component
                      :is="tile.icon"
                      class="size-3.5 text-(--class-color)"
                      aria-hidden="true"
                  />
                  {{ tile.label }}
                </dt>
                <dd class="text-lg font-medium tabular-nums text-foreground">
                  {{ tile.value }}
                </dd>
              </div>
            </dl>

            <dl class="flex flex-col divide-y divide-border text-sm">
              <div
                  v-for="row in reference"
                  :key="row.label"
                  class="flex items-baseline justify-between gap-4 py-2 first:pt-0 last:pb-0"
              >
                <dt class="text-muted-foreground">{{ row.label }}</dt>
                <dd class="text-right font-mono tabular-nums text-foreground">
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </div>

          <button
              v-if="object.metadata.image_url"
              type="button"
              class="group relative hidden aspect-square w-48 shrink-0 overflow-hidden rounded-lg @md:block @2xl:w-64"
              :aria-label="`Open image of ${title}`"
              @click="imageOpen = true"
          >
            <img
                :src="object.metadata.image_url"
                alt=""
                loading="lazy"
                class="size-full object-cover"
            />
            <span
                aria-hidden="true"
                class="pointer-events-none absolute inset-0 rounded-lg ring-1 ring-inset ring-(--class-color)/50 transition-shadow group-hover:ring-2 group-hover:ring-(--class-color)/90 group-focus-visible:ring-2 group-focus-visible:ring-(--class-color)"
            />
            <span
                aria-hidden="true"
                class="absolute right-2 top-2 rounded-md bg-black/60 p-1.5 text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100"
            >
              <Expand class="size-4" />
            </span>
          </button>
        </div>

        <Dialog v-if="object.metadata.image_url" v-model:open="imageOpen">
          <DialogContent
              class="w-fit max-w-[95vw] gap-0 border-0 bg-transparent p-0 shadow-none sm:max-w-[95vw]"
          >
            <DialogTitle class="sr-only">{{ title }}</DialogTitle>
            <DialogDescription class="sr-only">Full-size image</DialogDescription>
            <img
                :src="object.metadata.image_url"
                :alt="title"
                class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain"
            />
          </DialogContent>
        </Dialog>
      </div>

      <Empty v-else class="border-0">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Telescope aria-hidden="true" />
          </EmptyMedia>
          <EmptyDescription>Select an object to see its details.</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  </div>
</template>