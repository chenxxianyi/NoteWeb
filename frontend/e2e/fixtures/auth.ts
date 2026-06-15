import { test as base, type Page } from '@playwright/test'

export type AuthFixtures = {
  /** Page already logged in with a fresh test user */
  authedPage: Page
  /** Credentials of the authed user */
  user: { username: string; email: string; password: string }
}

/** Create a unique test user and log in, returning the page + credentials */
export const test = base.extend<AuthFixtures>({
  user: async ({}, use) => {
    const ts = Date.now()
    await use({
      username: `tester_${ts}`,
      email: `tester_${ts}@test.com`,
      password: 'Test1234',
    })
  },

  authedPage: async ({ browser, user }, use) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()

    // Go to login page
    await page.goto('/login')

    // Switch to register mode
    await page.click('button:has-text("注册")')
    await page.waitForSelector('#username')

    // Fill register form
    await page.fill('#username', user.username)
    await page.fill('#email', user.email)
    await page.fill('#password', user.password)
    await page.fill('#confirm-password', user.password)
    await page.click('button[type="submit"]')

    // After register success → alert → switched to login mode
    // Accept alert
    page.once('dialog', (d) => d.accept())
    await page.waitForTimeout(500)

    // Now login with the same credentials
    await page.fill('#email', user.email)
    await page.fill('#password', user.password)
    await page.click('button[type="submit"]')

    // Wait for redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10000 })
    await page.waitForTimeout(500)

    await use(page)
    await ctx.close()
  },
})

export { expect } from '@playwright/test'
