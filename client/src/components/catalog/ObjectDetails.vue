<script setup lang="ts">
import {type Component, computed, ref} from "vue";
import {Expand, Image as ImageIcon, Ruler, Scaling, Sun, Telescope, Thermometer, X,} from "@lucide/vue";
import {Button} from "@/components/ui/button";
import {Skeleton} from "@/components/ui/skeleton";
import {formatAngularSize, formatDeclination, formatDistance, formatNumber, formatRightAscension,} from "@/lib/format";
import type {CatalogObject} from "@/types/api";

interface Tile {
  label: string;
  value: string;
  icon: Component;
  wide?: boolean;
}

interface Row {
  label: string;
  value: string;
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

const dialogRef = ref<HTMLDialogElement | null>(null);

const title = computed(
    () =>
        props.object?.metadata.common_name ??
        props.fallbackTitle ??
        `Object #${props.object?.id}`,
);

const tiles = computed<Tile[]>(() => {
  const object = props.object;
  if (!object) return [];

  const result: Tile[] = [];
  const magnitude = object.star?.visual_mag ?? object.dso?.visual_mag ?? null;

  if (magnitude != null) {
    result.push({
      label: "Visual magnitude",
      value: formatNumber(magnitude, 2),
      icon: Sun,
    });
  }
  if (object.star?.spectral_class) {
    result.push({
      label: "Spectral class",
      value: object.star.spectral_class,
      icon: Thermometer,
    });
  }
  if (
      object.dso &&
      (object.dso.major_axis != null || object.dso.minor_axis != null)
  ) {
    result.push({
      label: "Apparent size",
      value: formatAngularSize(object.dso.major_axis, object.dso.minor_axis),
      icon: Scaling,
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

  const result: Row[] = [
    {
      label: "RA / Dec",
      value: `${formatRightAscension(object.position.ra)} ${formatDeclination(object.position.dec)}`,
    },
  ];
  if (object.dso?.pos_angle != null) {
    result.push({
      label: "Position angle",
      value: formatNumber(object.dso.pos_angle, 0, "°"),
    });
  }
  return result;
});

function openImage() {
  dialogRef.value?.showModal();
}

function closeImage() {
  dialogRef.value?.close();
}
</script>

<template>
  <div class="object-card" :data-class="object?.class">
    <div class="@container rounded-xl border border-border bg-card p-6">
      <div v-if="loading" class="flex flex-col gap-4">
        <Skeleton class="h-5 w-1/3"/>
        <Skeleton class="h-7 w-2/3"/>
        <Skeleton class="h-4 w-full"/>
        <div class="grid grid-cols-2 gap-3">
          <Skeleton class="h-16"/>
          <Skeleton class="h-16"/>
        </div>
      </div>

      <div v-else-if="error" class="flex flex-col items-start gap-3">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" @click="emit('retry')">
          Try again
        </Button>
      </div>

      <div v-else-if="object" class="flex flex-col gap-6">
        <header class="flex flex-col gap-3">
          <div class="flex flex-wrap items-center gap-2">
          <span
              class="rounded-full bg-(--class-color)/15 px-2.5 py-0.5 text-xs font-medium uppercase text-(--class-color)"
          >
            {{ object.class }}
          </span>
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
              @click="openImage"
          >
            <ImageIcon class="size-4" aria-hidden="true"/>
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
                  :class="{'col-span-2': tile.wide}"
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
                <dd class="text-right font-[monospace] tabular-nums text-foreground">
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
              @click="openImage"
          >
            <img
                :src="object.metadata.image_url"
                :alt="title"
                loading="lazy"
                class="size-full object-cover"
            />
            <span
                aria-hidden="true"
                class="pointer-events-none absolute inset-0 rounded-lg ring-1 ring-inset ring-(--class-color)/50 transition-shadow group-hover:ring-2 group-hover:ring-(--class-color)/90 group-focus-visible:ring-2 group-focus-visible:ring-(--class-color)"
            />
            <span
                class="absolute right-2 top-2 rounded-md bg-black/60 p-1.5 text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100"
            >
            <Expand class="size-4" aria-hidden="true"/>
          </span>
          </button>
        </div>

        <dialog
            v-if="object.metadata.image_url"
            ref="dialogRef"
            class="m-auto max-h-[95vh] max-w-[95vw] bg-transparent p-0 backdrop:bg-black/85"
            @click="closeImage"
        >
          <div class="relative">
            <img
                :src="object.metadata.image_url"
                :alt="title"
                class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain"
            />
            <button
                type="button"
                class="absolute right-2 top-2 rounded-md bg-black/60 p-2 text-white hover:bg-black/80"
                aria-label="Close"
                @click.stop="closeImage"
            >
              <X class="size-4" aria-hidden="true"/>
            </button>
          </div>
        </dialog>
      </div>

      <div
          v-else
          class="flex flex-col items-center gap-3 py-16 text-center text-muted-foreground"
      >
        <Telescope class="size-8" :stroke-width="1.5" aria-hidden="true"/>
        <p class="text-sm">Select an object to see its details.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.object-card {
  position: relative;
  isolation: isolate;
}

.object-card::before {
  content: "";
  position: absolute;
  inset: 0;
  z-index: -1;
  border-radius: var(--radius-xl);
  background: var(--class-color);
  transform: translate(-8px, -8px);
}
</style>