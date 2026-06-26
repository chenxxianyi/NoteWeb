# NoteWeb 前端测试方案

本文档基于 `.codex/skills/webapp-testing` 中的 Playwright 测试能力整理，目标是为 NoteWeb 前端建立一套可执行、可维护、可逐步扩展的测试方案。

## 测试原则

- 默认优先使用 `build`、类型检查、单元测试做快速验证。
- 只有在明确需要浏览器级验证时才运行 Playwright E2E 测试。
- E2E 测试不调用真实 DeepSeek 或第三方 AI 服务，AI 场景默认使用网络拦截或测试后端模拟。
- 避免硬编码等待时间，优先使用 Playwright 的自动等待和 `expect` 断言。
- 优先使用语义选择器、可访问名称和 `data-testid`，避免依赖脆弱的 DOM 层级。
- 截图和 trace 默认用于失败排查，不作为每次测试的强制输出。

## 当前项目基础

前端已具备以下测试脚本：

```bash
npm.cmd run build
npm.cmd run test:unit
npm.cmd run test:e2e
npm.cmd run test:e2e:headed
```

现有 E2E 配置：

- 配置文件：`frontend/e2e/playwright.config.ts`
- 测试目录：`frontend/e2e/specs`
- 默认浏览器：Chromium
- 默认地址：`http://127.0.0.1:5174`
- 失败截图：`screenshot: 'only-on-failure'`
- 重试 trace：`trace: 'on-first-retry'`

现有 E2E 用例文件：

- `auth.spec.ts`
- `documents.spec.ts`
- `note.spec.ts`
- `annotations.spec.ts`
- `ai.spec.ts`

## 测试分层

### 1. 快速验证

适合每次前端代码更新后运行：

```bash
npm.cmd run build
npm.cmd run test:unit
```

覆盖目标：

- Vue/TypeScript 编译错误
- Vite 打包错误
- PDF 批注几何、历史、绘图类型等纯逻辑回归

### 2. 关键 E2E 验证

适合改动登录、文档、阅读器、AI、PDF 批注等核心流程后运行：

```bash
npm.cmd run test:e2e
```

覆盖目标：

- 页面路由和登录状态
- 用户真实操作路径
- 表单、上传、编辑、阅读、批注、AI 面板等交互
- 网络失败、权限失败、空数据等异常状态

### 3. 调试型浏览器验证

适合排查 UI 问题、布局重叠、流式输出肉眼效果时运行：

```bash
npm.cmd run test:e2e:headed
```

覆盖目标：

- 观察浏览器真实渲染
- 检查移动端布局
- 调试 Playwright selector
- 复现用户截图中的 UI 问题

## 优先测试范围

### P0：核心阻断流程

1. 登录与认证
   - 正确账号密码可登录。
   - 错误账号密码显示错误提示。
   - 未登录访问受保护页面会跳转登录页。
   - 刷新页面后认证状态保持正确。

2. 文档列表与打开
   - 文档列表可加载。
   - 空列表、加载失败有明确提示。
   - 上传文本类文件后可出现在列表中。
   - 点击文档后进入阅读器页面。

3. Reader 顶栏一致性
   - PDF、DOCX、Markdown 都只显示一条主功能导航栏。
   - PDF 顶栏采用与文本类文档一致的三段式布局：标题区、功能区、侧边动作区。
   - AI、搜索、保存状态、更多菜单入口不重叠。
   - 移动端宽度下主要入口可用。

4. AI 配置与调用
   - 设置页可保存 DeepSeek 配置。
   - API Key 不以明文回显。
   - 未配置 API Key 时调用 AI 显示失败提示。
   - AI 调用失败时不返回 mock 内容。
   - AI SSE 响应可被前端正确解析并展示。

5. PDF 阅读与批注
   - PDF 页面可渲染。
   - 画笔、高亮、文本、形状、橡皮擦工具可切换 active 状态。
   - 页码跳转、上一页、下一页可用。
   - 缩放菜单可打开，缩放状态正确更新。
   - 批注列表侧栏可打开并显示批注。

### P1：高价值功能

1. Markdown 编辑器
   - 加粗、斜体、标题、列表、引用、链接、图片入口可用。
   - 保存状态从“保存中”变为“已保存”。
   - 搜索面板可打开并定位匹配项。
   - 大纲侧栏可开关。

2. AI 流式输出体验
   - 收到多段 SSE delta 后，最终内容完整。
   - 前端展示队列不会吞字、乱序或重复。
   - AI 面板滚动位置跟随新内容。

3. 响应式布局
   - Desktop Chrome：主工作流完整可用。
   - Mobile viewport：顶栏不溢出，AI 面板和更多菜单可操作。
   - 窄屏下 PDF 工具低频入口合理折叠。

### P2：增强验证

1. 视觉回归
   - Reader 顶栏截图。
   - 设置页 AI 配置卡片截图。
   - PDF 工具栏移动端截图。

2. 网络与异常
   - API 500、401、404 时 UI 有可理解提示。
   - AI SSE 中途返回 error event 时显示失败。
   - 上传失败、导出失败、分享失败状态正确。

3. 多浏览器
   - Chromium 作为默认。
   - 稳定后可扩展 Firefox。
   - 如需要验证 Safari 兼容，可增加 WebKit 项目。

## 建议新增或强化的 E2E 用例

### `reader-toolbar.spec.ts`

目标：验证 PDF 与文本类文档顶栏格式统一。

建议用例：

- 打开 PDF 文档，断言页面只有一个 `reader-topbar`。
- 打开文本类文档，断言编辑器只有一个 `mde-topbar`。
- PDF 顶栏包含标题、PDF 阅读工具、保存状态、AI、搜索、更多入口。
- PDF 顶栏不存在旧版居中浮动胶囊样式。
- 在移动端 viewport 下，PDF 顶栏不出现水平溢出。

### `ai-stream.spec.ts`

目标：验证 AI SSE 响应的前端解析和展示。

建议用例：

- 拦截 `/api/v1/ai/chat/stream`，返回 `delta` 和 `done` 事件。
- 用户提问后，AI 面板出现助手消息。
- 最终消息内容等于所有 delta 的拼接结果。
- 拦截返回 `error` event 时，页面显示失败提示。
- 确认失败时不会出现 mock 文案。

注意：普通 `route.fulfill` 往往会一次性返回完整 body，适合验证 SSE 解析和最终内容；如果要验证真实逐段渲染，需要使用测试后端或本地 mock server 分块写入响应。

### `settings-ai.spec.ts`

目标：验证 AI 配置页面。

建议用例：

- 选择 DeepSeek。
- 输入 API Key 并保存。
- 保存成功后显示 `has_api_key` 状态。
- API Key 输入框不显示明文已保存 key。
- 后端返回失败时显示错误提示。

### `pdf-tools.spec.ts`

目标：验证 PDF 阅读和批注工具。

建议用例：

- 打开 PDF 文档后页面渲染 canvas。
- 点击画笔、高亮、文本、形状、橡皮擦，断言 active 状态变化。
- 打开样式面板，选择颜色和线宽。
- 点击页码，输入目标页码并跳转。
- 打开缩放菜单，选择 100%、适配宽度、适配页面。
- 打开批注侧栏。

### `markdown-editor.spec.ts`

目标：验证文本类文档编辑体验。

建议用例：

- 打开 Markdown/DOCX 文档后显示编辑器顶栏。
- 输入内容后保存状态变为 dirty/saving/saved。
- 点击加粗、斜体、标题、列表等工具按钮。
- 打开搜索面板并定位关键词。
- 开关大纲侧栏。
- 点击 AI 助手按钮后打开统一 AI 面板。

## Mock 与测试数据策略

### API Mock

适合使用 Playwright `page.route` 拦截：

- 登录接口
- 文档列表接口
- 文档详情接口
- AI 配置接口
- AI SSE 接口
- 批注列表与保存接口

推荐原则：

- E2E 默认不依赖真实 DeepSeek。
- AI Key 使用测试字符串，不写入真实密钥。
- SSE 测试使用固定 delta，保证断言稳定。
- 网络错误用 mock 返回 500 或 SSE error event。

### 文件数据

现有测试数据：

- `frontend/e2e/data/sample.txt`

建议补充：

- `sample.md`：用于 Markdown 编辑器。
- `sample.pdf`：用于 PDF 渲染和批注。
- `sample.docx`：用于文档类型识别和 Reader 顶栏一致性。

## Selector 规范

优先级从高到低：

1. `data-testid`
2. `getByRole`
3. `getByLabel`
4. `getByText`
5. 稳定 class
6. CSS 层级选择器

建议逐步给关键组件补 `data-testid`：

- `reader-topbar`
- `markdown-topbar`
- `ai-panel`
- `ai-input`
- `ai-send-button`
- `pdf-tool-pen`
- `pdf-tool-highlighter`
- `pdf-tool-eraser`
- `pdf-page-label`
- `pdf-zoom-menu`
- `settings-ai-provider`
- `settings-ai-save`

## 执行建议

### 日常开发

```bash
npm.cmd run build
npm.cmd run test:unit
```

### 改动核心交互后

```bash
npm.cmd run test:e2e
```

### 调试某个测试文件

```bash
npx playwright test e2e/specs/ai.spec.ts --config e2e/playwright.config.ts
```

### 调试浏览器可视化行为

```bash
npm.cmd run test:e2e:headed
```

## CI 建议

第一阶段：

- 只跑 `npm.cmd run build`
- 只跑 `npm.cmd run test:unit`

第二阶段：

- 对主分支或 PR 跑 Chromium E2E。
- 保留失败截图和 trace。

第三阶段：

- 增加移动端 viewport。
- 增加 Firefox 或 WebKit。
- 对 Reader 顶栏、AI 设置、PDF 工具栏做有限视觉回归。

## 维护规则

- 每个测试只验证一个核心行为，避免一个用例覆盖过多流程。
- 测试之间不共享脏状态。
- 登录态通过 fixture 管理。
- 上传、批注、AI 配置等测试数据在测试结束后清理。
- 新增 UI 关键交互时，同步补充 `data-testid` 和对应 E2E。
- 对易变文案少做强绑定，优先断言状态、角色、按钮可用性和关键结果。

## 推荐实施顺序

1. 补 `reader-toolbar.spec.ts`，验证 PDF 与文本类文档顶栏统一。
2. 补 `ai-stream.spec.ts`，验证 SSE 解析、失败提示和无 mock 回答。
3. 补 `settings-ai.spec.ts`，验证 DeepSeek 配置保存流程。
4. 补 `pdf-tools.spec.ts`，覆盖 PDF 批注和页码缩放。
5. 补 `markdown-editor.spec.ts`，覆盖文本编辑器常用操作。
6. 增加移动端 project，并筛选 P0 用例在移动端执行。

