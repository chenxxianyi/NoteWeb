import { test, expect } from '../fixtures/auth'

test.describe('Notes', () => {

  test('1. 笔记列表加载', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    // The notes page should load; either show notes or empty state
    const empty = authedPage.locator('.nl-empty')
    const items = authedPage.locator('.nl-items .nl-item')

    // One of the two should be present
    await expect(empty.or(items).first()).toBeVisible()
  })

  test('2. 笔记列表为空时显示空状态', async ({ authedPage }) => {
    // Fresh user → fresh notes page
    const ctx = await authedPage.context()
    const page = await ctx.newPage()
    const ts = Date.now()
    const email = `notempty_${ts}@test.com`

    await page.goto('/login')
    await page.click('button:has-text("注册")')
    await page.fill('#username', `notempty_${ts}`)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.fill('#confirm-password', 'Test1234')
    page.once('dialog', (d) => d.accept())
    await page.click('button[type="submit"]')
    await page.waitForTimeout(300)
    await page.fill('#email', email)
    await page.fill('#password', 'Test1234')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/dashboard', { timeout: 10000 })

    await page.goto('/notes')
    await page.waitForTimeout(1000)

    await expect(page.locator('.nl-empty')).toContainText('暂无笔记')
    await page.close()
  })

  test('3. 笔记页面布局正确', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    // Should have note list sidebar
    await expect(authedPage.locator('.note-list')).toBeVisible()
    // Should have editor area
    await expect(authedPage.locator('.note-editor')).toBeVisible()
  })
})
