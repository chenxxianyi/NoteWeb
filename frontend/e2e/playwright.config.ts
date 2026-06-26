import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './specs',
  fullyParallel: false,
  retries: 1,
  workers: 1,
  reporter: 'list',
  timeout: 60000,

  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:5177',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
