<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const panelLeftOpen = ref(false)
const panelRightOpen = ref(false)
const topbarHidden = ref(false)
const showAnnoTab = ref(true)
let lastScroll = 0

function toggleLeft() { panelLeftOpen.value = !panelLeftOpen.value }
function toggleRight() { panelRightOpen.value = !panelRightOpen.value }
function closePanels() { panelLeftOpen.value = false; panelRightOpen.value = false }

function handleScroll() {
  const cur = window.scrollY
  topbarHidden.value = cur > 200 && cur > lastScroll
  lastScroll = cur
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

const tocItems = [
  { level: 1, label: '引言', active: true },
  { level: 1, label: '第1章 深度学习概述' },
  { level: 2, label: '1.1 什么是深度学习' },
  { level: 2, label: '1.2 深度学习的历史' },
  { level: 3, label: '1.2.1 神经网络的早期发展' },
  { level: 3, label: '1.2.2 深度学习的崛起' },
  { level: 1, label: '第2章 感知机' },
  { level: 2, label: '2.1 感知机是什么' },
  { level: 2, label: '2.2 简单逻辑电路' },
  { level: 1, label: '第3章 神经网络' },
]

const annotations = [
  { text: '"深度学习是机器学习的一个分支，它试图使用包含复杂结构或由多重非线性变换构成的多个处理层对数据进行高层抽象。"', meta: '高亮', page: '第2页 · 昨天' },
  { text: '感知机的权重和偏置参数可以通过数据自动学习。', meta: '笔记', page: '第12页 · 2天前' },
]
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

    <!-- Left Panel: TOC -->
    <div :class="['panel-left', { open: panelLeftOpen }]">
      <div class="panel__header">
        <h3>目录</h3>
        <button @click="panelLeftOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div
          v-for="(item, idx) in tocItems"
          :key="idx"
          :class="['toc-item', `level-${item.level}`, { active: item.active }]"
        >
          {{ item.label }}
        </div>
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
          <button
            :class="['pt-btn', { active: showAnnoTab }]"
            @click="showAnnoTab = true"
          >
            批注 ({{ annotations.length }})
          </button>
          <button
            :class="['pt-btn', { active: !showAnnoTab }]"
            @click="showAnnoTab = false"
          >
            AI 助手
          </button>
        </div>

        <!-- Annotations tab -->
        <div v-if="showAnnoTab">
          <div v-for="(anno, idx) in annotations" :key="idx" class="anno-card">
            <div class="anno-card__text">{{ anno.text }}</div>
            <div class="anno-card__meta">
              <span class="anno-highlight">{{ anno.meta }}</span>
              {{ anno.page }}
            </div>
          </div>
        </div>

        <!-- AI tab -->
        <div v-else>
          <div class="ai-message">
            <div class="ai-message__label">AI 总结</div>
            <p>本文介绍了深度学习的基本概念，包括感知机、神经网络和多层网络。核心思想是通过多层非线性变换来学习数据的层次化表示。</p>
          </div>
          <div class="ai-question">
            <p class="ai-message__label" style="margin-top:1rem;">向 AI 提问</p>
            <div class="ai-input">
              <input type="text" placeholder="输入你的问题..." />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Right page edge decoration -->
    <div class="page-edge"></div>

    <!-- Reading Content -->
    <div class="reader-content">
      <div class="reader-inner">
        <h1 class="doc-title">深度学习入门：基于Python的理论与实现</h1>
        <div class="doc-meta">
          PDF · 斋藤康毅 著 · 已读至第3章 · 上次阅读 2 小时前
        </div>

        <div class="doc-body">
          <h2>第3章 神经网络</h2>

          <p>上一章我们学习了感知机。本章我们将介绍<strong>神经网络</strong>。具体来说，我们将从神经元的结构开始，逐步深入到前向传播、激活函数、输出层设计等内容。</p>

          <h3>3.1 从感知机到神经网络</h3>
          <p>神经网络和感知机本质上有很多共同点。感知机中，输入信号被乘以权重后求和，然后通过阶跃函数输出结果。而在神经网络中，我们把阶跃函数换成了其他函数——<strong>激活函数</strong>。</p>

          <div class="highlight-marker">"激活函数是连接感知机和神经网络的关键桥梁。它将输入的线性组合转换为非线性输出，使网络能够表达复杂的模式。"</div>

          <p>让我们来看一个简单的例子。假设有一个两层的神经网络，输入层有2个神经元，隐藏层有3个神经元，输出层有2个神经元。</p>

          <pre><code>import numpy as np

def sigmoid(x):
    return 1 / (1 + np.exp(-x))

def init_network():
    network = {}
    network['W1'] = np.array([[0.1, 0.3, 0.5], [0.2, 0.4, 0.6]])
    network['b1'] = np.array([0.1, 0.2, 0.3])
    network['W2'] = np.array([[0.1, 0.4], [0.2, 0.5], [0.3, 0.6]])
    network['b2'] = np.array([0.1, 0.2])
    network['W3'] = np.array([[0.1, 0.3], [0.2, 0.4]])
    network['b3'] = np.array([0.1, 0.2])
    return network</code></pre>

          <h3>3.2 激活函数</h3>
          <p>激活函数的选择对神经网络的性能有着重要影响。常用的激活函数包括：</p>

          <blockquote>Sigmoid 函数：将输入映射到 (0, 1) 区间，适合二分类问题。但存在梯度消失问题。<br/>
          ReLU 函数：max(0, x)，计算简单且能缓解梯度消失，是目前最常用的激活函数之一。<br/>
          Tanh 函数：将输入映射到 (-1, 1) 区间，输出均值为0。</blockquote>

          <h3>3.3 输出层设计</h3>
          <p>神经网络可以用在分类问题和回归问题上。根据问题的性质改变输出层的激活函数：</p>

          <div class="highlight-marker">"Softmax 函数的输出之和为1，因此可以解释为概率。这使我们能够将神经网络的输出视为"置信度"的度量。"</div>

          <h3>3.4 手写数字识别</h3>
          <p>让我们通过 MNIST 数据集来实践一下。MNIST 是手写数字图像集...</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reader-page {
  position: relative;
  min-height: 100vh;
  background: var(--bg-page);
  background-image: repeating-linear-gradient(0deg, transparent, transparent 1px, rgba(0,0,0,0.005) 1px, rgba(0,0,0,0.005) 2px);
  background-size: 100% 2px;
  font-family: var(--font-body);
}

.panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.15);
  z-index: 15;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s;
}
.panel-overlay.show { opacity: 1; pointer-events: auto; }

/* Top Bar */
.reader-topbar {
  position: fixed;
  top: 0.75rem; left: 50%;
  transform: translateX(-50%);
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.35rem 0.6rem;
  background: rgba(250,248,245,0.92);
  backdrop-filter: blur(8px);
  border: 1px solid var(--border-color);
  border-radius: 24px;
  box-shadow: 0 2px 12px rgba(61,46,36,0.08);
  transition: opacity 0.3s, transform 0.3s;
  font-family: var(--font-ui);
}
.reader-topbar.hidden {
  opacity: 0;
  transform: translateX(-50%) translateY(-10px);
  pointer-events: none;
}
.tb-btn {
  width: 34px; height: 34px;
  border: none;
  border-radius: 50%;
  background: transparent;
  display: flex;
  align-items: center; justify-content: center;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.12s;
}
.tb-btn:hover { background: var(--accent-light); color: var(--accent); }
.tb-btn svg { width: 18px; height: 18px; }
.tb-divider { width: 1px; height: 20px; background: var(--border-color); margin: 0 0.2rem; }
.tb-label { font-size: 0.7rem; color: var(--text-muted); padding: 0 0.4rem; }

/* Panels */
.panel-left, .panel-right {
  position: fixed;
  top: 0; bottom: 0;
  width: 300px;
  background: var(--bg-card);
  border-color: var(--border-color);
  z-index: 20;
  transition: transform 0.3s cubic-bezier(0.4,0,0.2,1);
  display: flex;
  flex-direction: column;
}
.panel-left {
  left: 0;
  border-right: 1px solid var(--border-color);
  transform: translateX(-100%);
}
.panel-left.open { transform: translateX(0); }
.panel-right {
  right: 0;
  border-left: 1px solid var(--border-color);
  transform: translateX(100%);
}
.panel-right.open { transform: translateX(0); }
.panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border-color);
  font-family: var(--font-ui);
  flex-shrink: 0;
}
.panel__header h3 { font-size: 0.9rem; font-weight: 500; color: var(--text-primary); }
.panel__header button { background: none; border: none; cursor: pointer; color: var(--text-muted); font-size: 1.2rem; padding: 0.2rem; }
.panel__header button:hover { color: var(--text-primary); }
.panel__body { flex: 1; overflow-y: auto; padding: 1rem 1.25rem; }

.panel-tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.pt-btn {
  padding: 0.3rem 1rem;
  border-radius: 20px;
  border: none;
  font-family: var(--font-ui);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.12s;
  background: transparent;
  color: var(--text-muted);
}
.pt-btn.active { background: var(--accent-light); color: var(--accent); font-weight: 500; }
.pt-btn:not(.active):hover { color: var(--text-secondary); }

.toc-item {
  padding: 0.5rem 0;
  cursor: pointer;
  color: var(--text-secondary);
  font-family: var(--font-ui);
  font-size: 0.85rem;
  border-bottom: 1px solid var(--border-color);
  transition: color 0.1s;
}
.toc-item:hover { color: var(--accent); }
.toc-item.level-1 { font-weight: 500; }
.toc-item.level-2 { padding-left: 1.2rem; font-size: 0.8rem; }
.toc-item.level-3 { padding-left: 2.4rem; font-size: 0.75rem; }
.toc-item.active { color: var(--accent); font-weight: 600; }

.anno-card { padding: 0.8rem 0; border-bottom: 1px solid var(--border-color); }
.anno-card:last-child { border-bottom: none; }
.anno-card__text { font-size: 0.85rem; line-height: 1.6; color: var(--text-primary); }
.anno-card__meta { font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); margin-top: 0.3rem; display: flex; gap: 0.4rem; align-items: center; }
.anno-highlight { display: inline-block; background: #FDE68A; padding: 0 0.3rem; border-radius: 3px; font-size: 0.65rem; color: #92400E; }

.ai-message { padding: 0.8rem; background: var(--accent-light); border-radius: 8px; margin-bottom: 0.8rem; font-size: 0.85rem; line-height: 1.7; }
.ai-message__label { font-family: var(--font-ui); font-size: 0.7rem; color: var(--accent); font-weight: 600; margin-bottom: 0.3rem; }
.ai-input { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
.ai-input input { flex: 1; padding: 0.5rem 0.8rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-page); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-primary); outline: none; }
.ai-input input:focus { border-color: var(--accent); }

/* Page edge decoration */
.page-edge {
  position: fixed;
  top: 0; right: 0; bottom: 0;
  width: clamp(60px, 8vw, 140px);
  background: linear-gradient(to right, transparent, rgba(0,0,0,0.015) 40%, rgba(0,0,0,0.025));
  pointer-events: none;
  z-index: 1;
}

/* Reading content */
.reader-content {
  max-width: 920px;
  margin: 0 auto;
  padding: 4.5rem 3rem 6rem;
  min-height: 100vh;
}
.reader-inner { max-width: 680px; margin: 0 auto; }

.doc-title {
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.3;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}
.doc-meta {
  font-family: var(--font-ui);
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-color);
}
.doc-body { font-size: 1rem; line-height: 1.9; color: var(--text-primary); }
.doc-body h2 { font-family: var(--font-display); font-size: 1.4rem; font-weight: 600; margin: 2rem 0 0.8rem; }
.doc-body h3 { font-family: var(--font-display); font-size: 1.15rem; font-weight: 500; margin: 1.5rem 0 0.5rem; }
.doc-body p { margin-bottom: 1rem; }
.doc-body code {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  background: var(--accent-light);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}
.doc-body pre {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  padding: 1rem;
  overflow-x: auto;
  font-family: var(--font-mono);
  font-size: 0.85rem;
  line-height: 1.6;
  margin: 1rem 0;
}
.doc-body blockquote {
  border-left: 3px solid var(--accent);
  padding: 0.5rem 1rem;
  margin: 1rem 0;
  color: var(--text-secondary);
  font-style: italic;
}
.highlight-marker {
  background: #FDE68A;
  padding: 0.1rem 0.2rem;
  border-radius: 2px;
  cursor: pointer;
  display: inline;
}
.highlight-marker:hover { background: #FCD34D; }

@media (max-width: 1024px) {
  .reader-content { max-width: 100%; padding: 4.5rem 2rem 5rem; }
  .reader-inner { max-width: 100%; }
  .reader-topbar { padding: 0.3rem 0.4rem; gap: 0.15rem; }
  .tb-btn { width: 30px; height: 30px; }
  .tb-btn svg { width: 16px; height: 16px; }
  .tb-label { font-size: 0.65rem; }
}
@media (max-width: 600px) {
  .page-edge { display: none; }
  .reader-content { padding: 3.5rem 1rem 4rem; }
  .reader-inner { max-width: 100%; }
  .doc-title { font-size: 1.3rem; }
  .doc-meta { font-size: 0.7rem; margin-bottom: 1.2rem; }
  .doc-body { font-size: 0.9rem; line-height: 1.75; }
  .doc-body h2 { font-size: 1.15rem; margin: 1.2rem 0 0.5rem; }
  .doc-body h3 { font-size: 1rem; margin: 1rem 0 0.4rem; }
  .doc-body pre { font-size: 0.75rem; padding: 0.7rem; }
  .panel-left, .panel-right { width: 100%; }
  .reader-topbar { top: 0.5rem; left: 0.5rem; right: 0.5rem; transform: none; border-radius: var(--radius); padding: 0.25rem 0.3rem; gap: 0.1rem; justify-content: space-between; }
  .tb-btn { width: 28px; height: 28px; }
  .tb-btn svg { width: 15px; height: 15px; }
  .tb-divider { display: none; }
  .tb-label { font-size: 0.6rem; }
  .panel__header { padding: 0.7rem 1rem; }
  .panel__body { padding: 0.6rem 1rem; }
}
@media (max-width: 380px) {
  .reader-content { padding: 3rem 0.8rem 3rem; }
  .doc-title { font-size: 1.1rem; }
  .doc-body { font-size: 0.85rem; }
  .doc-body pre { font-size: 0.7rem; padding: 0.5rem; }
}
</style>
