<script setup>
import { computed, nextTick, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { Check, Loader2, X } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Stepper,
  StepperItem,
  StepperSeparator,
  StepperTitle,
  StepperTrigger,
} from "@/components/ui/stepper";
import RetryAlert from "@/components/catalog/RetryAlert.vue";

import StepSite from "../steps/StepSite.vue";
import StepTargets from "../steps/StepTargets.vue";
import StepReview from "../steps/StepReview.vue";
import { getOrCreateAccountKey } from "@/lib/account.ts";
import { toIsoWithOffset } from "@/lib/format";
import { apiClient } from "@/api/client";

const emit = defineEmits(["close"]);
const router = useRouter();

const steps = [
  {
    id: "site",
    label: "Site & time",
    title: "Where and when",
    description: "Pin the observing site and set the night you plan to observe.",
  },
  {
    id: "targets",
    label: "Targets",
    title: "What to observe",
    description:
        "Pick the objects you want to hop to. Browse a collection or search by name.",
  },
  {
    id: "review",
    label: "Review",
    title: "Check and create",
    description: "Make sure everything is right. Use Change to fix a detail.",
  },
];

const mission = reactive({
  site: { lat: null, lng: null, time: "", limiting_magnitude: null },
  targets: { objectives: [] },
});

const stepIndex = ref(0);
const stepRef = ref(null);
const mainEl = ref(null);
const headingEl = ref(null);
const isSubmitting = ref(false);
const submitError = ref("");

const step = computed(() => steps[stepIndex.value]);
const isFirst = computed(() => stepIndex.value === 0);
const isLast = computed(() => stepIndex.value === steps.length - 1);
const continueLabel = computed(() => {
  if (isSubmitting.value) return "Creating…";
  return isLast.value ? "Create mission" : "Continue";
});

async function goTo(index) {
  stepIndex.value = index;
  await nextTick();
  mainEl.value?.scrollTo({ top: 0 });
  headingEl.value?.focus();
}

function jumpTo(stepNumber) {
  if (stepNumber - 1 < stepIndex.value) goTo(stepNumber - 1);
}

function back() {
  if (!isFirst.value) goTo(stepIndex.value - 1);
}

function next() {
  if (!(stepRef.value?.validate?.() ?? true)) return;
  if (isLast.value) submit();
  else goTo(stepIndex.value + 1);
}

async function submit() {
  isSubmitting.value = true;
  submitError.value = "";
  try {
    await getOrCreateAccountKey();

    const { site, targets } = mission;
    const { mission_id } = await apiClient.createMission({
      location: { lat: Number(site.lat), long: Number(site.lng) },
      time: toIsoWithOffset(site.time),
      objectives: targets.objectives,
      conditions: { limiting_magnitude: Number(site.limiting_magnitude) },
    });
    router.push(`/missions/${mission_id}`);
  } catch {
    submitError.value = "Could not create the mission. Try again.";
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="flex h-dvh flex-col bg-background text-foreground">
    <header class="shrink-0 border-b border-border px-6 py-4">
      <div class="mx-auto flex w-full max-w-5xl items-start justify-between gap-4">
        <div>
          <h1 class="font-heading text-xl font-semibold leading-tight">
            New mission
          </h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Set up a night of observing, step by step.
          </p>
        </div>
      </div>

      <Stepper
          class="mx-auto mt-6 flex w-full max-w-2xl items-start gap-2"
          :model-value="stepIndex + 1"
          @update:model-value="jumpTo"
      >
        <StepperItem
            v-for="(item, index) in steps"
            :key="item.id"
            v-slot="{ state }"
            class="relative flex flex-1 flex-col items-center gap-2"
            :step="index + 1"
        >
          <StepperSeparator
              v-if="index < steps.length - 1"
              class="absolute left-1/2 top-[18px] h-0.5 w-full -translate-y-1/2 bg-border data-[state=completed]:bg-primary"
          />
          <StepperTrigger as-child :disabled="state === 'inactive'">
            <Button
                :variant="state === 'inactive' ? 'outline' : 'default'"
                size="icon"
                class="z-10 size-9 rounded-full text-xs"
                :aria-current="state === 'active' ? 'step' : undefined"
            >
              <Check v-if="state === 'completed'" class="size-4" aria-hidden="true" />
              <span v-else aria-hidden="true">{{ index + 1 }}</span>
              <span class="sr-only">
                Step {{ index + 1 }}: {{ item.label }}
                <template v-if="state === 'completed'">, completed</template>
              </span>
            </Button>
          </StepperTrigger>
          <StepperTitle
              as="span"
              aria-hidden="true"
              class="text-center text-xs"
              :class="state === 'inactive' ? 'text-muted-foreground' : 'text-foreground'"
          >
            {{ item.label }}
          </StepperTitle>
        </StepperItem>
      </Stepper>
    </header>

    <main ref="mainEl" class="flex-1 overflow-y-auto px-6 py-8">
      <div class="mx-auto w-full max-w-5xl">
        <div class="mb-8 flex max-w-2xl flex-col gap-1">
          <p class="text-xs uppercase tracking-wide text-muted-foreground">
            Step {{ stepIndex + 1 }} of {{ steps.length }}
          </p>
          <h2
              ref="headingEl"
              tabindex="-1"
              class="font-heading text-2xl tracking-tight outline-none"
          >
            {{ step.title }}
          </h2>
          <p class="text-sm text-muted-foreground">{{ step.description }}</p>
        </div>

        <StepSite v-if="step.id === 'site'" ref="stepRef" v-model="mission.site" />
        <StepTargets
            v-else-if="step.id === 'targets'"
            ref="stepRef"
            v-model="mission.targets"
        />
        <StepReview
            v-else
            :mission="mission"
            @edit="goTo(steps.findIndex((s) => s.id === $event))"
        />
      </div>
    </main>

    <footer class="shrink-0 border-t border-border px-6 py-4">
      <div class="mx-auto flex w-full max-w-5xl flex-col gap-3">
        <RetryAlert v-if="submitError" :message="submitError" @retry="submit" />
        <div class="flex items-center justify-between">
          <Button
              variant="outline"
              :disabled="isFirst || isSubmitting"
              @click="back"
          >
            Back
          </Button>
          <Button :disabled="isSubmitting" @click="next">
            <Loader2 v-if="isSubmitting" class="animate-spin" aria-hidden="true" />
            {{ continueLabel }}
          </Button>
        </div>
      </div>
    </footer>
  </div>
</template>