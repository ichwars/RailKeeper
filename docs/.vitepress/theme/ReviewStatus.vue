<script setup lang="ts">
import { computed } from "vue";
import { useData, withBase } from "vitepress";

const { frontmatter, lang } = useData();
const german = computed(() => lang.value.startsWith("de"));
const date = computed(() => {
  const value = String(frontmatter.value.lastReviewed || "").slice(0, 10);
  return german.value ? value.replace(/^(\d{4})-(\d{2})-(\d{2})$/, "$3.$2.$1") : value;
});
</script>

<template>
  <aside v-if="frontmatter.reviewedVersion" class="doc-review" :aria-label="german ? 'Dokumentationsstand' : 'Documentation status'">
    <span>{{ german ? 'Versionsbezug' : 'Version scope' }}: {{ frontmatter.reviewedVersion }}</span>
    <span v-if="date">{{ german ? 'Redaktionell geprüft' : 'Editorial review' }}: <time :datetime="String(frontmatter.lastReviewed).slice(0, 10)">{{ date }}</time></span>
    <a :href="withBase(german ? '/de/reference/coverage' : '/reference/coverage')">{{ german ? 'Prüfumfang ansehen' : 'See review scope' }}</a>
  </aside>
</template>
