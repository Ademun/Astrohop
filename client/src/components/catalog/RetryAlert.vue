<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

defineProps<{
  message: string;
}>();

const emit = defineEmits<{
  retry: [];
}>();
</script>

<template>
  <Alert variant="destructive">
    <AlertDescription>
      <p>{{ message }}</p>
      <Button variant="outline" size="sm" class="mt-1" @click="emit('retry')">
        Try again
      </Button>
    </AlertDescription>
  </Alert>
</template>
