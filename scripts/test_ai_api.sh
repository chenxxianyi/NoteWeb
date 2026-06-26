#!/bin/bash

# AI 功能测试脚本
# 使用方法: ./test_ai_api.sh <base_url> <token>

BASE_URL="${1:-http://localhost:8020/api/v1}"
TOKEN="$2"

if [ -z "$TOKEN" ]; then
  echo "错误: 请提供认证token"
  echo "使用方法: ./test_ai_api.sh <base_url> <token>"
  exit 1
fi

echo "=== AI 功能API测试 ==="
echo "Base URL: $BASE_URL"
echo ""

# 测试1: 获取AI配置
echo "1. 获取AI配置..."
curl -s -X GET "$BASE_URL/ai/provider-config" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq .
echo ""

# 测试2: 更新AI配置
echo "2. 更新AI配置 (设置为mock模式)..."
curl -s -X PUT "$BASE_URL/ai/provider-config" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "mock",
    "model": "mock-model"
  }' | jq .
echo ""

# 测试3: 聊天 (需要document_id,这里使用1作为示例)
echo "3. 测试聊天..."
curl -s -X POST "$BASE_URL/ai/chat" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_id": 1,
    "question": "这个文档主要讲了什么?",
    "conversation_type": "chat"
  }' | jq .
echo ""

# 测试4: 联网搜索
echo "4. 测试联网搜索..."
curl -s -X POST "$BASE_URL/ai/search" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Vue 3 组合式API"
  }' | jq .
echo ""

# 测试5: 获取对话历史
echo "5. 获取对话历史..."
curl -s -X GET "$BASE_URL/ai/conversations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq .
echo ""

echo "=== 测试完成 ==="
