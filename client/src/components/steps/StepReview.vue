<script setup>
import { computed } from "vue";
import { MapPin, CalendarClock, Telescope, Gauge } from "@lucide/vue";
import { Badge } from "@/components/ui/badge";

// This step doesn't collect its own data — it just displays everything the
// earlier steps already put into the shared mission object.
const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  mission: { type: Object, required: true },
});
defineEmits(["update:modelValue", "update:valid"]);

const coordinatesTime = computed(() => props.mission["coordinates-time"] ?? {});
const objectives = computed(() => props.mission.targets?.objectives ?? []);
const limitingMagnitude = computed(
  () => props.mission.conditions?.limiting_magnitude ?? null,
);

const formattedTime = computed(() => {
  const iso = coordinatesTime.value.time;
  if (!iso) return "Not set";

  // "YYYY-MM-DDThh:mm:00+hh:mm" — split so we display the offset the user
  // actually chose rather than re-deriving it in the browser's own zone.
  const match = iso.match(
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):\d{2}([+-]\d{2}:\d{2})$/,
  );
  if (!match) return iso;

  const [, year, month, day, hour, minute, offset] = match;
  const date = new Date(Date.UTC(+year, +month - 1, +day));
  const monthName = date.toLocaleString("en-US", {
    month: "long",
    timeZone: "UTC",
  });

  return `${monthName} ${+day}, ${year}, ${hour}:${minute} (UTC${offset})`;
});

const formattedCoordinates = computed(() => {
  const { lat, lng } = coordinatesTime.value;
  if (lat === null || lat === undefined || lng === null || lng === undefined)
    return "Not set";
  return `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
});
</script>

<template>
  <div class="mx-auto flex max-w-2xl flex-col gap-4">
    <div>
      <h2 class="text-sm font-medium">Review</h2>
      <p class="mt-1 text-sm text-muted-foreground">
        Check everything below before creating the mission.
      </p>
    </div>

    <div class="divide-y divide-border rounded-lg border border-border bg-card">
      <div class="flex items-center justify-between gap-4 px-4 py-2.5">
        <div class="flex items-center gap-2 text-sm text-muted-foreground">
          <MapPin class="size-4 shrink-0" />
          Location
        </div>
        <p class="text-sm font-medium">{{ formattedCoordinates }}</p>
      </div>

      <div class="flex items-center justify-between gap-4 px-4 py-2.5">
        <div class="flex items-center gap-2 text-sm text-muted-foreground">
          <CalendarClock class="size-4 shrink-0" />
          Date & time
        </div>
        <p class="text-sm font-medium">{{ formattedTime }}</p>
      </div>

      <div class="flex items-center justify-between gap-4 px-4 py-2.5">
        <div
          class="flex shrink-0 items-center gap-2 text-sm text-muted-foreground"
        >
          <Telescope class="size-4 shrink-0" />
          Objectives ({{ objectives.length }})
        </div>
        <div class="flex flex-wrap justify-end gap-1.5">
          <span
            v-if="objectives.length === 0"
            class="text-sm font-medium text-muted-foreground"
          >
            None selected
          </span>
          <Badge
            v-for="obj in objectives"
            :key="obj.oid"
            variant="secondary"
            class="text-xs"
          >
            {{ obj.name }}
          </Badge>
        </div>
      </div>

      <div class="flex items-center justify-between gap-4 px-4 py-2.5">
        <div class="flex items-center gap-2 text-sm text-muted-foreground">
          <Gauge class="size-4 shrink-0" />
          Limiting magnitude
        </div>
        <p class="text-sm font-medium">
          {{
            limitingMagnitude !== null
              ? limitingMagnitude.toFixed(1)
              : "Not set"
          }}
        </p>
      </div>
    </div>
  </div>
</template>
