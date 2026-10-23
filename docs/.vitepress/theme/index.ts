import type { EnhanceAppContext } from "vitepress";
import { h } from "vue";
import DefaultTheme from "vitepress/theme";
import GuideFigure from "./GuideFigure.vue";
import CoverageTable from "./CoverageTable.vue";
import ReviewStatus from "./ReviewStatus.vue";

import "./custom.css";

export default {
  extends: DefaultTheme,
  enhanceApp({ app }: EnhanceAppContext) {
    app.component("CoverageTable", CoverageTable);
    app.component("GuideFigure", GuideFigure);
  },
  Layout: () => h(DefaultTheme.Layout, null, { "doc-before": () => h(ReviewStatus) }),
};
