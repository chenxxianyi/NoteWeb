import { test, expect } from '../fixtures/auth'

test.describe('AI Panel', () => {

  test('1. 文件库页面加载正常', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    await expect(authedPage.locator('.topbar__left h1')).toContainText('文件库')
    await expect(authedPage.locator('.filter-bar')).toBeVisible()
  })

  test('2. 上传文件功能可用', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Upload a file
    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test-ai.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('Sample document for AI testing.')
    })

    await authedPage.waitForTimeout(3000)
    await expect(authedPage.locator('.upload-zone')).toBeVisible()
  })
})
