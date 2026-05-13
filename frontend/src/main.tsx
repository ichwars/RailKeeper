import React from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import "./app/styles.css";
import { readLanguage, setLanguage } from "./shared/i18n";
import { applyThemePreference, readThemePreference } from "./shared/theme";

setLanguage(readLanguage());
applyThemePreference(readThemePreference());

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
