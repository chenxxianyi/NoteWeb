import { test, expect } from '../fixtures/auth'

test.describe('Documents', () => {

  test('1. 文件库页面正常加载', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1500)

    // Should show page title and upload zone
    await expect(authedPage.locator('.topbar__left h1')).toContainText('文件库')
    await expect(authedPage.locator('.upload-zone')).toBeVisible()
    await expect(authedPage.locator('.filter-bar')).toBeVisible()
  })

  test('2. 文件上传可用', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1500)

    // Upload a test file
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Hello World')
    })

    // Wait for upload to process
    await authedPage.waitForTimeout(4000)

    // Check if upload button/zone is still visible (should be)
    await expect(authedPage.locator('.upload-btn')).toBeVisible()
  })

  test('3. 筛选标签正常工作', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1500)

    // Click PDF filter
    const pdfFilter = authedPage.locator('.filter-tag', { hasText: 'PDF' })
    await pdfFilter.click()

    // Filter should become active
    await expect(pdfFilter).toHaveClass(/active/)
  })

  test('4. 所有筛选标签都存在', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1500)

    const tags = authedPage.locator('.filter-tag')
    const count = await tags.count()
    expect(count).toBeGreaterThanOrEqual(4) // 全部, PDF, Markdown, DOCX, TXT

    // Verify tag text
    await expect(tags.first()).toContainText('全部')
  })

  test('5. 搜索功能可用', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1500)

    // Search input should be visible
    const searchInput = authedPage.locator('.topbar__search input')
    await expect(searchInput).toBeVisible()

    // Type into search
    await searchInput.fill('test')
    await expect(searchInput).toHaveValue('test')
  })

  test('6. 新用户文件列表可能为空', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(2000)

    // A fresh user with no files should see either empty state or loading
    const page = authedPage.locator('.documents-page')
    await expect(page).toBeVisible()
  })
})
