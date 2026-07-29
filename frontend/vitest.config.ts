import path from 'node:path'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  resolve: {
    alias: {
      '~': path.resolve(__dirname, '.'),
      '@': path.resolve(__dirname, '.')
    }
  },
  test: {
    environment: 'happy-dom',
    include: ['tests/**/*.{spec,test}.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['composables/**/*.ts', 'server/utils/**/*.ts', 'utils/**/*.ts']
    }
  }
})
