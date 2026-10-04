<script setup>
import { computed } from "vue";
import { CalendarClock, Gauge, MapPin, Telescope } from "@lucide/vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatLocalInput, utcOffset } from "@/lib/format";

const props = defineProps({
  mission: { type: Object, required: true },
});
const emit = defineEmits(["edit"]);

const groups = computed(() => {
  const { site, targets } = props.mission;
  return [
    {
      id: "site",
      title: "Site and time",
      rows: [
        {
          icon: MapPin,
          label: "Location",
          value: `${Number(site.lat).toFixed(6)}, ${Number(site.lng).toFixed(6)}`,
        },
        {
          icon: CalendarClock,
          label: "Date and time",
          value: `${formatLocalInput(site.time)} (UTC${utcOffset(site.time)})`,
        },
        {
          icon: Gauge,
          label: "Limiting magnitude",
          value: Number(site.limiting_magnitude).toFixed(1),
        },
      ],
    },
    {
      id: "targets",
      title: "Targets",
      rows: [
        {
          icon: Telescope,
          label: `Objectives (${targets.objectives.length})`,
          objectives: targets.objectives,
        },
      ],
    },
  ];
});
</script>

<template>
  <div class="flex max-w-2xl flex-col gap-4">
    <section
        v-for="group in groups"
        :key="group.id"
        class="rounded-lg border border-border bg-card"
    >
      <header
          class="flex items-center justify-between border-b border-border px-4 py-2"
      >
        <h3 class="text-sm font-medium">{{ group.title }}</h3>
        <Button variant="link" size="sm" @click="emit('edit', group.id)">
          Change<span class="sr-only"> {{ group.title.toLowerCase() }}</span>
        </Button>
      </header>

      <dl class="divide-y divide-border">
        <div
            v-for="row in group.rows"
            :key="row.label"
            class="grid gap-1 px-4 py-3 text-sm sm:grid-cols-[11rem_minmax(0,1fr)] sm:gap-4"
        >
          <dt class="flex items-center gap-2 text-muted-foreground">
            <component :is="row.icon" class="size-4 shrink-0" aria-hidden="true" />
            {{ row.label }}
          </dt>
          <dd class="font-medium">
            <ul v-if="row.objectives" class="flex flex-wrap gap-1.5">
              <li v-for="objective in row.objectives" :key="objective.oid">
                <Badge variant="secondary">{{ objective.name }}</Badge>
              </li>
            </ul>
            <template v-else>{{ row.value }}</template>
          </dd>
        </div>
      </dl>
    </section>
  </div>
</template>