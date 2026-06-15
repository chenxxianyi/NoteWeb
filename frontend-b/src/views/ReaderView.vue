<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { useAnnotationStore } from '../stores/annotationStore'

const route = useRoute()
const router = useRouter()
const documentStore = useDocumentStore()
const annotationStore = useAnnotationStore()

const docId = Number(route.params.documentId)
const activeTab = ref<'annotations' | 'ai'>('annotations')

onMounted(async () => {
  if (docId) {
    await documentStore.fetchDocument(docId)
    await annotationStore.fetchAnnotations(docId)
  }
})

function goBack() { router.push('/documents') }

// Toolbar action stubs
const zoomLevel = ref(100)
function zoomIn() { zoomLevel.value = Math.min(200, zoomLevel.value + 10) }
function zoomOut() { zoomLevel.value = Math.max(50, zoomLevel.value - 10) }
function toggleReadingMode() { /* future: toggle between scroll/paginated */ }
function showSearch() { /* future: open search bar */ }
</script>

<template>
  <div class="reader-page">
    <!-- Top Toolbar -->
    <div class="toolbar">
      <div class="toolbar__left">
        <button class="tlb" @click="goBack" title="返回">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
        </button>
        <div class="tlb-divider"></div>
        <button class="tlb" title="搜索" @click="showSearch"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></button>
        <button class="tlb" title="缩小" @click="zoomOut"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="8" y1="11" x2="14" y2="11"/></svg></button>
        <span class="tlb-label">{{ zoomLevel }}%</span>
        <button class="tlb" title="放大" @click="zoomIn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg></button>
        <div class="tlb-divider"></div>
        <button class="tlb" title="阅读模式" @click="toggleReadingMode"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg></button>
      </div>
      <div class="toolbar__center">{{ documentStore.currentDocument?.title || '阅读器' }}</div>
      <div class="toolbar__right">
        <button class="tlb-text" title="AI 总结">AI 总结</button>
        <div class="tlb-divider"></div>
        <button class="tlb" title="保存进度"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/></svg></button>
      </div>
    </div>

    <!-- Reader Body: 3 columns -->
    <div class="reader-body">
      <!-- Left: TOC -->
      <div class="left-panel">
        <div class="panel-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
          目录
        </div>
        <div class="panel-body">
          <div class="toc-item active">一、概述</div>
          <div class="toc-item l2">1.1 背景</div>
          <div class="toc-item l2">1.2 目标</div>
          <div class="toc-item">二、核心功能</div>
          <div class="toc-item l2">2.1 文档阅读</div>
          <div class="toc-item l3">PDF 支持</div>
          <div class="toc-item l3">Markdown 渲染</div>
          <div class="toc-item">三、技术架构</div>
        </div>
        <div class="left-footer">
          <div>共 {{ documentStore.currentDocument?.file_type?.toUpperCase() || '—' }} · {{ documentStore.currentDocument?.file_size ? `${(documentStore.currentDocument.file_size / 1024).toFixed(0)} KB` : '—' }}</div>
        </div>
      </div>

      <!-- Center: Content -->
      <div class="center-panel">
        <div class="page-indicator">第 1 页</div>
        <div class="doc-title">{{ documentStore.currentDocument?.title || '加载中...' }}</div>
        <div class="doc-meta">{{ documentStore.currentDocument?.file_type?.toUpperCase() || '—' }} · {{ documentStore.currentDocument?.id || '' }}</div>
        <div class="doc-body">
          <p>{{ documentStore.currentDocument?.parsed_content || '暂无内容' }}</p>
        </div>
      </div>

      <!-- Right: Annotations / AI -->
      <div class="right-panel">
        <div class="tab-bar">
          <div :class="['tab-item', { active: activeTab === 'annotations' }]" @click="activeTab = 'annotations'">批注</div>
          <div :class="['tab-item', { active: activeTab === 'ai' }]" @click="activeTab = 'ai'">AI</div>
        </div>
        <div class="right-body">
          <!-- Annotations tab -->
          <template v-if="activeTab === 'annotations'">
            <div v-if="annotationStore.annotations.length === 0" style="text-align:center;padding:2rem 0;color:var(--text-muted);font-size:0.85rem;">暂无批注<br>选中文本后可创建高亮</div>
            <div v-for="ann in annotationStore.annotations" :key="ann.id" class="anno-card">
              <div class="anno-card__text">{{ ann.selected_text }}</div>
              <div class="anno-card__meta">
                <span :class="['anno-badge', ann.type === 'highlight' ? 'badge-highlight' : 'badge-underline']">{{ ann.type === 'highlight' ? '高亮' : '下划线' }}</span>
                第{{ ann.page }}页
              </div>
            </div>
          </template>
          <!-- AI tab -->
          <template v-if="activeTab === 'ai'">
            <div class="ai-card">
              <div class="ai-card__label">📝 文档总结</div>
              <div class="ai-card__text">本文档主要介绍了相关主题的核心概念和应用场景……</div>
            </div>
            <div class="ai-card">
              <div class="ai-card__label">💬 问答</div>
              <div class="ai-card__text">选中任意文本，可进行解释、翻译、提问等操作</div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reader-page { display: flex; flex-direction: column; height: 100vh; overflow: hidden; }
.toolbar { height: 48px; background: var(--bg-card); border-bottom: 1px solid var(--border-color); display: flex; align-items: center; padding: 0 0.75rem; gap: 0.3rem; flex-shrink: 0; }
.toolbar__left, .toolbar__right { display: flex; align-items: center; gap: 0.3rem; }
.toolbar__center { flex: 1; text-align: center; font-size: 0.8rem; color: var(--text-secondary); font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tlb { width: 32px; height: 32px; border: none; border-radius: var(--radius); background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: all 0.1s; }
.tlb:hover { background: var(--accent-light); color: var(--accent); }
.tlb svg { width: 18px; height: 18px; }
.tlb-divider { width: 1px; height: 20px; background: var(--border-color); margin: 0 0.3rem; }
.tlb-label { font-size: 0.75rem; color: var(--text-muted); padding: 0 0.3rem; }
.tlb-text { padding: 0.3rem 0.6rem; border: none; border-radius: var(--radius); background: transparent; font-family: var(--font-sans); font-size: 0.75rem; color: var(--text-secondary); cursor: pointer; }
.tlb-text:hover { background: var(--accent-light); color: var(--accent); }
.reader-body { flex: 1; display: flex; overflow: hidden; }
.left-panel { width: var(--left-w); background: var(--bg-card); border-right: 1px solid var(--border-color); display: flex; flex-direction: column; flex-shrink: 0; }
.panel-header { padding: 0.7rem 1rem; border-bottom: 1px solid var(--border-color); font-size: 0.75rem; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; display: flex; align-items: center; gap: 0.4rem; }
.panel-header svg { width: 14px; height: 14px; }
.panel-body { flex: 1; overflow-y: auto; padding: 0.5rem 0; }
.toc-item { padding: 0.4rem 1rem; cursor: pointer; font-size: 0.8rem; color: var(--text-secondary); transition: all 0.1s; border-left: 2px solid transparent; }
.toc-item:hover { background: var(--accent-light); color: var(--text-primary); }
.toc-item.active { border-left-color: var(--accent); color: var(--accent); background: var(--accent-light); font-weight: 500; }
.toc-item.l2 { padding-left: 1.8rem; font-size: 0.75rem; }
.toc-item.l3 { padding-left: 2.6rem; font-size: 0.7rem; }
.left-footer { padding: 0.7rem 1rem; border-top: 1px solid var(--border-color); font-size: 0.7rem; color: var(--text-muted); }
.center-panel { flex: 1; overflow-y: auto; background: var(--bg-card); padding: 2rem 3rem; }
.center-panel .page-indicator { text-align: center; font-size: 0.7rem; color: var(--text-muted); margin-bottom: 1rem; }
.doc-title { font-size: 1.6rem; font-weight: 700; line-height: 1.3; margin-bottom: 0.3rem; }
.doc-meta { font-size: 0.75rem; color: var(--text-muted); margin-bottom: 1.5rem; padding-bottom: 0.8rem; border-bottom: 1px solid var(--border-color); }
.doc-body { font-family: var(--font-serif); font-size: 0.95rem; line-height: 1.85; max-width: 680px; margin: 0 auto; }
.doc-body p { margin-bottom: 0.9rem; }
.right-panel { width: var(--right-w); background: var(--bg-card); border-left: 1px solid var(--border-color); display: flex; flex-direction: column; flex-shrink: 0; }
.tab-bar { display: flex; border-bottom: 1px solid var(--border-color); }
.tab-item { flex: 1; padding: 0.6rem; text-align: center; font-size: 0.8rem; font-weight: 500; color: var(--text-muted); cursor: pointer; border-bottom: 2px solid transparent; transition: all 0.1s; }
.tab-item:hover { color: var(--text-primary); }
.tab-item.active { color: var(--accent); border-bottom-color: var(--accent); }
.right-body { flex: 1; overflow-y: auto; padding: 0.75rem; }
.anno-card { padding: 0.75rem; border: 1px solid var(--border-color); border-radius: var(--radius-lg); margin-bottom: 0.6rem; cursor: pointer; }
.anno-card:hover { border-color: var(--accent); }
.anno-card__text { font-size: 0.8rem; line-height: 1.5; color: var(--text-primary); display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
.anno-card__meta { font-size: 0.65rem; color: var(--text-muted); margin-top: 0.4rem; display: flex; gap: 0.4rem; align-items: center; }
.anno-badge { padding: 0.05rem 0.35rem; border-radius: 4px; font-size: 0.6rem; font-weight: 600; }
.badge-highlight { background: #FEF3C7; color: #92400E; }
.badge-underline { background: #DBEAFE; color: #1E40AF; }
.ai-card { padding: 0.75rem; background: var(--accent-light); border-radius: var(--radius-lg); margin-bottom: 0.6rem; }
.ai-card__label { font-size: 0.65rem; font-weight: 600; color: var(--accent); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.4rem; }
.ai-card__text { font-size: 0.8rem; line-height: 1.6; color: var(--text-primary); }
</style>
