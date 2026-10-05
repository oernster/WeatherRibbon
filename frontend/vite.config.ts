/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: { outDir: 'dist', emptyOutDir: true },
  test: {
    include: ['src/**/*.test.{ts,tsx}'],
    exclude: ['dist/**', 'wailsjs/**'],
    // ribbonkit's half of the page is installed as TypeScript source, so it is transformed rather
    // than handed to Node as a built dependency.
    server: { deps: { inline: [/@oernster\/ribbonkit/] } },
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
  },
})
