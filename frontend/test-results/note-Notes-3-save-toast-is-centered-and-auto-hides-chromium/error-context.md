# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: note.spec.ts >> Notes >> 3. save toast is centered and auto hides
- Location: e2e\specs\note.spec.ts:23:3

# Error details

```
Test timeout of 60000ms exceeded.
```

```
Error: locator.click: Target page, context or browser has been closed
Call log:
  - waiting for locator('.ne-footer__btn')

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - navigation [ref=e4]:
    - link "N" [ref=e5] [cursor=pointer]:
      - /url: "#/"
    - generic [ref=e6]:
      - link "仪表盘" [ref=e7] [cursor=pointer]:
        - /url: "#"
        - img [ref=e8]
        - generic: 仪表盘
      - link "文件库" [ref=e13] [cursor=pointer]:
        - /url: "#"
        - img [ref=e14]
        - generic: 文件库
      - link "笔记" [ref=e17] [cursor=pointer]:
        - /url: "#"
        - img [ref=e18]
        - generic: 笔记
      - link "设置" [ref=e21] [cursor=pointer]:
        - /url: "#"
        - img [ref=e22]
        - generic: 设置
    - generic [ref=e25] [cursor=pointer]: U
  - main [ref=e26]:
    - generic [ref=e27]:
      - generic [ref=e28]:
        - generic [ref=e29]:
          - heading "笔记" [level=2] [ref=e30]
          - button "新建笔记" [ref=e31] [cursor=pointer]:
            - img [ref=e32]
        - textbox "搜索笔记..." [ref=e34]
        - generic [ref=e35]:
          - button "全部" [ref=e36] [cursor=pointer]
          - button "阅读笔记" [ref=e37] [cursor=pointer]
          - button "技术" [ref=e38] [cursor=pointer]
          - button "随笔" [ref=e39] [cursor=pointer]
        - generic [ref=e40]: 暂无笔记
      - paragraph [ref=e42]: 选择一篇笔记开始编辑，或点击左上角新建笔记
```

# Test source

```ts
  1  | ﻿import { test, expect } from '../fixtures/auth'
  2  | 
  3  | test.describe('Notes', () => {
  4  | 
  5  |   test('1. notes list loads', async ({ authedPage }) => {
  6  |     await authedPage.goto('/notes')
  7  |     await authedPage.waitForTimeout(1500)
  8  | 
  9  |     const empty = authedPage.locator('.nl-empty')
  10 |     const items = authedPage.locator('.nl-items .nl-item')
  11 | 
  12 |     await expect(empty.or(items).first()).toBeVisible()
  13 |   })
  14 | 
  15 |   test('2. notes page layout is visible', async ({ authedPage }) => {
  16 |     await authedPage.goto('/notes')
  17 |     await authedPage.waitForTimeout(1500)
  18 | 
  19 |     await expect(authedPage.locator('.note-list')).toBeVisible()
  20 |     await expect(authedPage.locator('.note-editor')).toBeVisible()
  21 |   })
  22 | 
  23 |   test('3. save toast is centered and auto hides', async ({ authedPage }) => {
  24 |     await authedPage.goto('/notes')
  25 |     await authedPage.waitForTimeout(1500)
  26 | 
> 27 |     await authedPage.locator('.ne-footer__btn').click()
     |                                                 ^ Error: locator.click: Target page, context or browser has been closed
  28 | 
  29 |     const toast = authedPage.locator('.note-toast')
  30 |     await expect(toast).toBeVisible({ timeout: 5000 })
  31 | 
  32 |     const box = await toast.boundingBox()
  33 |     expect(box).not.toBeNull()
  34 |     if (!box) return
  35 | 
  36 |     const viewport = authedPage.viewportSize()
  37 |     expect(viewport).not.toBeNull()
  38 |     if (!viewport) return
  39 | 
  40 |     expect(Math.abs((box.x + box.width / 2) - viewport.width / 2)).toBeLessThan(80)
  41 |     expect(Math.abs((box.y + box.height / 2) - viewport.height / 2)).toBeLessThan(100)
  42 | 
  43 |     await expect(toast).toBeHidden({ timeout: 5000 })
  44 |   })
  45 | })
  46 | 
```