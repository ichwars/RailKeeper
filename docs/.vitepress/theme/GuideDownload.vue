<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{ file: string }>();
const examples = import.meta.glob<string>("../../site/assets/examples/*.{csv,json,txt,png}", {
  eager: true,
  query: "?url&no-inline",
  import: "default",
});
const source = computed(() => examples[`../../site/assets/examples/${props.file}`]);
</script>

<template>
  <a :href="source" :download="file"><slot /></a>
</template>
