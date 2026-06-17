import { test, expect } from '../fixtures/auth'

test.describe('Notes', () => {

  test('1. notes list loads', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    const empty = authedPage.locator('.nl-empty')
    const items = authedPage.locator('.nl-items .nl-item')

    await expect(empty.or(items).first()).toBeVisible()
  })

  test('2. notes page layout is visible', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await expect(authedPage.locator('.note-list')).toBeVisible()
    await expect(authedPage.locator('.note-editor')).toBeVisible()
  })

  test('3. save toast is centered and auto hides', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.locator('.ne-footer__btn').click()

    const toast = authedPage.locator('.note-toast')
    await expect(toast).toBeVisible({ timeout: 5000 })

    const box = await toast.boundingBox()
    expect(box).not.toBeNull()
    if (!box) return

    const viewport = authedPage.viewportSize()
    expect(viewport).not.toBeNull()
    if (!viewport) return

    expect(Math.abs((box.x + box.width / 2) - viewport.width / 2)).toBeLessThan(80)
    expect(Math.abs((box.y + box.height / 2) - viewport.height / 2)).toBeLessThan(100)

    await expect(toast).toBeHidden({ timeout: 5000 })
  })

  test('4. editor context menu opens and closes', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.locator('.nl-header__btn').click()
    const editorBody = authedPage.locator('.ne-body .ProseMirror')
    await expect(editorBody).toBeVisible({ timeout: 5000 })
    await editorBody.click({ button: 'right' })

    const menu = authedPage.locator('.note-context-menu')
    await expect(menu).toBeVisible()
    const box = await menu.boundingBox()
    expect(box).not.toBeNull()
    expect(box!.width).toBeLessThan(270)
    await expect(menu.locator('.ncm-row', { hasText: '复制 / 粘贴为...' })).toBeVisible()
    await expect(menu.getByRole('button', { name: '加粗' })).toBeVisible()
    await expect(menu.getByRole('button', { name: '插入链接' })).toBeVisible()

    await authedPage.locator('.note-list').click()
    await expect(menu).toBeHidden()
  })

  test('5. context menu paste inserts clipboard text when allowed', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.evaluate(() => {
      Object.defineProperty(navigator, 'clipboard', {
        configurable: true,
        value: {
          readText: async () => '菜单粘贴文本',
        },
      })
    })

    await authedPage.locator('.nl-header__btn').click()
    const editorBody = authedPage.locator('.ne-body .ProseMirror')
    await expect(editorBody).toBeVisible({ timeout: 5000 })
    await editorBody.click({ button: 'right' })

    await authedPage.getByRole('button', { name: '粘贴' }).click()

    await expect(editorBody).toContainText('菜单粘贴文本')
  })

  test('6. paragraph submenu shows block actions', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.locator('.nl-header__btn').click()
    const editorBody = authedPage.locator('.ne-body .ProseMirror')
    await expect(editorBody).toBeVisible({ timeout: 5000 })
    await editorBody.click({ button: 'right' })

    const menu = authedPage.locator('.note-context-menu')
    await expect(menu).toBeVisible()

    await menu.locator('.ncm-row', { hasText: '段落' }).hover()

    const flyout = menu.locator('.ncm-flyout[data-submenu="paragraph"]')
    await expect(flyout).toBeVisible()
    await expect(flyout.getByRole('button', { name: /一级标题/ })).toBeVisible()
    await expect(flyout.getByRole('button', { name: /三级标题/ })).toBeVisible()
    await expect(flyout.getByRole('button', { name: /六级标题/ })).toBeVisible()
    await expect(flyout.getByRole('button', { name: /提升标题级别/ })).toBeVisible()
    await expect(flyout.getByRole('button', { name: /表格/ })).toBeVisible()
    await expect(flyout.getByRole('button', { name: /任务列表/ })).toBeVisible()
    await expect(flyout.getByRole('button', { name: /在上方插入段落/ })).toBeVisible()
  })

  test('7. paragraph submenu can apply heading three', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.locator('.nl-header__btn').click()
    const editorBody = authedPage.locator('.ne-body .ProseMirror')
    await expect(editorBody).toBeVisible({ timeout: 5000 })
    await editorBody.click()
    await editorBody.type('三级标题测试')
    await editorBody.click({ button: 'right' })

    const menu = authedPage.locator('.note-context-menu')
    await expect(menu).toBeVisible()

    await menu.locator('.ncm-row', { hasText: '段落' }).hover()
    await menu.locator('.ncm-flyout[data-submenu="paragraph"] .ncm-flyout__item', { hasText: '三级标题' }).click()

    await expect(editorBody.locator('h3')).toHaveText('三级标题测试')
  })

  test('8. paragraph submenu inserts advanced blocks', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.locator('.nl-header__btn').click()
    const editorBody = authedPage.locator('.ne-body .ProseMirror')
    await expect(editorBody).toBeVisible({ timeout: 5000 })

    async function runBlockCommand(name: string) {
      await editorBody.click({ button: 'right' })
      const menu = authedPage.locator('.note-context-menu')
      await expect(menu).toBeVisible()
      await menu.locator('.ncm-row', { hasText: '段落' }).hover()
      await menu.locator('.ncm-flyout[data-submenu="paragraph"] .ncm-flyout__item', { hasText: name }).click()
    }

    await runBlockCommand('表格')
    await expect(editorBody.locator('table')).toBeVisible()

    await runBlockCommand('公式块')
    await expect(editorBody.locator('[data-type="formula-block"]')).toBeVisible()

    await runBlockCommand('警告框')
    await expect(editorBody.locator('[data-type="callout-block"]')).toBeVisible()

    await runBlockCommand('代码工具')
    await expect(editorBody.locator('pre')).toBeVisible()
  })

  test('9. task state submenu toggles task item', async ({ authedPage }) => {
    await authedPage.goto('/notes')
    await authedPage.waitForTimeout(1500)

    await authedPage.locator('.nl-header__btn').click()
    const editorBody = authedPage.locator('.ne-body .ProseMirror')
    await expect(editorBody).toBeVisible({ timeout: 5000 })
    await editorBody.click()
    await editorBody.type('任务一')
    await authedPage.keyboard.press('Enter')
    await authedPage.keyboard.press('Space')

    await editorBody.click({ button: 'right' })
    const menu = authedPage.locator('.note-context-menu')
    await expect(menu).toBeVisible()
    await menu.locator('.ncm-row', { hasText: '段落' }).hover()
    const taskStateItem = menu.locator('.ncm-flyout[data-submenu="paragraph"] .ncm-flyout__item', { hasText: '任务状态' })
    await taskStateItem.scrollIntoViewIfNeeded()
    await taskStateItem.click()

    await expect(editorBody.locator('li input[type="checkbox"]')).toHaveCount(1)
    await expect(editorBody.locator('li input[type="checkbox"]')).toBeChecked()
  })
})
