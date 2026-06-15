import { test as base, type Page } from '@playwright/test'

export type AuthFixtures = {
  authedPage: Page
  user: { username: string; email: string; password: string }
}

export const test = base.extend<AuthFixtures>({
  user: async ({}, use) => {
    const ts = Date.now()
    await use({ username: `u_${ts}`, email: `u_${ts}@t.com`, password: 'Pw1234' })
  },

  authedPage: async ({ browser, user }, use) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()

    await page.goto('/login')
    await page.waitForSelector('.mode-toggle', { timeout: 10000 })

    // Register
    await page.click('text=注册')
    await page.waitForSelector('#username', { timeout: 5000 })
    await page.fill('#username', user.username)
    await page.fill('#email', user.email)
    await page.fill('#password', user.password)
    await page.fill('#confirm-password', user.password)

    page.on('dialog', (d) => d.accept())
    await page.click('button[type="submit"]')
    await page.waitForTimeout(2000)

    // Login
    await page.fill('#email', user.email)
    await page.fill('#password', user.password)

    // Use Promise.race to handle navigation
    await Promise.all([
      page.waitForURL('**/dashboard', { timeout: 15000 }),
      page.click('button[type="submit"]'),
    ])

    await page.waitForTimeout(500)
    await use(page)
    await ctx.close()
  },
})

export { expect } from '@playwright/test'
