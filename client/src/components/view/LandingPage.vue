<script setup lang="ts">
import { RouterLink } from "vue-router";
import {
  ArrowRight,
  Eye,
  Flashlight,
  MapPin,
  Printer,
  QrCode,
  Route,
  Share2,
} from "@lucide/vue";
import { Button } from "@/components/ui/button";
import nebulaImg from "@/assets/img/nebula.webp";
import mockupImg from "@/assets/img/mockup.webp";
import andromedaImg from "@/assets/img/andromeda.webp";

const MOCKUP_SIZE = { width: 1200, height: 900 };

const stepImages = import.meta.glob<string>("@/assets/img/steps/*.webp", {
  eager: true,
  import: "default",
});

function stepImage(name: string): string | undefined {
  return Object.entries(stepImages).find(([path]) =>
      path.endsWith(`/${name}.webp`),
  )?.[1];
}

const steps = [
  {
    icon: MapPin,
    title: "Choose a site and targets",
    description:
        "Pin your observing spot, pick the night, and choose what to find from the catalog.",
    image: stepImage("plan"),
    imageAlt:
        "The mission wizard with a pinned observing site and a list of selected targets",
  },
  {
    icon: Route,
    title: "Get your route",
    description:
        "Astrohop checks what's visible from your sky and plans the hops from a bright anchor star.",
    image: stepImage("route"),
    imageAlt:
        "A star map with a hop-by-hop route from an anchor star to a deep-sky target",
  },
  {
    icon: Printer,
    title: "Take it outside",
    description:
        "Print the map or open it on your phone. Move it to another device with a QR scan.",
    image: stepImage("outside"),
    imageAlt: "A printed star-hopping map next to a phone showing the same mission",
  },
];

const primaryFeatures = [
  {
    icon: Route,
    title: "Step-by-step star-hopping",
    description:
        "A route from a bright anchor star to any deep-sky target, sized to your finder scope's field of view.",
  },
  {
    icon: Eye,
    title: "Visibility you can trust",
    description:
        "See what's actually visible tonight, calculated from atmospheric extinction, your sky's Bortle class, and the moon.",
  },
];

const secondaryFeatures = [
  {
    icon: QrCode,
    title: "Sync without an account",
    description:
        "Carry a mission to any device with a secret key and a QR scan. No email, no password.",
  },
  {
    icon: Flashlight,
    title: "Built for the dark",
    description:
        "Inverted, high-contrast maps read clean under a red headlamp and hold up printed in the field.",
  },
  {
    icon: Share2,
    title: "Share what you found",
    description:
        "Send a read-only map link with your coordinates fuzzed, so your spot stays yours.",
  },
];
</script>

<template>
  <div class="overflow-hidden bg-background">
    <section class="relative isolate">
      <div aria-hidden="true" class="absolute inset-0 -z-10 bg-background">
        <img
            :src="nebulaImg"
            alt=""
            decoding="async"
            class="absolute inset-x-0 top-0 h-[70%] w-full object-cover opacity-70"
            style="
            mask-image: linear-gradient(to bottom, black 35%, transparent 100%);
            -webkit-mask-image: linear-gradient(
              to bottom,
              black 35%,
              transparent 100%
            );
          "
        />
        <div
            class="absolute inset-0 bg-linear-to-b from-background/70 via-background/30 to-transparent lg:bg-linear-to-r lg:from-background/85 lg:via-background/45"
        />
      </div>

      <div
          class="mx-auto grid max-w-6xl grid-cols-1 items-center gap-12 px-6 pb-20 pt-20 sm:pt-24 lg:grid-cols-2 lg:gap-16 lg:pb-32 lg:pt-32"
      >
        <div class="flex flex-col items-start gap-6">
          <h1
              class="font-heading text-4xl leading-[1.1] tracking-tight text-foreground sm:text-5xl lg:text-6xl"
          >
            Star-hop to anything in the sky
          </h1>

          <p class="max-w-md text-base text-muted-foreground sm:text-lg">
            Astrohop builds print-ready routes from bright anchor stars to your
            target, checks what's really visible tonight, and works with no
            account required.
          </p>

          <div class="flex flex-wrap items-center gap-3">
            <Button as-child size="lg">
              <RouterLink to="/missions/new">
                Plan a mission
                <ArrowRight aria-hidden="true" />
              </RouterLink>
            </Button>
            <Button as-child variant="outline" size="lg">
              <a href="#how-it-works">See how it works</a>
            </Button>
          </div>
        </div>

        <figure class="flex flex-col gap-3">
          <img
              :src="mockupImg"
              :width="MOCKUP_SIZE.width"
              :height="MOCKUP_SIZE.height"
              fetchpriority="high"
              alt="Astrohop star map and star-hopping route preview"
              class="mx-auto h-auto w-full max-w-md lg:mx-0 lg:max-w-none"
          />
          <figcaption
              class="text-center text-sm text-muted-foreground lg:text-left"
          >
            A preview of a mission on a piece of A4 paper.
          </figcaption>
        </figure>
      </div>
    </section>

    <section
        id="how-it-works"
        class="relative scroll-mt-16 border-t border-border bg-background"
    >
      <div class="mx-auto max-w-6xl px-6 py-20 lg:py-28">
        <div class="mb-12 max-w-2xl lg:mb-16">
          <h2
              class="font-heading text-3xl tracking-tight text-foreground sm:text-4xl"
          >
            How it works
          </h2>
          <p class="mt-3 text-base text-muted-foreground">
            From an idea to a route in your hand in three steps.
          </p>
        </div>

        <ol class="grid grid-cols-1 gap-12 md:grid-cols-3 md:gap-8">
          <li v-for="(step, index) in steps" :key="step.title" class="flex flex-col gap-5">
            <div
                class="aspect-[4/3] overflow-hidden rounded-xl border border-border bg-card"
            >
              <img
                  v-if="step.image"
                  :src="step.image"
                  :alt="step.imageAlt"
                  loading="lazy"
                  decoding="async"
                  class="size-full object-cover"
              />
              <div
                  v-else
                  aria-hidden="true"
                  class="flex size-full items-center justify-center text-muted-foreground/60"
              >
                <component :is="step.icon" class="size-10" :stroke-width="1.25" />
              </div>
            </div>

            <div class="flex flex-col gap-2">
              <div class="flex items-center gap-3">
                <span
                    aria-hidden="true"
                    class="flex size-7 shrink-0 items-center justify-center rounded-full border border-border text-xs tabular-nums text-muted-foreground"
                >
                  {{ index + 1 }}
                </span>
                <h3 class="font-heading text-lg text-foreground">
                  {{ step.title }}
                </h3>
              </div>
              <p class="text-sm leading-relaxed text-muted-foreground">
                {{ step.description }}
              </p>
            </div>
          </li>
        </ol>
      </div>
    </section>

    <section
        id="features"
        class="relative scroll-mt-16 border-t border-border bg-background"
    >
      <div class="mx-auto max-w-6xl px-6 py-20 lg:py-28">
        <div class="mb-12 max-w-2xl lg:mb-16">
          <h2
              class="font-heading text-3xl tracking-tight text-foreground sm:text-4xl"
          >
            Built for the field
          </h2>
          <p class="mt-3 text-base text-muted-foreground">
            Everything you need to plan a night under the sky — and nothing you
            don't.
          </p>
        </div>

        <ul class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <li
              v-for="feature in primaryFeatures"
              :key="feature.title"
              class="flex flex-col gap-5 rounded-xl border border-border bg-card p-8"
          >
            <span
                class="flex size-12 items-center justify-center rounded-lg bg-primary/15 text-primary"
            >
              <component
                  :is="feature.icon"
                  class="size-6"
                  :stroke-width="1.5"
                  aria-hidden="true"
              />
            </span>
            <div class="flex flex-col gap-2">
              <h3 class="font-heading text-xl text-foreground">
                {{ feature.title }}
              </h3>
              <p class="text-base leading-relaxed text-muted-foreground">
                {{ feature.description }}
              </p>
            </div>
          </li>
        </ul>

        <ul class="mt-12 grid grid-cols-1 gap-x-10 gap-y-10 md:grid-cols-3">
          <li
              v-for="feature in secondaryFeatures"
              :key="feature.title"
              class="flex flex-col gap-3 border-t border-border pt-6"
          >
            <component
                :is="feature.icon"
                class="size-6 text-foreground"
                :stroke-width="1.5"
                aria-hidden="true"
            />
            <div class="flex flex-col gap-2">
              <h3 class="font-heading text-lg text-foreground">
                {{ feature.title }}
              </h3>
              <p class="text-sm leading-relaxed text-muted-foreground">
                {{ feature.description }}
              </p>
            </div>
          </li>
        </ul>
      </div>
    </section>

    <section
        class="relative isolate overflow-hidden border-t border-border bg-background"
    >
      <div
          aria-hidden="true"
          class="pointer-events-none absolute inset-0 -z-10"
      >
        <img
            :src="andromedaImg"
            alt=""
            loading="lazy"
            decoding="async"
            class="size-full object-cover object-[70%_100%] opacity-40"
            style="
            mask-image: radial-gradient(
              ellipse 90% 100% at 70% 100%,
              black 0%,
              black 25%,
              transparent 72%
            );
            -webkit-mask-image: radial-gradient(
              ellipse 90% 100% at 70% 100%,
              black 0%,
              black 25%,
              transparent 72%
            );
          "
        />
      </div>

      <div class="mx-auto max-w-2xl px-6 py-24 text-center lg:py-32">
        <svg
            viewBox="0 0 320 80"
            fill="none"
            class="mx-auto mb-8 h-16 w-full max-w-sm"
            aria-hidden="true"
        >
          <defs>
            <linearGradient id="route-line" x1="20" x2="300" y1="0" y2="0" gradientUnits="userSpaceOnUse">
              <stop offset="0" data-class="star" style="stop-color: var(--class-color)" />
              <stop offset="1" data-class="dso" style="stop-color: var(--class-color)" />
            </linearGradient>
          </defs>
          <path
              d="M20 60 Q 80 10, 150 40 T 300 25"
              stroke="url(#route-line)"
              stroke-width="1.25"
              stroke-dasharray="2 5"
              stroke-linecap="round"
          />
          <g data-class="star" transform="translate(20 60)" class="text-(--class-color)">
            <circle r="9" fill="currentColor" opacity="0.15" />
            <circle r="3" fill="currentColor" />
          </g>
          <circle cx="150" cy="40" r="1.5" fill="currentColor" class="text-foreground/60" />
          <g
              data-class="dso"
              transform="translate(300 25)"
              stroke="currentColor"
              stroke-width="1.2"
              stroke-linecap="round"
              class="text-(--class-color)"
          >
            <circle r="6" fill="none" />
            <line x1="-10" x2="-4" y1="0" y2="0" />
            <line x1="4" x2="10" y1="0" y2="0" />
            <line x1="0" x2="0" y1="-10" y2="-4" />
            <line x1="0" x2="0" y1="4" y2="10" />
            <circle r="1.5" fill="currentColor" stroke="none" />
          </g>
        </svg>

        <h2
            class="font-heading text-3xl tracking-tight text-foreground sm:text-4xl"
        >
          Ready for tonight?
        </h2>
        <p class="mx-auto mt-3 max-w-md text-base text-muted-foreground">
          Pick a target, get a route, and take it outside.
        </p>

        <div class="mt-8 flex flex-wrap justify-center gap-3">
          <Button as-child size="lg">
            <RouterLink to="/missions/new">
              Plan a mission
              <ArrowRight aria-hidden="true" />
            </RouterLink>
          </Button>
        </div>
      </div>
    </section>

    <footer class="border-t border-border bg-background">
      <div
          class="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-8 text-sm text-muted-foreground sm:flex-row sm:items-start sm:justify-between"
      >
        <p class="max-w-2xl leading-relaxed">
          Andromeda (M31): NASA, ESA, J. Dalcanton, B. F. Williams, L. C.
          Johnson, the PHAT team, and R. Gendler. Cropped and colour-graded.
          <a
              href="https://creativecommons.org/licenses/by/4.0/"
              rel="noopener"
              class="underline underline-offset-4 hover:text-foreground"
          >
            CC BY 4.0
          </a>.
        </p>
      </div>
    </footer>
  </div>
</template>