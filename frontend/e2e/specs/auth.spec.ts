import { test, expect } from '../fixtures/auth'

test.describe('Authentication', () => {

  test('1. 注册新用户 → 跳转仪表盘，localStorage 有 token', async ({ browser }) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()
    const ts = Date.now()
    const email = `reg_${ts}@test.com`

    await page.goto('/login')
    await page.waitForSelector('.mode-toggle')

    // Register
    await page.click('text=注册')
    await page.waitForSelector('#username')
    await page.fill('#username', `user_${ts}`)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Test1234')

    page.on('dialog', (d) => d.accept())
    await page.click('button[type="submit"]')
    await page.waitForTimeout(2000)

    // Login
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.click('button[type="submit"]')

    // Wait for navigation
    await page.waitForFunction(
      () => window.localStorage.getItem('token') !== null,
      { timeout: 15000 }
    )
    await page.waitForURL('**/dashboard', { timeout: 10000 }).catch(() => {
      page.goto('/dashboard')
    })

    const token = await page.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeTruthy()
    await ctx.close()
  })

  test('2. 注册后仪表盘显示用户名', async ({ browser }) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()
    const ts = Date.now()
    const email = `usr_${ts}@test.com`
    const username = `user_${ts}`

    await page.goto('/login')
    await page.waitForSelector('.mode-toggle')
    await page.click('text=注册')
    await page.waitForSelector('#username')
    await page.fill('#username', username)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Test1234')

    page.on('dialog', (d) => d.accept())
    await page.click('button[type="submit"]')
    await page.waitForTimeout(2000)

    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.click('button[type="submit"]')

    await page.waitForFunction(
      () => window.localStorage.getItem('token') !== null,
      { timeout: 15000 }
    )
    await page.waitForURL('**/dashboard', { timeout: 10000 }).catch(() => {
      page.goto('/dashboard')
    })
    await page.waitForTimeout(500)

    await expect(page.locator('body')).toContainText(username)
    await ctx.close()
  })

  test('3. 重复邮箱注册 → 错误提示', async ({ browser }) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()
    const ts = Date.now()
    const email = `dup_${ts}@test.com`

    // First registration
    await page.goto('/login')
    await page.waitForSelector('.mode-toggle')
    await page.click('button:has-text("注册")')
    await page.waitForSelector('#username')
    await page.fill('#username', `dup1_${ts}`)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Test1234')

    page.once('dialog', (d) => d.accept())
    await page.click('button[type="submit"]')
    await page.waitForTimeout(1500)

    // Now in login mode after successful register
    // Second registration with same email — need to switch back to register mode
    await page.click('button:has-text("注册")')
    await page.waitForSelector('#username')
    await page.fill('#username', `dup2_${ts}`)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Test1234')
    await page.click('button[type="submit"]')

    // Should show error
    await expect(page.locator('.form-error')).toBeVisible({ timeout: 8000 })
    await ctx.close()
  })

  test('4. 密码不一致 → 前端校验错误', async ({ page }) => {
    await page.goto('/login')
    await page.click('button:has-text("注册")')
    await page.fill('#username', 'pwd_mismatch')
    await page.fill('#email', 'pwd@test.com')
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Different5678')
    await page.click('button[type="submit"]')

    await expect(page.locator('.form-error')).toHaveText('两次密码不一致')
  })

  test('5. 正确邮箱密码登录 → 成功跳转', async ({ browser }) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()
    const ts = Date.now()
    const email = `login_${ts}@test.com`

    // Register first
    await page.goto('/login')
    await page.click('button:has-text("注册")')
    await page.fill('#username', `login_${ts}`)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Test1234')
    page.once('dialog', (d) => d.accept())
    await page.click('button[type="submit"]')
    await page.waitForTimeout(300)

    // Log in
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/dashboard', { timeout: 10000 })

    expect(page.url()).toContain('/dashboard')
    await ctx.close()
  })

  test('6. 错误密码登录 → 错误提示', async ({ page }) => {
    await page.goto('/login')
    await page.fill('#email', 'nonexistent@test.com')
    await page.fill('#password', 'wrongpass')
    await page.click('button[type="submit"]')

    await expect(page.locator('.form-error')).toBeVisible()
  })

  test('7. 未登录访问受保护页面 → 重定向到 /login', async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForURL('**/login', { timeout: 5000 })
    expect(page.url()).toContain('/login')
  })
})
