import { defineConfig } from 'vitest/config';
import { fileURLToPath } from 'node:url';

export default defineConfig({
  esbuild: {
    jsx: 'automatic',
  },
  resolve: {
    alias: {
      'better-sqlite3': fileURLToPath(new URL('./tests/mocks/better-sqlite3.js', import.meta.url)),
    },
  },
  test: {
    globals: true,
    environment: 'node',
    include: ['tests/**/*.test.ts', 'src/**/__tests__/**/*.test.ts'],
    benchmark: {
      include: ['tests/bench/**/*.bench.ts', 'tests/bench/**/*-bench.ts'],
      reporter: 'default',
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.ts'],
      exclude: ['src/persist/sqlite-checkpoint.ts'],
      thresholds: {
        // v8 覆盖率在不同 Node 版本间存在行/语句计数差异；CI 的 ts-test
        // 使用 Node 24（需要原生 .ts 执行能力），当前实测约 64.8%。
        // 阈值按 CI 实际运行环境校准，留约 0.5 个百分点余量。
        lines: 64,
        functions: 63,
        branches: 57,
        statements: 64,
      },
    },
  },
});
