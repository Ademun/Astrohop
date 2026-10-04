<script setup>
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from "vue";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { Clock, Crosshair, Loader2 } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { nowLocalInput, utcOffset } from "@/lib/format";

const draft = defineModel({ required: true });

const DEFAULT_ZOOM = 9;

const coordinateFields = [
  { key: "lat", label: "Latitude", min: -90, max: 90, placeholder: "52.520008" },
  { key: "lng", label: "Longitude", min: -180, max: 180, placeholder: "13.404954" },
];

const rootEl = ref(null);
const mapEl = ref(null);
const locating = ref(false);
const locationError = ref("");
const submitted = ref(false);
const touched = reactive({
  lat: false,
  lng: false,
  time: false,
  limiting_magnitude: false,
});

const isBlank = (value) =>
    value === null || value === "" || Number.isNaN(Number(value));

const rules = {
  lat: (value) => {
    if (isBlank(value)) return "Enter a latitude.";
    return Math.abs(value) > 90 ? "Latitude must be between -90 and 90." : "";
  },
  lng: (value) => {
    if (isBlank(value)) return "Enter a longitude.";
    return Math.abs(value) > 180 ? "Longitude must be between -180 and 180." : "";
  },
  time: (value) => (value ? "" : "Pick a date and time."),
  limiting_magnitude: (value) =>
      isBlank(value) ? "Enter the limiting magnitude." : "",
};

const errors = computed(() =>
    Object.fromEntries(
        Object.entries(rules).map(([key, rule]) => [key, rule(draft.value[key])]),
    ),
);

const shownErrors = computed(() =>
    Object.fromEntries(
        Object.entries(errors.value).map(([key, message]) => [
          key,
          touched[key] || submitted.value ? message : "",
        ]),
    ),
);

const offset = computed(() => utcOffset(draft.value.time));

const point = computed(() =>
    errors.value.lat || errors.value.lng
        ? null
        : [Number(draft.value.lat), Number(draft.value.lng)],
);

const describedBy = (key, hintId) =>
    shownErrors.value[key] ? `site-${key}-error` : hintId;

const round = (value) => Math.round(value * 1e6) / 1e6;

let map = null;
let marker = null;
let resizeObserver = null;

const pinIcon = L.divIcon({
  className: "",
  html: '<div class="observation-marker mission-pin"><div class="observation-marker__ring"></div><div class="observation-marker__dot"></div></div>',
  iconSize: [32, 32],
  iconAnchor: [16, 16],
});

function setPoint({ lat, lng }) {
  draft.value.lat = round(lat);
  draft.value.lng = round(lng);
  touched.lat = true;
  touched.lng = true;
}

function syncMarker() {
  if (!map) return;
  if (!point.value) {
    marker?.remove();
    marker = null;
    return;
  }
  if (marker) {
    marker.setLatLng(point.value);
    return;
  }
  marker = L.marker(point.value, {
    icon: pinIcon,
    draggable: true,
    title: "Observation site",
    alt: "Observation site",
  }).addTo(map);
  marker.on("dragend", () => setPoint(marker.getLatLng()));
}

function flyToPoint() {
  if (map && point.value) {
    map.setView(point.value, Math.max(map.getZoom(), DEFAULT_ZOOM));
  }
}

function locate() {
  if (!navigator.geolocation) {
    locationError.value = "Geolocation is not available in this browser.";
    return;
  }
  locating.value = true;
  locationError.value = "";
  navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        locating.value = false;
        setPoint({ lat: coords.latitude, lng: coords.longitude });
        flyToPoint();
      },
      () => {
        locating.value = false;
        locationError.value =
            "Could not get your location. Check the browser permission and try again.";
      },
      { timeout: 10000 },
  );
}

function fillNow() {
  draft.value.time = nowLocalInput();
  touched.time = true;
}

function roundMagnitude() {
  if (!isBlank(draft.value.limiting_magnitude)) {
    draft.value.limiting_magnitude =
        Math.round(Number(draft.value.limiting_magnitude) * 10) / 10;
  }
}

function validate() {
  submitted.value = true;
  const valid = Object.values(errors.value).every((message) => !message);
  if (!valid) {
    nextTick(() =>
        rootEl.value?.querySelector('[aria-invalid="true"]')?.focus(),
    );
  }
  return valid;
}

defineExpose({ validate });

watch(point, syncMarker);

onMounted(() => {
  map = L.map(mapEl.value, {
    center: point.value ?? [20, 0],
    zoom: point.value ? DEFAULT_ZOOM : 2,
  });

  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    className: "map-tiles-night",
    maxZoom: 18,
  }).addTo(map);

  map.on("click", (event) => setPoint(event.latlng));
  syncMarker();

  resizeObserver = new ResizeObserver(() => map.invalidateSize());
  resizeObserver.observe(mapEl.value);
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  map?.remove();
});
</script>

<template>
  <div
      ref="rootEl"
      class="grid gap-8 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)] lg:items-start"
  >
    <div class="flex flex-col gap-8">
      <FieldSet>
        <FieldLegend>Location</FieldLegend>
        <FieldGroup class="gap-4">
          <div class="grid grid-cols-2 gap-3">
            <Field
                v-for="field in coordinateFields"
                :key="field.key"
                :data-invalid="Boolean(shownErrors[field.key])"
            >
              <FieldLabel :for="`site-${field.key}`">{{ field.label }}</FieldLabel>
              <Input
                  :id="`site-${field.key}`"
                  v-model.number="draft[field.key]"
                  type="number"
                  inputmode="decimal"
                  step="any"
                  :min="field.min"
                  :max="field.max"
                  :placeholder="field.placeholder"
                  :aria-invalid="Boolean(shownErrors[field.key])"
                  :aria-describedby="describedBy(field.key)"
                  @blur="touched[field.key] = true"
                  @change="flyToPoint"
              />
              <FieldError
                  v-if="shownErrors[field.key]"
                  :id="`site-${field.key}-error`"
              >
                {{ shownErrors[field.key] }}
              </FieldError>
            </Field>
          </div>

          <Button
              type="button"
              variant="outline"
              size="sm"
              class="self-start"
              :disabled="locating"
              @click="locate"
          >
            <Loader2 v-if="locating" class="animate-spin" aria-hidden="true" />
            <Crosshair v-else aria-hidden="true" />
            {{ locating ? "Locating…" : "Use my location" }}
          </Button>
          <FieldError v-if="locationError">{{ locationError }}</FieldError>
        </FieldGroup>
      </FieldSet>

      <FieldSet>
        <FieldLegend>Date and time</FieldLegend>
        <FieldGroup class="gap-4">
          <Field :data-invalid="Boolean(shownErrors.time)">
            <FieldLabel for="site-time">Local date and time</FieldLabel>
            <Input
                id="site-time"
                v-model="draft.time"
                type="datetime-local"
                :aria-invalid="Boolean(shownErrors.time)"
                :aria-describedby="describedBy('time', 'site-time-hint')"
                @blur="touched.time = true"
            />
            <FieldError v-if="shownErrors.time" id="site-time-error">
              {{ shownErrors.time }}
            </FieldError>
            <FieldDescription v-else id="site-time-hint">
              Saved with this device's UTC offset ({{ offset }}).
            </FieldDescription>
          </Field>

          <Button
              type="button"
              variant="outline"
              size="sm"
              class="self-start"
              @click="fillNow"
          >
            <Clock aria-hidden="true" />
            Use now
          </Button>
        </FieldGroup>
      </FieldSet>

      <FieldSet>
        <FieldLegend>Sky conditions</FieldLegend>
        <Field :data-invalid="Boolean(shownErrors.limiting_magnitude)">
          <FieldLabel for="site-limiting_magnitude">Limiting magnitude</FieldLabel>
          <Input
              id="site-limiting_magnitude"
              v-model.number="draft.limiting_magnitude"
              type="number"
              inputmode="decimal"
              step="0.1"
              placeholder="6.5"
              :aria-invalid="Boolean(shownErrors.limiting_magnitude)"
              :aria-describedby="
              describedBy('limiting_magnitude', 'site-limiting_magnitude-hint')
            "
              @blur="touched.limiting_magnitude = true"
              @change="roundMagnitude"
          />
          <FieldError
              v-if="shownErrors.limiting_magnitude"
              id="site-limiting_magnitude-error"
          >
            {{ shownErrors.limiting_magnitude }}
          </FieldError>
          <FieldDescription v-else id="site-limiting_magnitude-hint">
            The faintest magnitude you expect to see tonight, to one decimal.
            Around 6.5 on a dark night, 3–4 under city lights.
          </FieldDescription>
        </Field>
      </FieldSet>
    </div>

    <div class="flex flex-col gap-2">
      <div
          ref="mapEl"
          role="region"
          aria-label="Map of the observation site"
          class="h-80 w-full overflow-hidden rounded-lg border border-border lg:h-[34rem]"
      />
      <p class="text-xs text-muted-foreground">
        Click the map to drop a pin or drag the pin to fine-tune it. You can
        also type the coordinates.
      </p>
    </div>
  </div>
</template>

<style scoped>
:deep(.map-tiles-night) {
  filter: invert(1) hue-rotate(180deg) brightness(0.85) contrast(1.05)
  saturate(0.6);
}

:deep(.leaflet-container) {
  background: var(--card);
  font-family: var(--font-sans);
}

:deep(.mission-pin) {
  width: 32px;
  height: 32px;
}

:deep(.mission-pin .observation-marker__dot) {
  inset: 10px;
  box-shadow:
      0 0 0 3px var(--background),
      0 0 12px 3px rgba(184, 71, 45, 0.85);
}

:deep(.mission-pin .observation-marker__ring) {
  border-width: 2px;
}
</style>