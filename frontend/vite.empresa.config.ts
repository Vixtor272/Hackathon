import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

const here = (path: string): string => fileURLToPath(new URL(path, import.meta.url));

// Company portal: cashier / courier board and notifications
// (http://localhost:5174). It proxies to the company surface of the API
// (:8081), which does not expose the WhatsApp channel or payments.
export default defineConfig({
  root: here('./empresa'),
  publicDir: false,
  plugins: [svelte({ configFile: here('./svelte.config.js') })],
  server: {
    port: 5174,
    strictPort: true,
    proxy: {
      '/api': { target: 'http://localhost:8081', changeOrigin: false },
    },
  },
  preview: { port: 4174 },
  build: { outDir: here('./dist/empresa'), emptyOutDir: true },
});
