# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e\specs\note.spec.ts >> Notes >> 5. context menu paste inserts clipboard text when allowed
- Location: e2e\specs\note.spec.ts:68:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://127.0.0.1:5173/login
Call log:
  - navigating to "http://127.0.0.1:5173/login", waiting until "load"

```

# Test source

```ts
  1  | import { test as base, type Page } from '@playwright/test'
  2  | 
  3  | export type AuthFixtures = {
  4  |   authedPage: Page
  5  |   user: { username: string; email: string; password: string }
  6  | }
  7  | 
  8  | export const test = base.extend<AuthFixtures>({
  9  |   user: async ({}, use) => {
  10 |     const ts = Date.now()
  11 |     await use({ username: `u_${ts}`, email: `u_${ts}@t.com`, password: 'Pw1234' })
  12 |   },
  13 | 
  14 |   authedPage: async ({ browser, user }, use) => {
  15 |     const ctx = await browser.newContext({
  16 |       baseURL: process.env.BASE_URL || 'http://127.0.0.1:5174',
  17 |     })
  18 |     const page = await ctx.newPage()
  19 | 
  20 |     // Navigate to any page first so relative fetch URLs work
> 21 |     await page.goto('/login')
     |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://127.0.0.1:5173/login
  22 |     await page.waitForTimeout(300)
  23 | 
  24 |     // Step 1: Register via backend API directly
  25 |     await page.evaluate(async (user) => {
  26 |       await fetch('/api/v1/auth/register', {
  27 |         method: 'POST',
  28 |         headers: { 'Content-Type': 'application/json' },
  29 |         body: JSON.stringify({
  30 |           username: user.username,
  31 |           email: user.email,
  32 |           password: user.password,
  33 |           confirm_password: user.password,
  34 |         }),
  35 |       })
  36 |     }, user)
  37 | 
  38 |     // Step 2: Login via UI
  39 |     await page.waitForSelector('#email', { timeout: 10000 })
  40 |     await page.fill('#email', user.email)
  41 |     await page.fill('#password', user.password)
  42 |     await page.click('button[type="submit"]')
  43 | 
  44 |     // Step 3: Wait for token + dashboard
  45 |     try {
  46 |       await page.waitForFunction(
  47 |         () => window.localStorage.getItem('token') !== null,
  48 |         { timeout: 10000 }
  49 |       )
  50 |       await page.waitForURL('**/dashboard', { timeout: 5000 }).catch(() => {})
  51 |     } catch {
  52 |       // If still not logged in, try once more through the UI
  53 |       await page.fill('#email', user.email)
  54 |       await page.fill('#password', user.password)
  55 |       await page.click('button[type="submit"]')
  56 |       await page.waitForFunction(
  57 |         () => window.localStorage.getItem('token') !== null,
  58 |         { timeout: 10000 }
  59 |       )
  60 |     }
  61 | 
  62 |     // Ensure we're on dashboard
  63 |     if (!page.url().includes('/dashboard')) {
  64 |       await page.goto('/dashboard')
  65 |     }
  66 |     await page.waitForTimeout(500)
  67 | 
  68 |     await use(page)
  69 |     await ctx.close()
  70 |   },
  71 | })
  72 | 
  73 | export { expect } from '@playwright/test'
  74 | 
```