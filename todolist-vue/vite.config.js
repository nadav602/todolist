import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// Simple Vite config for Vue with dev proxy to the TodoList API.
// Assumes the API is available on http://localhost:8080.
// Proxies the REST endpoints so the frontend can call them without CORS issues.
export default defineConfig({
  plugins: [vue()],
  base: "./",
  server: {
    port: 5173,
    proxy: {
      "/todos": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/livez": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/readyz": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
