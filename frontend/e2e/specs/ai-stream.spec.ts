import { test, expect } from '../fixtures/auth'

test.describe('AI Panel', () => {

  test('1. 上传文档后阅读器可打开', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload a text file
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'ai-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('This is a test document for AI panel testing.\nIt has some content.'),
    })
    await authedPage.waitForTimeout(3000)

    // Click the card to open reader
    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // Should navigate to reader
    expect(authedPage.url()).toContain('/reader/')

    // Markdown editor should be visible
    await expect(authedPage.locator('.mde-title')).toBeVisible({ timeout: 5000 })
  })

  test('2. AI 助手按钮在编辑器中可见', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload and open
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'ai-panel-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Content for AI panel testing.'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // AI toggle button should be visible in the MarkdownDocumentEditor toolbar
    await expect(authedPage.locator('[aria-label="AI助手"]').first()).toBeVisible()
  })

  test('3. AI 助手面板在 PDF 阅读器中可通过按钮打开', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload a PDF file to test PDF reader AI panel
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles('D:/project/NoteWeb/frontend/e2e/data/sample.pdf')
    await authedPage.waitForTimeout(4000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // PDF reader should have the reader-topbar with AI button
    const topbar = authedPage.locator('.reader-topbar')
    await expect(topbar).toBeVisible({ timeout: 8000 })

    // AI toggle button in the PDF reader toolbar
    const aiBtn = topbar.locator('[aria-label="AI助手"]')
    await expect(aiBtn).toBeVisible()
  })

  test('4. 设置页面 AI 配置可以打开保存', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1000)

    // Find AI provider selector
    const providerSelect = authedPage.locator('select').filter({ hasText: /DeepSeek|OpenAI/ }).first()
    await expect(providerSelect).toBeVisible()

    // Fill API key
    const apiKeyInput = authedPage.locator('input[placeholder="sk-xxxxxxxxxxxxxxxx"]')
    await apiKeyInput.fill('sk-test-key')

    // Click save
    const saveBtn = authedPage.getByRole('button', { name: '保存配置' })
    await saveBtn.click()

    // Should show notification toast (success or error)
    await authedPage.waitForTimeout(2000)
  })

  test('5. 未登录无法访问设置页', async ({ browser }) => {
    const ctx = await browser.newContext()
    const page = await ctx.newPage()
    await page.goto('/settings')
    await page.waitForTimeout(1000)

    // Should redirect to login
    expect(page.url()).toContain('/login')
    await ctx.close()
  })
})
