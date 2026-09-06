<script setup>
import { ref, reactive, computed, markRaw } from "vue";
import { useRouter } from "vue-router";
import { Check, X, Loader2 } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Stepper,
  StepperItem,
  StepperTrigger,
  StepperIndicator,
  StepperSeparator,
  StepperTitle,
} from "@/components/ui/stepper";

import StepCoordinatesTime from "../steps/StepCoordinatesTime.vue";
import StepTargets from "../steps/StepTargets.vue";
import StepConditions from "../steps/StepConditions.vue";
import StepReview from "../steps/StepReview.vue";
import { getOrCreateAccountKey } from "@/lib/account.ts";
import { apiClient } from "@/api/client";

const emit = defineEmits(["close"]);
const router = useRouter();

const steps = [
  {
    id: "coordinates-time",
    title: "Coordinates & time",
    component: markRaw(StepCoordinatesTime),
  },
  { id: "targets", title: "Targets", component: markRaw(StepTargets) },
  { id: "conditions", title: "Conditions", component: markRaw(StepConditions) },
  { id: "review", title: "Review", component: markRaw(StepReview) },
];

const currentIndex = ref(0);
const currentStepNumber = computed(() => currentIndex.value + 1);

// One reactive bucket per step, keyed by step id. Each step component
// owns the shape of its own slice and reports back whether it's valid.
const mission = reactive({
  "coordinates-time": { lat: null, lng: null, time: null },
  targets: { objectives: [] },
  conditions: { limiting_magnitude: null },
  review: {},
});

const stepValidity = reactive({
  "coordinates-time": false,
  targets: false,
  conditions: false,
  review: true,
});

const isFirstStep = computed(() => currentIndex.value === 0);
const isLastStep = computed(() => currentIndex.value === steps.length - 1);
const canAdvance = computed(() => stepValidity[steps[currentIndex.value].id]);

const isSubmitting = ref(false);
const submitError = ref("");

function stepState(index) {
  if (index < currentIndex.value) return "completed";
  if (index === currentIndex.value) return "active";
  return "inactive";
}

function goBack() {
  if (!isFirstStep.value) currentIndex.value -= 1;
}

function goNext() {
  if (isLastStep.value) {
    submitMission();
    return;
  }
  if (canAdvance.value) currentIndex.value += 1;
}

function goToStep(index) {
  // Only allow jumping backward to a step already completed.
  if (index < currentIndex.value) currentIndex.value = index;
}

async function submitMission() {
  isSubmitting.value = true;
  submitError.value = "";
  try {
    await getOrCreateAccountKey();

    const coordinatesTime = mission["coordinates-time"];
    const missionData = {
      location: { lat: coordinatesTime.lat, long: coordinatesTime.lng },
      time: coordinatesTime.time,
      objectives: mission.targets.objectives,
      conditions: { limiting_magnitude: mission.conditions.limiting_magnitude },
    };

    const { mission_id } = await apiClient.createMission(missionData);
    router.push(`/missions/${mission_id}`);
  } catch (err) {
    submitError.value = "Could not create the mission. Try again.";
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="flex h-dvh flex-col bg-background text-foreground">
    <!-- Header -->
    <header class="shrink-0 border-b border-border px-6 py-4">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h1 class="font-heading text-xl font-semibold leading-tight">
            New mission
          </h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Set up a night of observing, step by step.
          </p>
        </div>
        <Button
          variant="ghost"
          size="icon"
          class="shrink-0"
          @click="emit('close')"
        >
          <X class="size-4" />
          <span class="sr-only">Close</span>
        </Button>
      </div>

      <!-- Stepper -->
      <Stepper
        class="mt-6 flex w-full items-start gap-2"
        :model-value="currentStepNumber"
      >
        <StepperItem
          v-for="(step, index) in steps"
          :key="step.id"
          class="relative flex flex-1 flex-col items-center gap-2"
          :step="index + 1"
        >
          <StepperSeparator
            v-if="index !== steps.length - 1"
            class="absolute left-1/2 top-4 h-px w-full -translate-y-1/2 bg-border data-[state=completed]:bg-primary"
          />
          <StepperTrigger
            as-child
            :disabled="index > currentIndex"
            @click="goToStep(index)"
          >
            <button
              type="button"
              class="z-10 flex size-8 shrink-0 items-center justify-center rounded-full border text-xs font-medium transition-colors"
              :class="[
                stepState(index) === 'active' &&
                  'border-primary bg-primary text-primary-foreground',
                stepState(index) === 'completed' &&
                  'border-primary bg-primary text-primary-foreground',
                stepState(index) === 'inactive' &&
                  'border-border bg-background text-muted-foreground',
              ]"
            >
              <Check v-if="stepState(index) === 'completed'" class="size-4" />
              <span v-else>{{ index + 1 }}</span>
            </button>
          </StepperTrigger>
          <StepperTitle
            class="hidden text-center text-xs sm:block"
            :class="
              stepState(index) === 'inactive'
                ? 'text-muted-foreground'
                : 'text-foreground'
            "
          >
            {{ step.title }}
          </StepperTitle>
        </StepperItem>
      </Stepper>
      <!-- Mobile: current step label only -->
      <p class="mt-3 text-center text-xs text-muted-foreground sm:hidden">
        Step {{ currentStepNumber }} of {{ steps.length }} ·
        {{ steps[currentIndex].title }}
      </p>
    </header>

    <!-- Step content -->
    <main class="flex-1 overflow-y-auto px-6 py-6">
      <component
        :is="steps[currentIndex].component"
        v-model="mission[steps[currentIndex].id]"
        :mission="mission"
        @update:valid="(v) => (stepValidity[steps[currentIndex].id] = v)"
      />
    </main>

    <!-- Footer -->
    <footer class="shrink-0 border-t border-border px-6 py-4">
      <Separator class="mb-4 sm:hidden" />
      <p v-if="submitError" class="mb-3 text-sm text-destructive">
        {{ submitError }}
      </p>
      <div class="flex items-center justify-between">
        <Button
          variant="outline"
          :disabled="isFirstStep || isSubmitting"
          @click="goBack"
        >
          Back
        </Button>
        <Button :disabled="!canAdvance || isSubmitting" @click="goNext">
          <Loader2 v-if="isSubmitting" class="size-4 animate-spin" />
          {{
            isSubmitting
              ? "Creating…"
              : isLastStep
                ? "Create mission"
                : "Continue"
          }}
        </Button>
      </div>
    </footer>
  </div>
</template>
