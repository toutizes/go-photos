import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
export default defineConfig(({ mode }) => ({
  base: mode === "staging" ? "/app/react-test/" : "/app/",
  plugins: [react()],
  server: { host: "127.0.0.1", proxy: { "/db": "http://127.0.0.1:8080" } },
}));
