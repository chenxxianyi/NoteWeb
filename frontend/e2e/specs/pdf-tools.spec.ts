import { test, expect } from '../fixtures/auth'

test.describe('PDF Reader', () => {

  test('1. PDF 文件可上传并显示在文件列表中', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload a PDF file
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles('D:/project/NoteWeb/frontend/e2e/data/sample.pdf')
    await authedPage.waitForTimeout(4000)

    // Card should appear in the list
    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
  })

  test('2. 点击 PDF 卡片进入阅读器', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles('D:/project/NoteWeb/frontend/e2e/data/sample.pdf')
    await authedPage.waitForTimeout(4000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(3000)

    // Should navigate to reader
    expect(authedPage.url()).toContain('/reader/')

    // PDF reader-topbar exists (only for PDF files)
    const topbar = authedPage.locator('.reader-topbar')
    await expect(topbar).toBeVisible({ timeout: 8000 })
  })

  test('3. PDF 阅读器顶栏包含标题和基本工具', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles('D:/project/NoteWeb/frontend/e2e/data/sample.pdf')
    await authedPage.waitForTimeout(4000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(3000)

    // PDF reader-topbar should be visible
    const topbar = authedPage.locator('.reader-topbar')
    await expect(topbar).toBeVisible({ timeout: 8000 })

    // PDF title area
    await expect(topbar.locator('.pdf-title')).toBeVisible()
    await expect(topbar.locator('.pdf-title h1')).toBeVisible()

    // PDF action toolbar
    await expect(topbar.locator('.pdf-actions')).toBeVisible()

    // Basic tools: pen, highlighter, eraser
    await expect(topbar.locator('[aria-label="画笔"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="高亮"]')).toBeVisible()
  })

  test('4. PDF 阅读器有翻页和缩放按钮', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles('D:/project/NoteWeb/frontend/e2e/data/sample.pdf')
    await authedPage.waitForTimeout(4000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(3000)

    const topbar = authedPage.locator('.reader-topbar')
    await expect(topbar).toBeVisible({ timeout: 8000 })

    // Page navigation
    await expect(topbar.locator('[aria-label="上一页"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="下一页"]')).toBeVisible()

    // Zoom controls
    await expect(topbar.locator('[aria-label="缩小"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="放大"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="缩放菜单"]')).toBeVisible()
  })

  test('5. PDF 阅读器有 AI、搜索和更多按钮', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles('D:/project/NoteWeb/frontend/e2e/data/sample.pdf')
    await authedPage.waitForTimeout(4000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(3000)

    const topbar = authedPage.locator('.reader-topbar')
    await expect(topbar).toBeVisible({ timeout: 8000 })

    // AI, search, annotation list, more buttons
    await expect(topbar.locator('[aria-label="AI助手"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="搜索文档"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="批注列表"]')).toBeVisible()
    await expect(topbar.locator('[aria-label="更多"]')).toBeVisible()
  })

  test('6. 返回按钮从阅读器回到文件库', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'back-nav-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Testing back navigation.'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // Click back button
    const backBtn = authedPage.locator('[aria-label="返回"]').first()
    await backBtn.click()
    await authedPage.waitForTimeout(1000)

    // Should navigate back to documents
    expect(authedPage.url()).toContain('/documents')
  })
})
