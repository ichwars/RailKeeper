import React from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import "./app/styles.css";
import { applyThemePreference, readThemePreference } from "./shared/theme";

applyThemePreference(readThemePreference());

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
