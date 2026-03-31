import { fileURLToPath } from "node:url";

import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

const testRoot = fileURLToPath(new URL(".", import.meta.url));

export default defineConfig({
  root: testRoot,
  plugins: [react()],
  server: {
    fs: {
      allow: [".."],
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    include: ["web/**/*.test.tsx"],
  },
});
