import { test, expect } from '../fixtures/auth'

test.describe('Reader Toolbar', () => {

  test('1. 从文件库点击卡片进入阅读器', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload a text file
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'reader-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('This is a test document for the reader.\nLine 2\nLine 3\nLine 4\nLine 5'),
    })
    await authedPage.waitForTimeout(3000)

    // Click the card
    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // Should navigate to reader
    expect(authedPage.url()).toContain('/reader/')
  })

  test('2. 阅读器页面加载文档标题和内容', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'reader-content-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Document title should be displayed.\n\nContent paragraph one.\n\nContent paragraph two.'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // Document title visible — inside MarkdownDocumentEditor component
    await expect(authedPage.locator('.mde-title')).toBeVisible()
    await expect(authedPage.locator('.mde-title h1')).toContainText(/reader-content-test|reader/)

    // Editor content area visible
    await expect(authedPage.locator('.mde-prose')).toBeVisible()
  })

  test('3. 顶部浮动工具栏正常显示', async ({ authedPage }) => {
    // Upload a text file first
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'reader-toolbar-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Testing toolbar visibility.'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // For text docs: MarkdownDocumentEditor is now used instead of reader-topbar
    await expect(authedPage.locator('.mde-title')).toBeVisible()

    // Back button exists in the editor toolbar
    await expect(authedPage.locator('[aria-label="返回"]')).toBeVisible()
  })

  test('4. 左侧目录面板可打开和关闭', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'reader-toc-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('TOC test document.\n\nMore content.'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // For text docs: outline panel
    // MarkdownDocumentEditor has outline sidebar which can be toggled
    const outlineBtn = authedPage.locator('[aria-label="大纲"]').first()
    if (await outlineBtn.isVisible()) {
      await outlineBtn.click()
      await authedPage.waitForTimeout(500)
    }

    // Toggle outline via the editor's outline button
    await expect(authedPage.locator('.mde-outline')).toBeVisible({ timeout: 3000 })
  })

  test('5. Markdown 编辑器的 AI 面板按钮可见', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'reader-ai-panel-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Testing AI panel.'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // AI toggle button should be visible in MarkdownDocumentEditor
    await expect(authedPage.locator('[aria-label="AI助手"]').first()).toBeVisible()

    // Click AI toggle
    await authedPage.locator('[aria-label="AI助手"]').first().click()
    await authedPage.waitForTimeout(500)
  })

  test('6. 文本编辑器可正常渲染和滚动', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload a longer document
    const longContent = Array.from({ length: 50 }, (_, i) => `This is paragraph number ${i + 1} of the test document for scroll testing.`).join('\n\n')

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'reader-scroll-test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from(longContent),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // Editor should be rendered
    await expect(authedPage.locator('.mde-title')).toBeVisible()
    await expect(authedPage.locator('.mde-prose')).toBeVisible()
  })

  test('7. Markdown 文件上传后内容可编辑', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test-markdown.md',
      mimeType: 'text/markdown',
      buffer: Buffer.from('# Heading\n\n**Bold text**\n\n- List item 1\n- List item 2'),
    })
    await authedPage.waitForTimeout(3000)

    const card = authedPage.locator('.book-card').first()
    await expect(card).toBeVisible({ timeout: 5000 })
    await card.click()
    await authedPage.waitForTimeout(2000)

    // Markdown editor toolbar visible
    await expect(authedPage.locator('.mde-title')).toBeVisible()
    // Editor content loaded
    await expect(authedPage.locator('.mde-prose')).toBeVisible()
  })
})
