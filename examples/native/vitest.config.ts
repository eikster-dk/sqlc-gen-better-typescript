import { defineConfig } from "vitest/config"

export default defineConfig({
  test: {
    include: ["test/**/*.test.ts"],
    testTimeout: 60000, // 60s for container startup
    hookTimeout: 60000,
    pool: "forks",
    poolOptions: {
      forks: {
        singleFork: false,
      },
    },
  },
})
