<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Field, FieldError, FieldDescription } from '@/components/ui/field'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ limiting_magnitude: null }),
  },
})
const emit = defineEmits(['update:modelValue', 'update:valid'])

const limitingMagnitude = ref(props.modelValue.limiting_magnitude)
const error = reactive({ value: '' })

function validate() {
  const num = Number(limitingMagnitude.value)
  error.value =
    limitingMagnitude.value === null || limitingMagnitude.value === '' || Number.isNaN(num)
      ? 'Enter the limiting magnitude.'
      : ''
  const valid = !error.value
  emit('update:valid', valid)
  return valid
}

function onChange() {
  // Keep it to one decimal place, matching what the field is meant to record.
  if (limitingMagnitude.value !== null && limitingMagnitude.value !== '') {
    limitingMagnitude.value = Math.round(Number(limitingMagnitude.value) * 10) / 10
  }

  const valid = validate()
  emit('update:modelValue', {
    limiting_magnitude: valid ? Number(limitingMagnitude.value) : null,
  })
}

onMounted(() => validate())
</script>

<template>
  <div class="mx-auto max-w-md">
    <h2 class="text-sm font-medium">Conditions</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      The faintest magnitude you expect to see from the site tonight.
    </p>

    <Field class="mt-4">
      <Label for="limiting-magnitude">Limiting magnitude</Label>
      <Input
        id="limiting-magnitude"
        v-model.number="limitingMagnitude"
        type="number"
        step="0.1"
        placeholder="6.5"
        @blur="onChange"
        @change="onChange"
      />
      <FieldError v-if="error.value">{{ error.value }}</FieldError>
      <FieldDescription v-else>
        Rounded to one decimal place.
      </FieldDescription>
    </Field>
  </div>
</template>
