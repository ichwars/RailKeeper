import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { gzipSync } from "node:zlib";
import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";

function compressStaticAssets(): Plugin {
  return {
    name: "railkeeper-static-gzip",
    apply: "build",
    async writeBundle({ dir }, bundle) {
      if (!dir) return;
      const fileNames = Object.keys(bundle).filter(fileName => /^assets\/.+\.(js|css)$/.test(fileName));
      await Promise.all(fileNames.map(async (fileName) => {
        const path = resolve(dir, fileName);
        const source = await readFile(path);
        if (source.length < 1024) return;
        const gzip = gzipSync(source);
        if (gzip.length < source.length) await writeFile(`${path}.gz`, gzip);
      }));
    }
  };
}

export default defineConfig({
  plugins: [react(), compressStaticAssets()],
  server: {
    proxy: {
      "/api": "http://127.0.0.1:8080",
      "/health": "http://127.0.0.1:8080"
    }
  }
});

