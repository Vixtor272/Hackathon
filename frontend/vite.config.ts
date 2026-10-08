import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Customer app: WhatsApp simulator, checkout and DeUna (http://localhost:5173).
// The company portal has its own config: vite.empresa.config.ts.
export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: false },
    },
  },
  build: { outDir: 'dist/cliente', emptyOutDir: true },
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
  },
});
