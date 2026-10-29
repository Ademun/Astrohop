import { ref, watch } from "vue";
import { useDebounceFn } from "@vueuse/core";

type SaveStatus = "idle" | "saving" | "saved" | "error";

interface Options<T> {
  key: () => string | null;
  value: () => T;
  save: (key: string, value: T) => Promise<void>;
  delay?: number;
}

export function useAutosave<T>({ key, value, save, delay = 600 }: Options<T>) {
  const status = ref<SaveStatus>("idle");

  let baseline: string | null = null;
  let queue: Promise<void> = Promise.resolve();

  function reset() {
    baseline = key() === null ? null : JSON.stringify(value());
    status.value = "idle";
  }

  function flush(): Promise<void> {
    const target = key();
    if (target === null) return queue;

    const snapshot = value();
    const serialized = JSON.stringify(snapshot);
    if (serialized === baseline) return queue;

    baseline = serialized;
    status.value = "saving";

    queue = queue.then(async () => {
      try {
        await save(target, snapshot);
        if (key() === target && baseline === serialized) status.value = "saved";
      } catch {
        if (key() === target) {
          baseline = null;
          status.value = "error";
        }
      }
    });
    return queue;
  }

  const scheduleFlush = useDebounceFn(flush, delay);

  watch(
    () => (key() === null ? null : JSON.stringify(value())),
    (serialized) => {
      if (serialized !== null) scheduleFlush();
    },
  );

  return { status, flush, reset };
}
