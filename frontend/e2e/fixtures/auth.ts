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

    // Navigate to any page first so relative fetch URLs work
    await page.goto('/login')
    await page.waitForTimeout(300)

    // Step 1: Register via backend API directly
    await page.evaluate(async (user) => {
      await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: user.username,
          email: user.email,
          password: user.password,
          confirm_password: user.password,
        }),
      })
    }, user)

    // Step 2: Login via UI
    await page.waitForSelector('#email', { timeout: 10000 })
    await page.fill('#email', user.email)
    await page.fill('#password', user.password)
    await page.click('button[type="submit"]')

    // Step 3: Wait for token + dashboard
    try {
      await page.waitForFunction(
        () => window.localStorage.getItem('token') !== null,
        { timeout: 10000 }
      )
      await page.waitForURL('**/dashboard', { timeout: 5000 }).catch(() => {})
    } catch {
      // If still not logged in, try once more through the UI
      await page.fill('#email', user.email)
      await page.fill('#password', user.password)
      await page.click('button[type="submit"]')
      await page.waitForFunction(
        () => window.localStorage.getItem('token') !== null,
        { timeout: 10000 }
      )
    }

    // Ensure we're on dashboard
    if (!page.url().includes('/dashboard')) {
      await page.goto('/dashboard')
    }
    await page.waitForTimeout(500)

    await use(page)
    await ctx.close()
  },
})

export { expect } from '@playwright/test'
