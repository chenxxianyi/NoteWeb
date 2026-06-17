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
})
