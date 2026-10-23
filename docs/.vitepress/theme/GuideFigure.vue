<script setup lang="ts">
import { computed } from "vue";
import { useData } from "vitepress";

const props = defineProps<{ name: string; alt: string; caption: string }>();
const { lang } = useData();
const screenshots = import.meta.glob<string>("../../site/assets/screenshots/*/*.png", {
  eager: true,
  query: "?url",
  import: "default",
});
const source = computed(() => screenshots[
  `../../site/assets/screenshots/${lang.value.startsWith("de") ? "de" : "en"}/${props.name}.png`
]);
</script>

<template>
  <figure class="guide-figure">
    <a :href="source"><img :src="source" :alt="alt" loading="lazy" /></a>
    <figcaption>{{ caption }}</figcaption>
  </figure>
</template>
