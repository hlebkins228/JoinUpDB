import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// During development the Vite dev server runs on a separate port from the
// Go API. Requests to `/api` are proxied so the frontend can use the same
// origin in production (where the static bundle would be served by the API).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
});
