import { defineConfig } from "vitest/config"

export default defineConfig({
  test: {
    include: ["test/**/*.test.ts"],
    testTimeout: 60000, // 60s for container startup
    hookTimeout: 60000,
    pool: "forks", // Use forks for better isolation with testcontainers
    poolOptions: {
      forks: {
        singleFork: false,
      },
    },
  },
})
