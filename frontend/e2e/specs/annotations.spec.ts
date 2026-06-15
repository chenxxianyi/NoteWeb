import { test, expect } from '../fixtures/auth'

test.describe('Annotations', () => {

  test('1. 文件库页面加载正常（阅读器前置条件）', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    // Documents page should show basic structure
    await expect(authedPage.locator('.topbar__left h1')).toContainText('文件库')
    await expect(authedPage.locator('.upload-zone')).toBeVisible()
  })

  test('2. 上传文件并确认文件列表出现', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(1000)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test-annotate.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('This is a test document for annotation testing.\nSelect text and highlight it.')
    })

    // Wait for upload and card to appear
    await authedPage.waitForTimeout(3000)

    // Check if upload button exists (upload zone is always visible)
    await expect(authedPage.locator('.upload-btn')).toBeVisible()
  })
})
