# 搜索弹窗设计

## 概述

将 Timeline 视图中的内联搜索改为 Modal 弹窗模式。

## 交互流程

1. 点击 🔍 图标 → 弹出 `SearchModal` 弹窗
2. 输入关键词 → 实时搜索（debounce 300ms）
3. 显示搜索结果列表
4. 点击结果 → 关闭弹窗并编辑该笔记
5. 点击外部/按 ESC/点击 X → 关闭弹窗

## 实现

### 新增组件

**`SearchModal.vue`**

Props:
- `show: boolean` - 控制显示/隐藏

Events:
- `close` - 关闭弹窗
- `select(memo)` - 选择了一条搜索结果

Emits:
- 点击外部关闭
- ESC 键关闭
- X 按钮关闭

### 修改组件

**`TimelineView.vue`**

- 移除 `div.search-container`（内联搜索）
- 移除 `showSearch` ref（替换为 `showSearchModal`）
- 移除 `searchQuery`、`searchResults`、`searchMode`、`debouncedSearch`、`doSearch`（移入 SearchModal）
- 点击搜索图标 → `showSearchModal = true`
- 点击结果 → 编辑笔记

### 复用

- `api.searchMemos(q)` 接口不变
- `MemoCard` 组件复用
