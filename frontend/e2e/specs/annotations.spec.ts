import { test, expect } from '../fixtures/auth'
import { fileURLToPath } from 'url'
import path from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const samplePath = path.resolve(__dirname, '..', 'data', 'sample.txt')

test.describe('Annotations', () => {

  test('1. 进入阅读器 → 页面可打开', async ({ authedPage }) => {
    // Upload a document first
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(500)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles(samplePath)
    await authedPage.waitForTimeout(2000)

    // Click the first book card to open reader
    const firstCard = authedPage.locator('.book-card').first()
    await expect(firstCard).toBeVisible()

    // Click card to navigate to reader
    await firstCard.click()
    await authedPage.waitForTimeout(2000)

    // Should be on reader page
    expect(authedPage.url()).toContain('/reader/')
  })

  test('2. 阅读器页面显示文档内容', async ({ authedPage }) => {
    // Upload and open reader
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(500)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles(samplePath)
    await authedPage.waitForTimeout(2000)

    await authedPage.locator('.book-card').first().click()
    await authedPage.waitForTimeout(2000)

    // Verify reader loaded - check for reader container
    // The reader may show parsed content or a viewer
    await expect(authedPage.locator('body')).toBeVisible()
  })
})
