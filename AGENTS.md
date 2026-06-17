# Project memory

## Notes

- NoteWeb / 该项目中的.codex中的skill… ## User 该项目中的.codex中的skill我也想在reasonix中使用，你帮我拷贝一下deepseek也可以调用。

## Development preferences

- 以后开发默认不要使用 Playwright 进行测试、截图检查或浏览器自动化验证。
- 只有当用户明确要求“用 Playwright 测试 / 截图 / 浏览器验证”时，才运行 Playwright。
- 以后每次更新代码后，默认不要执行 E2E 测试；只有用户明确要求“跑 e2e / 执行端到端测试 / 用 Playwright 验证”时才运行。
- 需要验证时，优先使用类型检查、构建、单元测试、后端测试或代码审查；如果跳过 Playwright，需要在最终回复里说明。
