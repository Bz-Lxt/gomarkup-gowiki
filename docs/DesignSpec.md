# GoWiki Design Spec

## 方向

编辑室纸稿：暖色纸张底、墨色正文、赤陶强调。拒绝紫色渐变与 Inter 套路。中文阅读优先，编辑区域像摊开的稿纸。

## 色板

| Token | Hex | 用途 |
|---|---|---|
| `--paper` | `#F3EDE3` | 全局背景 |
| `--ink` | `#1C1712` | 主文字 |
| `--muted` | `#6B6258` | 次级文字 |
| `--terracotta` | `#C45C26` | 主操作 / 当前文档 |
| `--forest` | `#2F6F4E` | 成功 / 在线 |
| `--sidebar` | `#E7DDD0` | 左栏 |
| `--card` | `#FFFBF5` | 卡片 |
| `--line` | `#D8CCBA` | 分割线 |
| `--danger` | `#A33B2B` | 删除 |

## 字体

- 展示 / 标题：`"Fraunces", "Noto Serif SC", serif`
- 正文：`"IBM Plex Sans", "Noto Sans SC", sans-serif`
- 代码：`"IBM Plex Mono", ui-monospace, monospace`

## 布局

三栏工作区：左 280px 文档树，中弹性编辑器，右 320px 版本 / 在线协作者。工作台为满宽卡片网格，无 `max-w-*` 限宽（登录页除外）。

## 组件

- 树节点：12px 圆角、拖拽时 0.96 缩放与赤陶描边
- 按钮：实心赤陶 / 幽灵墨色，禁用降低 40% 透明度
- Toast：自定义，可手动关闭，5s 消失
- Modal：自定义遮罩，禁止 `alert/confirm`
- Diff：删除行 `#F4D6D0`，新增行 `#D7EBD8`

## 响应式

768px 隐藏左栏为抽屉；480px 单列工作台，工具条折行。
