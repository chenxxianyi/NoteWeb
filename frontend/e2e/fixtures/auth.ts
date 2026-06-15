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
    await page.waitForTimeout(2500)

    // Login — page should now be in login mode
    // Re-fill email in case it was cleared
    await page.fill('#email', user.email)
    await page.fill('#password', user.password)
    await page.click('button[type="submit"]')

    // Wait for login redirect to dashboard
    try {
      await page.waitForURL('**/dashboard', { timeout: 10000 })
    } catch {
      // Fallback: navigate manually
      await page.goto('/dashboard')
    }
    await page.waitForTimeout(500)
    await use(page)
    await ctx.close()
  },
})

export { expect } from '@playwright/test'
