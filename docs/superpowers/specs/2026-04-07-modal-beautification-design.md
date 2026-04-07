# Modal Beautification Design

## Overview

Beautify all modal dialogs (SearchModal, SanaEditor) and replace native `alert()`/`confirm()` with custom Toast components, while maintaining the existing warm paper aesthetic.

## Design Language

**Color Palette (Warm Paper — Style B)**
- Background: `#F5F0E8`
- Card: `#EAE5DC`
- Border: `#D4CCBA`
- Text primary: `#3D3830`
- Text secondary: `#A09888`
- Accent/Primary: `#6B8FCC` (blue)
- Danger: `#C06050`
- Cancel button: `#E0DBD3` bg, `#6B6358` text

**Overlay**
- Background: `rgba(61, 56, 48, 0.3)`
- `backdrop-filter: blur(4px)`

**Animation**
- Enter: `opacity 0→1, scale 0.95→1`, 200ms ease-out
- Exit: `opacity 1→0, scale 1→0.95`, 150ms ease-in

## Components

### 1. SearchModal (搜索弹框)
- Rounded corners: 20px
- Shadow: `0 8px 32px rgba(61,56,48,0.12)`
- Width: 90% max 560px, max-height 70vh
- Header: title + ✕ close button
- Input: 1px `#D4CCBA` border, 8px radius, `#6B8FCC` on focus
- Results: scrollable list of SanaCard

### 2. SanaEditor (编辑弹框)
- Same overlay and animation as SearchModal
- Rounded corners: 20px
- Shadow: `0 8px 32px rgba(61,56,48,0.12)`
- Width: 90% max 600px
- Textarea: full width, 200px min-height, 16px padding
- Footer: cancel (warm gray) + save (blue `#6B8FCC`) buttons

### 3. Toast (替换原生 alert/confirm)
- Fixed position bottom-center
- Rounded pill shape: 12px radius
- Shadow: `0 4px 12px rgba(0,0,0,0.15)`
- Auto-dismiss: 3s for success/info, manual dismiss for errors
- Types:
  - **Success**: soft green bg `#D4EDDA`, green text `#155724`
  - **Error**: soft red bg `#F8D7DA`, red text `#721C24`
  - **Info**: warm white bg `#EAE5DC`, dark text `#3D3830`
- Enter animation: slide up + fade in
- Exit animation: slide down + fade out

### 4. ConfirmDialog (替换原生 confirm)
- Shared overlay + animation as other modals
- Card with title, message, Cancel + Confirm buttons
- Danger actions (delete): Confirm button uses `#C06050` (danger red)
- Non-danger: uses `#6B8FCC` (accent blue)

## Interactions

- ESC key closes SearchModal and SanaEditor
- Click outside closes SearchModal (but NOT SanaEditor — prevents accidental data loss)
- Click outside closes ConfirmDialog
- Toast auto-dismiss unless error type

## Files to Modify

1. `frontend/src/components/SearchModal.vue` — add animation, blur overlay, warm styling
2. `frontend/src/components/SanaEditor.vue` — add animation, blur overlay, warm styling
3. `frontend/src/components/Toast.vue` (new) — toast notification component
4. `frontend/src/components/ConfirmDialog.vue` (new) — confirm dialog component
5. `frontend/src/views/TimelineView.vue` — replace `alert()`/`confirm()` calls with Toast/ConfirmDialog
6. `frontend/src/App.vue` (optional) — if any global toast needed
7. `frontend/src/main.js` or `App.vue` — register Toast/ConfirmDialog as global Vue plugins

## Acceptance Criteria

- [ ] SearchModal has blur overlay + scale animation on open
- [ ] SanaEditor has blur overlay + scale animation on open
- [ ] ESC closes SearchModal and SanaEditor
- [ ] Native `alert('导出失败')` → Toast (error type)
- [ ] Native `alert('导入完成 N 条')` → Toast (success type)
- [ ] Native `confirm('确定删除')` → ConfirmDialog
- [ ] All toast auto-dismiss in 3s (except error)
- [ ] No regressions in existing functionality
