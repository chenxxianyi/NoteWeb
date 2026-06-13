<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '../components/layout/AppLayout.vue'

const activeNoteId = ref(1)
const searchQuery = ref('')
const activeTag = ref('全部')
const tags = ['全部', '阅读笔记', '技术', '随笔']

const notes = ref([
  { id: 1, title: '注意力机制的理解与思考', preview: '注意力机制的本质是让模型学会关注输入序列中重要的部分，而不是平均对待每一个元素...', tag: '阅读笔记', tagClass: 'tag-reading', doc: '深度学习入门 · 2 小时前' },
  { id: 2, title: '虚拟DOM的性能分析', preview: '虚拟 DOM 的性能优势不在于比真实 DOM 快，而在于它让开发者不必手动操作 DOM...', tag: '技术', tagClass: 'tag-tech', doc: 'Vue.js 设计与实现 · 昨天' },
  { id: 3, title: 'Rust所有权系统笔记', preview: 'Rust 的所有权系统在编译时就能保证内存安全，无需垃圾回收。每个值有且只有一个所有者...', tag: '阅读笔记', tagClass: 'tag-reading', doc: 'Rust 程序设计 · 3 天前' },
  { id: 4, title: '分布式系统的挑战', preview: '分布式系统中最难处理的是部分失败和网络分区问题。CAP 定理告诉我们一致性、可用性...', tag: '技术', tagClass: 'tag-tech', doc: 'DDIA · 5 天前' },
  { id: 5, title: '周末读书随想', preview: '今天读完了《设计数据密集型应用》的第一部分，对分布式系统的理解又加深了一层...', tag: '随笔', tagClass: 'tag-life', doc: '1 周前' },
])

function selectNote(id: number) {
  activeNoteId.value = id
}
</script>

<template>
  <AppLayout>
    <div class="notes-page">
      <!-- Note List -->
      <div class="note-list">
        <div class="nl-header">
          <h2>笔记</h2>
          <button class="nl-header__btn" title="新建笔记">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
        </div>
        <div class="nl-search">
          <input v-model="searchQuery" type="text" placeholder="搜索笔记..." />
        </div>
        <div class="nl-tags">
          <button
            v-for="tag in tags"
            :key="tag"
            :class="['nl-tag', { active: activeTag === tag }]"
            @click="activeTag = tag"
          >
            {{ tag }}
          </button>
        </div>
        <div class="nl-items">
          <div
            v-for="note in notes"
            :key="note.id"
            :class="['nl-item', { active: activeNoteId === note.id }]"
            @click="selectNote(note.id)"
          >
            <div class="nl-item__title">{{ note.title }}</div>
            <div class="nl-item__preview">{{ note.preview }}</div>
            <div class="nl-item__meta">
              <span :class="['nl-item__tag', note.tagClass]">{{ note.tag }}</span>
              {{ note.doc }}
            </div>
          </div>
        </div>
      </div>

      <!-- Note Editor -->
      <div class="note-editor">
        <div class="ne-toolbar">
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/><path d="M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="17" y1="10" x2="3" y2="10"/><line x1="21" y1="6" x2="3" y2="6"/><line x1="21" y1="14" x2="3" y2="14"/><line x1="17" y1="18" x2="3" y2="18"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/><path d="M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/></svg></button>
          <div class="ne-divider"></div>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M10 6H6a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-4"/><path d="M14 2h6v6"/><path d="M10 14L20 4"/></svg></button>
          <div class="ne-divider"></div>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg></button>
        </div>
        <input class="ne-title" type="text" placeholder="笔记标题..." value="注意力机制的理解与思考" />
        <textarea class="ne-body" placeholder="开始书写...">注意力机制的本质是让模型学会关注输入序列中重要的部分，而不是平均对待每一个元素。

## 核心思想
- 通过 query、key、value 的机制计算注意力权重
- 权重越高，模型对该位置的关注度越大

## 常见类型
1. **加性注意力** — 使用前馈网络计算注意力分数
2. **乘性注意力** — 使用点积计算，效率更高
3. **自注意力** — Transformer 的核心机制

Transformer 中使用了多头注意力（Multi-Head Attention），将输入映射到多个子空间，并行计算注意力。</textarea>
        <div class="ne-footer">
          <span>最后编辑：2 小时前 · 来自「深度学习入门」</span>
          <div class="ne-footer__actions">
            <button class="ne-footer__btn">从批注生成</button>
            <button class="ne-footer__btn">保存</button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.notes-page {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* Left: Note List */
.note-list {
  width: 340px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.nl-header {
  padding: 1rem 1.2rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.nl-header h2 { font-family: var(--font-display); font-size: 1.1rem; font-weight: 500; }
.nl-header__btn {
  width: 32px; height: 32px; border: 1px solid var(--border-color); border-radius: 50%;
  background: transparent; display: flex; align-items: center; justify-content: center;
  cursor: pointer; color: var(--accent); transition: all 0.12s;
}
.nl-header__btn:hover { background: var(--accent-light); }
.nl-header__btn svg { width: 16px; height: 16px; }

.nl-search { padding: 0.7rem 1rem; border-bottom: 1px solid var(--border-color); }
.nl-search input {
  width: 100%; padding: 0.4rem 0.7rem; border: 1px solid var(--border-color);
  border-radius: 16px; background: var(--bg-page); font-family: var(--font-ui);
  font-size: 0.78rem; color: var(--text-primary); outline: none;
}
.nl-search input:focus { border-color: var(--accent); }

.nl-tags { padding: 0.5rem 1rem; border-bottom: 1px solid var(--border-color); display: flex; gap: 0.3rem; flex-wrap: wrap; }
.nl-tag { padding: 0.15rem 0.5rem; border: 1px solid var(--border-color); border-radius: 10px; background: transparent; font-family: var(--font-ui); font-size: 0.65rem; color: var(--text-secondary); cursor: pointer; transition: all 0.1s; }
.nl-tag:hover { border-color: var(--accent); color: var(--accent); }
.nl-tag.active { background: var(--accent); color: #fff; border-color: var(--accent); }

.nl-items { flex: 1; overflow-y: auto; }
.nl-item { padding: 0.8rem 1.2rem; border-bottom: 1px solid var(--border-color); cursor: pointer; transition: background 0.1s; }
.nl-item:hover { background: var(--accent-light); }
.nl-item.active { background: var(--accent-light); border-left: 3px solid var(--accent); }
.nl-item__title { font-size: 0.88rem; font-weight: 500; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.nl-item__preview { font-size: 0.78rem; color: var(--text-secondary); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-top: 0.15rem; }
.nl-item__meta { font-family: var(--font-ui); font-size: 0.65rem; color: var(--text-muted); margin-top: 0.25rem; display: flex; gap: 0.3rem; align-items: center; }
.nl-item__tag { padding: 0.05rem 0.3rem; border-radius: 3px; font-size: 0.6rem; font-weight: 600; }
.tag-reading { background: #FDE68A; color: #92400E; }
.tag-tech { background: #DBEAFE; color: #1E40AF; }
.tag-life { background: #D1FAE5; color: #065F46; }

/* Right: Note Editor */
.note-editor { flex: 1; display: flex; flex-direction: column; background: var(--bg-page); }
.ne-toolbar { padding: 0.6rem 1.5rem; border-bottom: 1px solid var(--border-color); display: flex; align-items: center; gap: 0.3rem; flex-wrap: wrap; }
.ne-btn { width: 30px; height: 30px; border: none; border-radius: 4px; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: all 0.1s; }
.ne-btn:hover { background: var(--accent-light); color: var(--accent); }
.ne-btn svg { width: 16px; height: 16px; }
.ne-divider { width: 1px; height: 18px; background: var(--border-color); margin: 0 0.2rem; }

.ne-title { padding: 1.2rem 1.5rem 0.5rem; border: none; font-family: var(--font-display); font-size: 1.3rem; font-weight: 500; color: var(--text-primary); background: transparent; outline: none; width: 100%; }
.ne-title::placeholder { color: var(--text-muted); }

.ne-body { flex: 1; padding: 0.5rem 1.5rem 2rem; border: none; font-family: var(--font-body); font-size: 0.95rem; line-height: 1.8; color: var(--text-primary); background: transparent; outline: none; resize: none; width: 100%; }
.ne-body::placeholder { color: var(--text-muted); }

.ne-footer { padding: 0.6rem 1.5rem; border-top: 1px solid var(--border-color); display: flex; align-items: center; justify-content: space-between; font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); }
.ne-footer__actions { display: flex; gap: 0.5rem; }
.ne-footer__btn { padding: 0.25rem 0.7rem; border: 1px solid var(--border-color); border-radius: 14px; background: var(--bg-card); font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-secondary); cursor: pointer; transition: all 0.1s; }
.ne-footer__btn:hover { border-color: var(--accent); color: var(--accent); }

@media (max-width: 820px) {
  .note-list { width: 280px; }
  .ne-title { font-size: 1.1rem; padding: 1rem 1rem 0.5rem; }
  .ne-body { padding: 0.5rem 1rem 2rem; }
  .ne-toolbar { padding: 0.5rem 1rem; }
}
@media (max-width: 620px) {
  .note-list { width: 100%; }
  .note-editor { display: none; }
}
</style>
