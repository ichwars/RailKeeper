<script setup lang="ts">
import { computed } from "vue";
import { useData, withBase } from "vitepress";
import coverage from "../../coverage.json";

const { lang } = useData();
const german = computed(() => lang.value.startsWith("de"));
const pageLink = (path: string) => withBase(`/${path.replace(/(?:index)?\.md$/, "")}`);
</script>

<template>
  <table>
    <thead>
      <tr>
        <th>{{ german ? "Thema und Anleitung" : "Topic and guide" }}</th>
        <th>{{ german ? "Nachweis des Einstiegs" : "Entry-page evidence" }}</th>
        <th>{{ german ? "Prüfdatum" : "Review date" }}</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="topic in coverage.topics" :key="topic.id">
        <td><a :href="pageLink(german ? topic.germanPath : topic.englishPath)">{{ german ? topic.labelDE : topic.labelEN }}</a></td>
        <td>{{ topic.reviewEvidence === "demo-ui"
          ? (german ? "Quellabgleich + Demo-Oberfläche" : "Source review + demo interface")
          : (german ? "Quellabgleich" : "Source review") }}</td>
        <td><time :datetime="topic.reviewedAt">{{ german ? topic.reviewedAt.split("-").reverse().join(".") : topic.reviewedAt }}</time></td>
      </tr>
    </tbody>
  </table>
</template>
