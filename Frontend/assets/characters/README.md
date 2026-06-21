# 角色资源系统说明

## 工作原理

`CharacterRenderer` 支持**双模式自动切换**：

1. **资源模式（优先）**：检测到 `assets/characters/manifest.json` 时，加载并使用图片资源
2. **矢量模式（降级）**：找不到清单文件或加载失败时，使用 Canvas 矢量绘制卡通人形

切换是**全自动**的，无需修改代码。

## 8方向支持

系统支持完整的 8 方向移动和动画：

```
        up(4)
   up_left(3)  up_right(5)
        |
left(2)─+─right(6)
        |
  down_left(1) down_right(7)
        down(0)
```

### 方向编号与命名

| 编号 | 方向名 | 说明 |
|------|--------|------|
| 0 | down | 下 |
| 1 | down_left | 左下 |
| 2 | left | 左 |
| 3 | up_left | 左上 |
| 4 | up | 上 |
| 5 | up_right | 右上 |
| 6 | right | 右 |
| 7 | down_right | 右下 |

### 方向判定算法

使用 `Math.atan2(dy, dx)` 计算移动角度，将 360° 分成 8 个 45° 区间，自动映射到对应方向。无论玩家向哪个方向移动，都会自动选择最接近的方向动画。

## 如何替换为图片资源

### 步骤 1：准备资源

按以下目录结构放置图片（推荐 48x48 PNG，带透明通道）：

```
Frontend/
└── assets/
    └── characters/
        ├── manifest.json          ← 清单文件（必需，启用资源模式）
        ├── male/
        │   ├── idle_down_0.png
        │   ├── idle_down_left_0.png
        │   ├── walk_down_0.png
        │   ├── attack_down_left_0.png
        │   └── ...
        └── female/
            ├── idle_down_0.png
            └── ...
```

### 步骤 2：创建清单文件

复制 `manifest.example.json` 为 `manifest.json`，按实际拥有的图片填写。

清单格式（8方向完整版）：

```json
{
  "version": 2,
  "base_dir": "assets/characters/",
  "frame_duration": 120,
  "directions": ["down", "down_left", "left", "up_left", "up", "up_right", "right", "down_right"],
  "characters": {
    "male": {
      "idle_down": ["male/idle_down_0.png", "male/idle_down_1.png"],
      "idle_down_left": ["male/idle_down_left_0.png"],
      "walk_down": ["male/walk_down_0.png", "male/walk_down_1.png", "male/walk_down_2.png"],
      "attack_down_left": ["male/attack_down_left_0.png"]
    },
    "female": { ... }
  }
}
```

### 步骤 3：刷新页面

控制台会显示：
- `[CharacterRenderer] 资源模式已启用 N 帧` → 资源模式生效
- `[CharacterRenderer] 清单不存在，使用矢量模式` → 降级到矢量模式

## 命名规范

| 字段 | 取值 |
|------|------|
| gender | `male` / `female` |
| state | `idle` / `walk` / `attack` / `cast` / `dead` |
| dir | `down` / `down_left` / `left` / `up_left` / `up` / `up_right` / `right` / `down_right` |
| frame | `0`, `1`, `2`, ... |

动画键格式：`{state}_{dir}`，例如 `idle_down_left`、`attack_up_right`。

## 灵活的降级策略

### 资源模式降级链

1. **精确匹配 8 方向**：如 `attack_down_left`
2. **对角方向降级到正方向**：
   - `down_left` → `down`
   - `up_left` → `left`（或 `up`）
   - `up_right` → `right`（或 `up`）
   - `down_right` → `down`（或 `right`）
3. **方向无关动画**：如 `dead`（不区分方向）
4. **idle_down 兜底**：所有方向都没有时使用

### 其他降级特性

- **缺少某个状态**：自动降级到 `idle_down`
- **某帧加载失败**：跳过该帧，其他帧正常使用
- **所有帧都失败**：整体降级到矢量模式

## 4方向资源兼容

如果你只有 4 方向资源（down/left/right/up），也可以正常使用。系统会自动将对角方向降级到正方向：

```json
{
  "characters": {
    "male": {
      "idle_down": ["male/idle_down_0.png"],
      "idle_left": ["male/idle_left_0.png"],
      "idle_right": ["male/idle_right_0.png"],
      "idle_up": ["male/idle_up_0.png"]
    }
  }
}
```

当玩家向左下移动时，会自动使用 `idle_down` 资源。

## 部分替换

可以只为部分状态/方向提供资源，其余自动降级到矢量绘制。例如：

```json
{
  "characters": {
    "male": {
      "idle_down_left": ["male/idle_down_left_0.png"],
      "attack_up_right": ["male/attack_up_right_0.png"]
    }
  }
}
```

这样只有 `male` 的 `idle_down_left` 和 `attack_up_right` 使用图片，其他状态/方向仍使用矢量绘制。

## 矢量模式的 8 方向处理

矢量模式（无资源时）目前只绘制 4 正方向：
- 对角方向（左下、左上、右上、右下）会自动映射到最近的正方向
- 例如：向左下移动时，矢量绘制使用"下"方向

这是为了保持矢量绘制的简洁性。如果你需要矢量模式也支持 8 方向，可以在 `_drawVector` 中扩展对角方向的绘制逻辑。

## 设计建议

- 图片尺寸：48x48 像素（与 tileSize 一致，自动缩放）
- 透明背景：PNG-24 with alpha
- 锚点：底部居中（绘制时自动对齐到格子底部）
- 帧率：`frame_duration` 控制序列帧播放速度（默认 120ms/帧）
- 8方向资源制作：建议使用 3D 建模或像素画工具生成 8 个方向的序列帧

## 关闭资源模式

删除或重命名 `manifest.json` 文件即可，无需改代码。

## 集成点

- 初始化：`Game.initMapEngine()` 中调用 `loadResources()`
- 渲染：`MapEngine.render()` 中调用 `characterRenderer.draw()`
- 攻击动画：`BattleSystem.playAttackAnimation()` 触发
- 施法动画：`Game.useSkill()` 触发
- 方向判定：`setDirection(dx, dy)` 自动根据移动向量计算 8 方向
