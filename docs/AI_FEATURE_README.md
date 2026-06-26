# AI 文档助手功能

## 功能概述

NoteWeb 现已集成 AI 文档助手功能,为用户提供智能化的文档阅读和分析体验。

### 核心功能

1. **智能问答**: 基于文档内容进行问答对话
2. **文档总结**: 自动生成文档核心内容总结
3. **联网搜索**: 实时搜索互联网信息并整合回答

## 后端配置

### 环境变量

在 `backend-go/.env` 文件中配置以下环境变量:

```env
# DeepSeek API配置(可选,用户可在前端设置页面配置)
DEEPSEEK_API_KEY=your_deepseek_api_key
DEEPSEEK_BASE_URL=https://api.deepseek.com
```

### 数据库迁移

新功能会自动创建以下数据表:
- `ai_conversations`: 存储用户的AI对话历史
- `ai_provider_configs`: 存储用户的AI服务商配置

后端服务启动时会自动执行数据库迁移。

## 前端使用

### 配置 AI 服务

1. 登录后进入"设置"页面
2. 找到"AI 设置"部分
3. 选择 AI 服务商 (DeepSeek / OpenAI / Mock)
4. 输入 API Key
5. 配置 API Base URL (可选)
6. 选择模型
7. 点击"保存"

### 使用 AI 助手

#### PDF 文档阅读

1. 打开任意 PDF 文档
2. 点击右侧工具栏的"批注和AI"按钮
3. 切换到"AI 助手"标签页
4. 即可开始使用 AI 功能

#### 功能操作

- **智能问答**: 在输入框中输入问题,按 Enter 或点击发送按钮
- **文档总结**: 点击"生成总结"按钮
- **联网搜索**: 点击"联网搜索"按钮,然后输入搜索关键词

### 支持的文档类型

- PDF 文档
- Markdown 文档
- DOCX 文档
- TXT 文档

## API 端点

### AI 相关端点

所有 AI 端点都需要认证 (Bearer Token)。

#### POST `/api/v1/ai/chat`
发送聊天消息

**请求体**:
```json
{
  "document_id": 123,
  "question": "这个文档主要讲了什么?",
  "conversation_type": "chat"
}
```

**响应**:
```json
{
  "answer": "这个文档主要介绍了...",
  "conversation_id": 456
}
```

#### GET `/api/v1/ai/documents/:id/summary`
获取文档总结

**响应**:
```json
{
  "summary": "文档总结内容...",
  "conversation_id": 789
}
```

#### POST `/api/v1/ai/search`
联网搜索

**请求体**:
```json
{
  "query": "Vue 3 组合式 API 最佳实践"
}
```

**响应**:
```json
{
  "answer": "搜索结果...",
  "sources": [],
  "conversation_id": 101
}
```

#### GET `/api/v1/ai/conversations`
获取对话历史

**查询参数**:
- `document_id` (可选): 文档ID

**响应**: 对话列表

#### DELETE `/api/v1/ai/conversations/:id`
删除对话

#### GET `/api/v1/ai/provider-config`
获取当前用户的 AI 配置

#### PUT `/api/v1/ai/provider-config`
更新 AI 配置

**请求体**:
```json
{
  "provider": "deepseek",
  "api_key": "sk-xxx",
  "base_url": "https://api.deepseek.com",
  "model": "deepseek-chat"
}
```

## 技术实现

### 后端技术栈

- **语言**: Go
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **AI API**: DeepSeek / OpenAI

### 前端技术栈

- **框架**: Vue 3 (Composition API)
- **状态管理**: Pinia
- **HTTP客户端**: Axios
- **UI组件**: Lucide Icons

### 架构设计

#### 对话历史存储

对话历史以 JSON 格式存储在 `ai_conversations` 表的 `messages` 字段中,每条消息包含:
- `role`: 消息角色 (user / assistant / system)
- `content`: 消息内容

#### 文档上下文构建

当用户针对文档提问时,系统会:
1. 获取文档的 `parsed_content` (解析后的文本)
2. 截取前 4000 tokens 作为上下文
3. 构建系统提示词
4. 调用 AI API

#### API 密钥安全

- 用户的 API Key 存储在数据库中
- 前端请求时不返回 API Key
- API Key 仅在调用 AI API 时使用

## 注意事项

### Token 限制

- 文档上下文限制: 4000 tokens (聊天) / 6000 tokens (总结)
- 超长文档会被截断,可能影响 AI 回答质量

### API 费用

- DeepSeek 和 OpenAI 的 API 调用会产生费用
- 费用由用户的 API Key 承担
- Mock 模式不产生实际 API 调用

### 错误处理

- API 调用失败时会显示友好的错误提示
- 支持 fallback 到 Mock 模式
- 网络超时时间: 60 秒

## 未来优化方向

1. **流式响应**: 支持 SSE 实现打字机效果
2. **上下文检索优化**: 使用向量数据库进行语义检索
3. **多轮对话优化**: 实现对话摘要避免 token 过多
4. **批量文档分析**: 支持跨文档问答
5. **导出对话**: 支持导出为 Markdown/PDF
6. **对话模板**: 预设常用的对话模板
7. **多语言支持**: 支持多语言的 AI 交互

## 开发指南

### 本地开发

1. 启动后端服务:
```bash
cd backend-go
go run cmd/server/main.go
```

2. 启动前端开发服务器:
```bash
cd frontend
npm run dev
```

3. 在浏览器中访问 http://localhost:5173

### 测试

后端测试:
```bash
cd backend-go
go test ./...
```

前端测试:
```bash
cd frontend
npm run test:unit
```

## 故障排查

### 常见问题

**Q: AI 不回复或报错**
A: 检查 API Key 是否正确配置,网络是否可访问 AI API

**Q: 文档总结不准确**
A: 可能文档内容过长被截断,建议分段提问

**Q: 对话历史丢失**
A: 检查数据库连接是否正常,对话历史存储在数据库中

### 日志查看

后端日志会输出到控制台,包含 API 调用的详细信息。

---

祝您使用愉快！如有问题或建议,欢迎反馈。
