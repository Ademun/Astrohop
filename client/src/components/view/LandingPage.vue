<script setup lang="ts">
import { RouterLink } from "vue-router";
import {
  ArrowRight,
  Eye,
  Flashlight,
  QrCode,
  Route,
  Share2,
  Sparkles,
} from "@lucide/vue";
import { Button } from "@/components/ui/button";
import nebulaImg from "@/assets/img/nebula.webp";
import mockupImg from "@/assets/img/mockup.webp";
import andromedaImg from "@/assets/img/andromeda.webp";

const features = [
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
    <!-- ═══════════ Hero ═══════════ -->
    <section class="relative isolate">
      <div class="absolute inset-0 -z-10 bg-background">
        <img
          :src="nebulaImg"
          alt=""
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
              <RouterLink to="/missions">
                Browse missions
                <ArrowRight class="size-4" aria-hidden="true" />
              </RouterLink>
            </Button>
            <Button as-child variant="outline" size="lg">
              <a href="#features">See how it works</a>
            </Button>
          </div>
        </div>

        <figure class="flex flex-col gap-3">
          <img
            :src="mockupImg"
            alt="Astrohop star map and star-hopping route preview"
            class="mx-auto w-full max-w-md lg:mx-0 lg:max-w-none"
          />
          <figcaption
            class="text-center text-sm text-muted-foreground lg:text-left"
          >
            A preview of a mission on a piece of A4 paper.
          </figcaption>
        </figure>
      </div>
    </section>

    <!-- ═══════════ Features ═══════════ -->
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

        <div
          class="grid grid-cols-1 gap-x-10 gap-y-12 sm:grid-cols-2 lg:grid-cols-6"
        >
          <div
            v-for="(feature, index) in features"
            :key="feature.title"
            class="flex flex-col gap-4 sm:last:col-span-2"
            :class="index < 3 ? 'lg:col-span-2' : 'lg:col-span-3'"
          >
            <component
              :is="feature.icon"
              class="size-7 text-foreground"
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
          </div>
        </div>
      </div>
    </section>

    <!-- ═══════════ Closing CTA ═══════════ -->
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
          class="absolute bottom-0 right-0 w-auto max-w-none object-cover object-bottom opacity-40"
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
          class="mx-auto mb-8 h-16 w-full max-w-sm text-foreground/50"
          aria-hidden="true"
        >
          <path
            d="M20 60 Q 80 10, 150 40 T 300 25"
            stroke="currentColor"
            stroke-width="1"
            stroke-dasharray="2 5"
            stroke-linecap="round"
          />
          <g transform="translate(20 60)">
            <circle r="9" fill="currentColor" opacity="0.12" />
            <circle r="3" fill="currentColor" />
          </g>
          <circle cx="150" cy="40" r="1.5" fill="currentColor" opacity="0.6" />
          <g
            transform="translate(300 25)"
            stroke="currentColor"
            stroke-width="1.2"
            stroke-linecap="round"
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
              <ArrowRight class="size-4" aria-hidden="true" />
            </RouterLink>
          </Button>
        </div>

        <p class="mt-10 text-sm leading-relaxed text-muted-foreground/50">
          Andromeda (M31): NASA, ESA, J. Dalcanton, B. F. Williams, L. C.
          Johnson, the PHAT team, and R. Gendler. Cropped and colour-graded. CC
          BY 4.0.
        </p>
      </div>
    </section>
  </div>
</template>
