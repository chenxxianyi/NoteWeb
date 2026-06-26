import { test, expect } from '../fixtures/auth'

test.describe('Settings & AI Config', () => {

  test('1. 设置页面结构加载正常', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1500)

    await expect(authedPage.locator('.page-title')).toContainText('设置')
    // Sections
    await expect(authedPage.locator('.section__title', { hasText: '用户资料' })).toBeVisible()
    await expect(authedPage.locator('.section__title', { hasText: '外观' })).toBeVisible()
    await expect(authedPage.locator('.section__title', { hasText: '存储空间' })).toBeVisible()
    await expect(authedPage.locator('.section__title', { hasText: 'AI 配置' })).toBeVisible()
  })

  test('2. 用户资料卡片信息可见', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1500)

    // Avatar, username, email, password sections
    await expect(authedPage.locator('.setting-item__label', { hasText: '头像' })).toBeVisible()
    await expect(authedPage.locator('.setting-item__label', { hasText: '用户名' })).toBeVisible()
    await expect(authedPage.locator('.setting-item__label', { hasText: '邮箱' })).toBeVisible()
    await expect(authedPage.locator('.setting-item__label', { hasText: '密码' })).toBeVisible()
  })

  test('3. 外观设置可选主题和字体', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1500)

    // Theme dots exist
    const themeDots = authedPage.locator('.theme-dot')
    expect(await themeDots.count()).toBeGreaterThanOrEqual(3)

    // Font selector exists
    const fontSelect = authedPage.locator('.section__card select').first()
    await expect(fontSelect).toBeVisible()

    // Reading mode toggle exists
    await expect(authedPage.locator('.toggle')).toBeVisible()
  })

  test('4. AI 配置区域完整可用', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1500)

    // Provider selector
    const providerSelect = authedPage.locator('select').filter({ hasText: /DeepSeek|OpenAI/ }).first()
    await expect(providerSelect).toBeVisible()

    // Model selector
    const modelSelect = authedPage.locator('.model-select-wrapper select')
    await expect(modelSelect).toBeVisible()

    // API Key input
    const apiKeyInput = authedPage.locator('input[placeholder="sk-xxxxxxxxxxxxxxxx"]')
    await expect(apiKeyInput).toBeVisible()
    await expect(apiKeyInput).toHaveAttribute('type', 'password')

    // Save button
    const saveBtn = authedPage.getByRole('button', { name: '保存配置' })
    await expect(saveBtn).toBeVisible()
  })

  test('5. 切换到 OpenAI 显示 Base URL 输入框', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1500)

    // Switch to OpenAI
    const providerSelect = authedPage.locator('select').filter({ hasText: /DeepSeek|OpenAI/ }).first()
    await providerSelect.selectOption('OpenAI')

    // Base URL should now be visible
    const baseUrlInput = authedPage.locator('input[placeholder="https://api.openai.com/v1"]')
    await expect(baseUrlInput).toBeVisible()
  })

  test('6. 保存 AI 配置时后端返回失败显示错误提示', async ({ authedPage }) => {
    // Mock a failed AI config save response
    await authedPage.route('**/api/v1/ai/provider', async (route) => {
      if (route.request().method() === 'PATCH' || route.request().method() === 'PUT') {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({ detail: '配置保存失败' }),
        })
      } else {
        await route.continue()
      }
    })

    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1500)

    // Fill API key
    const apiKeyInput = authedPage.locator('input[placeholder="sk-xxxxxxxxxxxxxxxx"]')
    await apiKeyInput.fill('sk-test-fail-key')

    // Click save
    const saveBtn = authedPage.getByRole('button', { name: '保存配置' })
    await saveBtn.click()

    // Should show error notification
    await expect(authedPage.locator('.notification-toast')).toBeVisible({ timeout: 5000 })
  })

  test('7. 密码修改弹框可正常打开关闭', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1000)

    // Click "修改密码" button
    await authedPage.getByRole('button', { name: '修改密码' }).click()
    await authedPage.waitForTimeout(300)

    // Modal should be visible
    await expect(authedPage.locator('.modal__title', { hasText: '修改密码' })).toBeVisible()

    // Modal has 3 password inputs
    const inputs = authedPage.locator('.modal__content input[type="password"]')
    expect(await inputs.count()).toBe(3)

    // Close modal
    await authedPage.getByRole('button', { name: '取消' }).click()
    await expect(authedPage.locator('.modal__title', { hasText: '修改密码' })).toBeHidden()
  })

  test('8. 导出数据按钮可用', async ({ authedPage }) => {
    await authedPage.goto('/settings')
    await authedPage.waitForTimeout(1000)

    const exportBtn = authedPage.getByRole('button', { name: '导出数据' })
    await expect(exportBtn).toBeVisible()
    await expect(exportBtn).toBeEnabled()
  })
})
