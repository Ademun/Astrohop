<script setup>
import { RouterLink } from "vue-router";
import { Button } from "@/components/ui/button";
import { Route, Eye, QrCode, Flashlight, Share2 } from "@lucide/vue";
import nebulaImg from "@/assets/img/nebula.webp";
import mockupImg from "@/assets/img/mockup.webp";

const features = [
  {
    icon: Route,
    title: "Step-by-step star-hopping",
    description:
      "Get a route from a bright anchor star to any deep-sky target, sized to your finder scope\u2019s field of view.",
    layoutClass: "lg:col-start-1 lg:border-t-0 lg:pt-0",
  },
  {
    icon: Eye,
    title: "Visibility you can trust",
    description:
      "See what\u2019s actually visible tonight, calculated from atmospheric extinction, your sky\u2019s Bortle class, and the moon.",
    layoutClass: "lg:col-start-3 lg:border-l lg:border-t-0 lg:pt-0",
  },
  {
    icon: QrCode,
    title: "Sync without an account",
    description:
      "Carry a mission to any device with a secret key and a QR scan. No email, no password.",
    layoutClass: "lg:col-start-5 lg:border-l lg:border-t-0 lg:pt-0",
  },
  {
    icon: Flashlight,
    title: "Built for the dark",
    description:
      "Inverted, high-contrast maps read clean under a red headlamp and hold up printed in the field.",
    layoutClass: "lg:col-start-2",
  },
  {
    icon: Share2,
    title: "Share what you found",
    description:
      "Send a read-only map link with your coordinates fuzzed, so your spot stays yours.",
    layoutClass: "lg:col-start-4 lg:border-l",
  },
];
</script>

<template>
  <div class="relative overflow-hidden bg-background">
    <!-- Hero -->
    <section class="relative isolate">
      <!-- background: nebula fading into the solid page background -->
      <div class="absolute inset-0 -z-10 bg-background">
        <img
          :src="nebulaImg"
          alt=""
          class="absolute inset-x-0 top-0 h-[65%] w-full object-cover opacity-80"
          style="
            mask-image: linear-gradient(to bottom, black 55%, transparent 100%);
            -webkit-mask-image: linear-gradient(
              to bottom,
              black 55%,
              transparent 100%
            );
          "
        />

        <!-- gradient that ties the nebula into the page background -->
        <div
          class="absolute inset-0 bg-linear-to-b from-transparent via-background/60 to-background"
        />
      </div>

      <div
        class="mx-auto grid max-w-6xl grid-cols-1 items-center gap-12 px-6 pb-20 pt-24 lg:grid-cols-2 lg:gap-8 lg:pb-32 lg:pt-32"
      >
        <div class="flex flex-col items-start gap-6 text-left">
          <h1
            class="font-heading text-4xl leading-tight text-foreground sm:text-5xl lg:text-6xl"
          >
            Star-hop to anything in the sky
          </h1>
          <p class="max-w-md text-base text-muted-foreground sm:text-lg">
            Astrohop builds print-ready routes from bright anchor stars to your
            target, checks what's really visible tonight, and works offline with
            no account required.
          </p>
          <Button as-child size="lg">
            <RouterLink to="/missions">Browse missions</RouterLink>
          </Button>
        </div>

        <div class="relative flex flex-col gap-3">
          <img
            :src="mockupImg"
            alt="Astrohop star map and star-hopping route preview"
            class="mx-auto w-full max-w-md lg:mx-0 lg:max-w-none"
          />
          <p class="text-center text-sm text-muted-foreground lg:text-left">
            Every route prints with distances, sky conditions and a legend sized
            for your gear.
          </p>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section class="relative border-t border-border bg-background">
      <div class="mx-auto max-w-6xl px-6 py-20 lg:py-28">
        <div class="grid grid-cols-1 lg:grid-cols-6 lg:gap-y-10">
          <div
            v-for="feature in features"
            :key="feature.title"
            class="flex flex-col items-center gap-4 border-t border-border pt-8 text-center first:border-t-0 first:pt-0 lg:col-span-2 lg:border-border lg:px-8 lg:pt-10"
            :class="feature.layoutClass"
          >
            <component
              :is="feature.icon"
              class="h-7 w-7 text-foreground"
              stroke-width="1.5"
            />
            <div class="flex flex-col gap-2">
              <h3 class="font-heading text-lg text-foreground">
                {{ feature.title }}
              </h3>
              <p
                class="mx-auto max-w-60 text-sm leading-relaxed text-muted-foreground"
              >
                {{ feature.description }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
