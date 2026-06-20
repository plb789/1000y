# UI模块配置说明

## 目录结构

```
Frontend/UI/
├── index.json              # UI模块总配置文件
├── Common/                 # 通用组件和主题
│   └── theme.json          # 主题配置
├── ChatUI/                # 聊天界面
│   └── config.json
├── RoleUI/                # 角色信息界面
│   └── config.json
├── SkillBarUI/            # 技能栏界面
│   └── config.json
├── InventoryUI/           # 背包界面
│   └── config.json
├── ShopUI/                # 商店界面
│   └── config.json
├── SocialUI/              # 社交界面（好友、组队、门派）
│   └── config.json
├── MiniMapUI/             # 小地图界面
│   └── config.json
└── SettingsUI/            # 设置界面
    └── config.json
```

## 配置格式

每个UI模块的config.json遵循以下格式：

```json
{
  "name": "模块名称",
  "version": "1.0",
  "description": "模块描述",
  "components": {
    "组件名": {
      "enabled": true,
      "position": {
        "x": 20,           // 数字或 'auto' 或 'center'
        "y": 60,
        "right": 20,       // 可选
        "bottom": 90       // 可选
      },
      "size": {
        "width": 220,
        "height": "auto"   // 数字或 'auto'
      },
      "style": {
        "background": "rgba(0, 0, 0, 0.8)",
        "border": "2px solid #e94560",
        "borderRadius": 10,
        "opacity": 1
      },
      "zIndex": 10         // 可选，相对于基础层级
    }
  }
}
```

## 位置说明

- `x`, `y`: 基础坐标
- `right`: 距离右侧的距离（设置后left失效）
- `bottom`: 距离底部的距离（设置后top失效）
- `center`: 居中对齐

## 尺寸说明

- 数字: 固定像素值
- `'auto'`: 自适应内容
- `'100%'`: 百分比（仅部分属性支持）

## 样式说明

- `background`: CSS背景属性
- `border`: CSS边框属性
- `borderRadius`: 圆角（数字，px）
- `opacity`: 透明度（0-1）

## 主题配置 (Common/theme.json)

```json
{
  "theme": {
    "primary": "#e94560",
    "secondary": "#4a5568",
    "success": "#4ade80",
    "warning": "#fbbf24",
    "danger": "#ef4444",
    "info": "#60a5fa"
  },
  "animations": {
    "duration": {
      "fast": 150,
      "normal": 300,
      "slow": 500
    }
  }
}
```

## 编辑器使用

1. 打开 `UIEditor.html` 或 `UIEditor/index.html`
2. 选择要编辑的UI模块
3. 修改属性（可视化或JSON编辑）
4. 保存/导出配置

## 客户端加载

客户端通过 `UIConfigManager` 和 `UIDynamicLoader` 自动加载UI配置：

```javascript
// 初始化
const configManager = new UIConfigManager(game);
await configManager.loadAll();

// 获取配置
const config = configManager.getConfig('ChatUI');

// 应用到元素
const loader = new UIDynamicLoader(game);
loader.applyConfigToElement('ChatUI', 'chatBox', document.getElementById('chat'));
```

## 扩展开发

如需添加新的UI模块：

1. 在 `Frontend/UI/` 下创建新模块文件夹
2. 添加 `config.json` 配置文件
3. 在 `index.json` 的 `modules` 数组中添加模块名
4. 更新编辑器模块列表
