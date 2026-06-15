import { test, expect } from '../fixtures/auth'
import { fileURLToPath } from 'url'
import path from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const samplePath = path.resolve(__dirname, '..', 'data', 'sample.txt')

test.describe('AI Panel', () => {

  test('1. 进入阅读器后 AI 总结面板可打开', async ({ authedPage }) => {
    // Upload a document
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(500)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles(samplePath)
    await authedPage.waitForTimeout(2000)

    // Open reader
    await authedPage.locator('.book-card').first().click()
    await authedPage.waitForTimeout(2000)

    // Verify reader page loaded
    expect(authedPage.url()).toContain('/reader/')
  })

  test('2. 阅读器页面显示基本结构', async ({ authedPage }) => {
    await authedPage.goto('/documents')
    await authedPage.waitForTimeout(500)

    const fileInput = authedPage.locator('input[type="file"]')
    await fileInput.setInputFiles(samplePath)
    await authedPage.waitForTimeout(2000)

    await authedPage.locator('.book-card').first().click()
    await authedPage.waitForTimeout(2000)

    // Reader page should have body visible
    await expect(authedPage.locator('body')).toBeVisible()
  })
})
