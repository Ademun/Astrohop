<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, nextTick } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { Crosshair, Clock } from '@lucide/vue'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Field, FieldError } from '@/components/ui/field'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ lat: null, lng: null, time: null }),
  },
})
const emit = defineEmits(['update:modelValue', 'update:valid'])

// --- Local form state ---------------------------------------------------
// Seeded from the incoming value so the step keeps what was entered
// earlier when the wizard navigates back to it.

const lat = ref(props.modelValue.lat)
const lng = ref(props.modelValue.lng)
const dateTimeLocal = ref(toDateTimeLocal(props.modelValue.time))

function toDateTimeLocal(iso) {
  // "YYYY-MM-DDThh:mm:00+hh:mm" -> "YYYY-MM-DDThh:mm"
  return iso ? iso.slice(0, 16) : ''
}

// The offset always comes from the device's clock for the date in
// question (so daylight saving is handled correctly) — no manual picker.
function offsetMinutesFor(dateTimeLocalStr) {
  const d = dateTimeLocalStr ? new Date(dateTimeLocalStr) : new Date()
  return -d.getTimezoneOffset()
}

function formatOffset(minutes) {
  const sign = minutes >= 0 ? '+' : '-'
  const abs = Math.abs(minutes)
  const hh = String(Math.floor(abs / 60)).padStart(2, '0')
  const mm = String(abs % 60).padStart(2, '0')
  return `${sign}${hh}:${mm}`
}

// --- Validation -----------------------------------------------------

const errors = reactive({ lat: '', lng: '', time: '' })

function validate() {
  errors.lat = lat.value === null || lat.value === '' || Number.isNaN(Number(lat.value))
    ? 'Enter a latitude.'
    : lat.value < -90 || lat.value > 90
      ? 'Latitude must be between -90 and 90.'
      : ''

  errors.lng = lng.value === null || lng.value === '' || Number.isNaN(Number(lng.value))
    ? 'Enter a longitude.'
    : lng.value < -180 || lng.value > 180
      ? 'Longitude must be between -180 and 180.'
      : ''

  errors.time = dateTimeLocal.value ? '' : 'Pick a date and time.'

  const valid = !errors.lat && !errors.lng && !errors.time
  emit('update:valid', valid)
  return valid
}

function emitModel() {
  const valid = validate()
  if (!valid) {
    emit('update:modelValue', { lat: null, lng: null, time: null })
    return
  }
  emit('update:modelValue', {
    lat: Number(lat.value),
    lng: Number(lng.value),
    time: `${dateTimeLocal.value}:00${formatOffset(offsetMinutesFor(dateTimeLocal.value))}`,
  })
}

// --- Map ---------------------------------------------------------------

const mapEl = ref(null)
let map = null
let marker = null
let syncingFromMap = false

const pinIcon = L.divIcon({
  className: '',
  html: '<div class="observation-marker mission-pin"><div class="observation-marker__ring"></div><div class="observation-marker__dot"></div></div>',
  iconSize: [32, 32],
  iconAnchor: [16, 16],
})

function placeMarker(latVal, lngVal) {
  if (!map) return
  const point = [latVal, lngVal]
  if (!marker) {
    marker = L.marker(point, { icon: pinIcon, draggable: true }).addTo(map)
    marker.on('dragend', () => {
      const { lat: dLat, lng: dLng } = marker.getLatLng()
      syncingFromMap = true
      lat.value = round(dLat)
      lng.value = round(dLng)
      syncingFromMap = false
      onFieldChange()
    })
  } else {
    marker.setLatLng(point)
  }
}

function round(n) {
  return Math.round(n * 1e6) / 1e6
}

function useMyLocation() {
  if (!navigator.geolocation || !map) return
  navigator.geolocation.getCurrentPosition((pos) => {
    const newLat = round(pos.coords.latitude)
    const newLng = round(pos.coords.longitude)

    syncingFromMap = true
    lat.value = newLat
    lng.value = newLng
    syncingFromMap = false

    placeMarker(newLat, newLng)
    // Recheck the container size before moving the view — if it changed
    // since the map was created, a stale size makes the first setView
    // land in the wrong place and only a second interaction corrects it.
    map.invalidateSize()
    map.setView([newLat, newLng], 9)

    onFieldChange()
  })
}

function useDeviceTime() {
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  dateTimeLocal.value = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`
  onFieldChange()
}

function onFieldChange() {
  if (!syncingFromMap && lat.value !== null && lng.value !== null && !errors.lat && !errors.lng) {
    const latNum = Number(lat.value)
    const lngNum = Number(lng.value)
    if (!Number.isNaN(latNum) && !Number.isNaN(lngNum)) {
      placeMarker(latNum, lngNum)
    }
  }
  emitModel()
}

onMounted(async () => {
  await nextTick()
  map = L.map(mapEl.value, {
    center: [lat.value ?? 20, lng.value ?? 0],
    zoom: lat.value !== null ? 9 : 2,
    zoomControl: true,
  })

  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    className: 'map-tiles-night',
    maxZoom: 18,
  }).addTo(map)

  map.on('click', (e) => {
    syncingFromMap = true
    lat.value = round(e.latlng.lat)
    lng.value = round(e.latlng.lng)
    syncingFromMap = false
    placeMarker(lat.value, lng.value)
    onFieldChange()
  })

  if (lat.value !== null && lng.value !== null) {
    placeMarker(lat.value, lng.value)
  }

  // The map is laid out inside a flex/grid parent, so its true size can
  // settle a frame after creation — resync once that happens.
  requestAnimationFrame(() => map.invalidateSize())

  validate()
})

onBeforeUnmount(() => {
  map?.remove()
})
</script>

<template>
  <div class="mx-auto grid max-w-5xl gap-8 md:grid-cols-[300px_1fr] md:items-start">
    <!-- Form (left) -->
    <div class="flex flex-col gap-6">
      <div>
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-medium">Location</h2>
          <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2 text-xs" @click="useMyLocation">
            <Crosshair class="size-3.5" />
            Use my location
          </Button>
        </div>

        <div class="mt-3 grid grid-cols-2 gap-3">
          <Field>
            <Label for="lat">Latitude</Label>
            <Input
              id="lat"
              v-model.number="lat"
              type="number"
              step="0.000001"
              min="-90"
              max="90"
              placeholder="52.520008"
              @blur="onFieldChange"
            />
            <FieldError v-if="errors.lat">{{ errors.lat }}</FieldError>
          </Field>

          <Field>
            <Label for="lng">Longitude</Label>
            <Input
              id="lng"
              v-model.number="lng"
              type="number"
              step="0.000001"
              min="-180"
              max="180"
              placeholder="13.404954"
              @blur="onFieldChange"
            />
            <FieldError v-if="errors.lng">{{ errors.lng }}</FieldError>
          </Field>
        </div>
      </div>

      <div>
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-medium">Date & time</h2>
          <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2 text-xs" @click="useDeviceTime">
            <Clock class="size-3.5" />
            Use now
          </Button>
        </div>

        <div class="mt-3">
          <Field>
            <Label for="obs-time">Local date & time</Label>
            <Input
              id="obs-time"
              v-model="dateTimeLocal"
              type="datetime-local"
              @change="onFieldChange"
            />
            <FieldError v-if="errors.time">{{ errors.time }}</FieldError>
          </Field>
        </div>
        <p class="mt-2 text-xs text-muted-foreground">
          Saved with your device's current UTC offset ({{ formatOffset(offsetMinutesFor(dateTimeLocal)) }}).
        </p>
      </div>
    </div>

    <!-- Map (right, larger) -->
    <div>
      <div
        ref="mapEl"
        class="h-72 w-full rounded-lg border border-border md:h-[560px]"
      />
      <p class="mt-2 text-xs text-muted-foreground">
        Click the map to drop a pin, or drag the pin to fine-tune it.
      </p>
    </div>
  </div>
</template>

<style scoped>
/* Darken OSM tiles so a bright basemap doesn't wreck night vision. */
:deep(.map-tiles-night) {
  filter: invert(1) hue-rotate(180deg) brightness(0.85) contrast(1.05) saturate(0.6);
}

:deep(.leaflet-container) {
  background: var(--card);
  font-family: var(--font-sans);
}

/* Bigger, higher-contrast pin so it reads clearly against any tile color
   once the night-mode filter is applied. */
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