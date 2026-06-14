<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { useAnnotationStore } from '../stores/annotationStore'

const route = useRoute()
const router = useRouter()
const documentStore = useDocumentStore()
const annotationStore = useAnnotationStore()

const docId = computed(() => Number(route.params.documentId))
const panelLeftOpen = ref(false)
const panelRightOpen = ref(false)
const topbarHidden = ref(false)
const showAnnoTab = ref(true)
const loading = ref(true)
let lastScroll = 0

const doc = computed(() => documentStore.currentDocument)
const content = computed(() => documentStore.documentContent)
const annotations = computed(() => annotationStore.annotations)

function toggleLeft() { panelLeftOpen.value = !panelLeftOpen.value }
function toggleRight() { panelRightOpen.value = !panelRightOpen.value }
function closePanels() { panelLeftOpen.value = false; panelRightOpen.value = false }

function handleScroll() {
  const cur = window.scrollY
  topbarHidden.value = cur > 200 && cur > lastScroll
  lastScroll = cur
}

onMounted(async () => {
  window.addEventListener('scroll', handleScroll)
  try {
    await Promise.all([
      documentStore.fetchDocument(docId.value),
      documentStore.fetchDocumentContent(docId.value),
      annotationStore.fetchAnnotations(docId.value),
    ])
  } catch {
    // document not found — redirect
    router.push('/documents')
  }
  loading.value = false
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <div class="reader-page">
    <!-- Overlay -->
    <div :class="['panel-overlay', { show: panelLeftOpen || panelRightOpen }]" @click="closePanels"></div>

    <!-- Floating Top Bar -->
    <div :class="['reader-topbar', { hidden: topbarHidden }]">
      <button class="tb-btn" title="返回" @click="router.push('/documents')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
      </button>
      <div class="tb-divider"></div>
      <button class="tb-btn" title="目录" @click="toggleLeft">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
      </button>
      <button class="tb-btn" title="批注和AI" @click="toggleRight">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
      </button>
      <div class="tb-divider"></div>
      <button class="tb-btn" title="搜索">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
      </button>
      <button class="tb-btn" title="缩小">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
      </button>
      <span class="tb-label">100%</span>
      <button class="tb-btn" title="放大">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
      </button>
      <div class="tb-divider"></div>
      <button class="tb-btn" title="AI总结">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
      </button>
    </div>

    <!-- Left Panel: TOC (placeholder — real TOC needs parsed content) -->
    <div :class="['panel-left', { open: panelLeftOpen }]">
      <div class="panel__header">
        <h3>目录</h3>
        <button @click="panelLeftOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div class="toc-item level-1 active">{{ doc?.title || '文档' }}</div>
        <div class="toc-placeholder">目录解析暂未实现</div>
      </div>
    </div>

    <!-- Right Panel: Annotations + AI -->
    <div :class="['panel-right', { open: panelRightOpen }]">
      <div class="panel__header">
        <h3>批注与 AI</h3>
        <button @click="panelRightOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div class="panel-tabs">
          <button :class="['pt-btn', { active: showAnnoTab }]" @click="showAnnoTab = true">
            批注 ({{ annotations.length }})
          </button>
          <button :class="['pt-btn', { active: !showAnnoTab }]" @click="showAnnoTab = false">
            AI 助手
          </button>
        </div>

        <!-- Annotations tab -->
        <div v-if="showAnnoTab">
          <div v-if="annotations.length === 0" class="anno-empty">暂无批注，选中文本后可添加</div>
          <div v-for="(anno, idx) in annotations" :key="idx" class="anno-card">
            <div class="anno-card__text">{{ anno.selected_text }}</div>
            <div class="anno-card__meta">
              <span class="anno-highlight">{{ anno.type }}</span>
              第{{ anno.page }}页
            </div>
          </div>
        </div>

        <!-- AI tab -->
        <div v-else>
          <div class="ai-message">
            <div class="ai-message__label">AI 总结</div>
            <p>你的 AI 阅读助手已就绪。选中文本后可进行解释、翻译或提问。</p>
          </div>
          <div class="ai-input">
            <input type="text" placeholder="向 AI 提问..." />
          </div>
        </div>
      </div>
    </div>

    <!-- Right page edge decoration -->
    <div class="page-edge"></div>

    <!-- Reading Content -->
    <div v-if="loading" class="reader-content">
      <div class="reader-inner"><p style="text-align:center;padding:4rem 0;color:var(--text-muted)">加载中...</p></div>
    </div>

    <div v-else class="reader-content">
      <div class="reader-inner">
        <h1 class="doc-title">{{ doc?.title || '未命名文档' }}</h1>
        <div class="doc-meta">
          {{ doc?.file_type?.toUpperCase() || '' }}
          {{ doc?.read_progress ? `· 已读至${doc.read_progress}%` : '' }}
        </div>

        <div v-if="content" class="doc-body" v-html="content"></div>
        <div v-else class="doc-body">
          <p style="color:var(--text-muted)">文档内容暂不可用</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reader-page { position: relative; min-height: 100vh; background: var(--bg-page); background-image: repeating-linear-gradient(0deg, transparent, transparent 1px, rgba(0,0,0,0.005) 1px, rgba(0,0,0,0.005) 2px); background-size: 100% 2px; font-family: var(--font-body); }

.panel-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.15); z-index: 15; opacity: 0; pointer-events: none; transition: opacity 0.3s; }
.panel-overlay.show { opacity: 1; pointer-events: auto; }

.reader-topbar { position: fixed; top: 0.75rem; left: 50%; transform: translateX(-50%); z-index: 30; display: flex; align-items: center; gap: 0.3rem; padding: 0.35rem 0.6rem; background: rgba(250,248,245,0.92); backdrop-filter: blur(8px); border: 1px solid var(--border-color); border-radius: 24px; box-shadow: 0 2px 12px rgba(61,46,36,0.08); transition: opacity 0.3s, transform 0.3s; font-family: var(--font-ui); }
.reader-topbar.hidden { opacity: 0; transform: translateX(-50%) translateY(-10px); pointer-events: none; }
.tb-btn { width: 34px; height: 34px; border: none; border-radius: 50%; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: all 0.12s; }
.tb-btn:hover { background: var(--accent-light); color: var(--accent); }
.tb-btn svg { width: 18px; height: 18px; }
.tb-divider { width: 1px; height: 20px; background: var(--border-color); margin: 0 0.2rem; }
.tb-label { font-size: 0.7rem; color: var(--text-muted); padding: 0 0.4rem; }

.panel-left, .panel-right { position: fixed; top: 0; bottom: 0; width: 300px; background: var(--bg-card); z-index: 20; transition: transform 0.3s; display: flex; flex-direction: column; }
.panel-left { left: 0; border-right: 1px solid var(--border-color); transform: translateX(-100%); }
.panel-left.open { transform: translateX(0); }
.panel-right { right: 0; border-left: 1px solid var(--border-color); transform: translateX(100%); }
.panel-right.open { transform: translateX(0); }
.panel__header { display: flex; align-items: center; justify-content: space-between; padding: 1rem 1.25rem; border-bottom: 1px solid var(--border-color); font-family: var(--font-ui); flex-shrink: 0; }
.panel__header h3 { font-size: 0.9rem; font-weight: 500; color: var(--text-primary); }
.panel__header button { background: none; border: none; cursor: pointer; color: var(--text-muted); font-size: 1.2rem; padding: 0.2rem; }
.panel__header button:hover { color: var(--text-primary); }
.panel__body { flex: 1; overflow-y: auto; padding: 1rem 1.25rem; }

.panel-tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.pt-btn { padding: 0.3rem 1rem; border-radius: 20px; border: none; font-family: var(--font-ui); font-size: 0.8rem; cursor: pointer; transition: all 0.12s; background: transparent; color: var(--text-muted); }
.pt-btn.active { background: var(--accent-light); color: var(--accent); font-weight: 500; }

.toc-item { padding: 0.5rem 0; cursor: pointer; color: var(--text-secondary); font-family: var(--font-ui); font-size: 0.85rem; border-bottom: 1px solid var(--border-color); }
.toc-item.level-1 { font-weight: 500; }
.toc-placeholder { padding: 1rem 0; color: var(--text-muted); font-family: var(--font-ui); font-size: 0.8rem; text-align: center; }

.anno-card { padding: 0.8rem 0; border-bottom: 1px solid var(--border-color); }
.anno-card__text { font-size: 0.85rem; line-height: 1.6; color: var(--text-primary); }
.anno-card__meta { font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); margin-top: 0.3rem; display: flex; gap: 0.4rem; align-items: center; }
.anno-highlight { display: inline-block; background: #FDE68A; padding: 0 0.3rem; border-radius: 3px; font-size: 0.65rem; color: #92400E; }
.anno-empty { text-align: center; padding: 2rem 0; color: var(--text-muted); font-family: var(--font-ui); font-size: 0.8rem; }

.ai-message { padding: 0.8rem; background: var(--accent-light); border-radius: 8px; margin-bottom: 0.8rem; font-size: 0.85rem; line-height: 1.7; }
.ai-message__label { font-family: var(--font-ui); font-size: 0.7rem; color: var(--accent); font-weight: 600; margin-bottom: 0.3rem; }
.ai-input { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
.ai-input input { flex: 1; padding: 0.5rem 0.8rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-page); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-primary); outline: none; }
.ai-input input:focus { border-color: var(--accent); }

.page-edge { position: fixed; top: 0; right: 0; bottom: 0; width: clamp(60px, 8vw, 140px); background: linear-gradient(to right, transparent, rgba(0,0,0,0.015) 40%, rgba(0,0,0,0.025)); pointer-events: none; z-index: 1; }

.reader-content { max-width: 920px; margin: 0 auto; padding: 4.5rem 3rem 6rem; min-height: 100vh; }
.reader-inner { max-width: 680px; margin: 0 auto; }
.doc-title { font-family: var(--font-display); font-size: 2rem; font-weight: 600; line-height: 1.3; margin-bottom: 0.5rem; color: var(--text-primary); }
.doc-meta { font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-muted); margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid var(--border-color); }
.doc-body { font-size: 1rem; line-height: 1.9; color: var(--text-primary); }
.doc-body p { margin-bottom: 1rem; }

@media (max-width: 1024px) { .reader-content { max-width: 100%; padding: 4.5rem 2rem 5rem; } .reader-inner { max-width: 100%; } }
@media (max-width: 600px) { .page-edge { display: none; } .reader-content { padding: 3.5rem 1rem 4rem; } .panel-left, .panel-right { width: 100%; } }
</style>
