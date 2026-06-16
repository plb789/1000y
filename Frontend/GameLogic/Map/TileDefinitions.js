/**
 * 瓦片类型定义 - 扩展版
 * 包含更多原版千年风格的瓦片类型
 */
const TileDefinitions = {
  // 基础地面
  grass: { id: 1, name: '草地', color: '#3a5f0b', passable: true, layer: 0, icon: '🌿' },
  grass2: { id: 2, name: '草地(暗)', color: '#2d4a09', passable: true, layer: 0, icon: '🌱' },
  grass3: { id: 3, name: '草地(花)', color: '#4a7c12', passable: true, layer: 0, icon: '🌼' },
  dirt: { id: 4, name: '泥土', color: '#8b4513', passable: true, layer: 0, icon: '🟤' },
  sand: { id: 5, name: '沙地', color: '#c2b280', passable: true, layer: 0, icon: '🏜️' },
  sand2: { id: 6, name: '沙漠', color: '#d4a574', passable: true, layer: 0, icon: '🏖️' },
  
  // 水域
  water: { id: 7, name: '水域', color: '#1e4d8c', passable: false, layer: 0, icon: '🌊' },
  waterShallow: { id: 8, name: '浅水', color: '#3d7ab5', passable: false, layer: 0, icon: '💧' },
  waterDeep: { id: 9, name: '深水', color: '#0f2d4d', passable: false, layer: 0, icon: '🌑' },
  waterfall: { id: 10, name: '瀑布', color: '#4a90c2', passable: false, layer: 1, icon: '💦' },
  
  // 岩石/石头
  stone: { id: 11, name: '石头', color: '#4a5568', passable: false, layer: 1, icon: '🪨' },
  rock: { id: 12, name: '岩石', color: '#5c6670', passable: false, layer: 1, icon: '🗿' },
  gravel: { id: 13, name: '碎石', color: '#6b7280', passable: true, layer: 0, icon: '⚪' },
  
  // 建筑材料
  woodFloor: { id: 14, name: '木地板', color: '#8b7355', passable: true, layer: 0, icon: '🪵' },
  stoneFloor: { id: 15, name: '石板地', color: '#696969', passable: true, layer: 0, icon: '⬜' },
  brickFloor: { id: 16, name: '砖地', color: '#a0522d', passable: true, layer: 0, icon: '🧱' },
  tileFloor: { id: 17, name: '地砖', color: '#d2b48c', passable: true, layer: 0, icon: '⬛' },
  
  // 墙壁
  wallWood: { id: 18, name: '木墙', color: '#654321', passable: false, layer: 1, icon: '🧱' },
  wallStone: { id: 19, name: '石墙', color: '#5c5c5c', passable: false, layer: 1, icon: '🧱' },
  wallBrick: { id: 20, name: '砖墙', color: '#8b4513', passable: false, layer: 1, icon: '🧱' },
  wallDark: { id: 21, name: '暗墙', color: '#2d3748', passable: false, layer: 1, icon: '🧱' },
  
  // 门
  doorWood: { id: 22, name: '木门', color: '#654321', passable: false, layer: 1, icon: '🚪' },
  doorIron: { id: 23, name: '铁门', color: '#708090', passable: false, layer: 1, icon: '🚪' },
  doorOpen: { id: 24, name: '开门', color: '#8b7765', passable: true, layer: 1, icon: '🚪' },
  
  // 楼梯
  stairsUp: { id: 25, name: '楼梯(上)', color: '#a0522d', passable: true, layer: 0, icon: '🪜' },
  stairsDown: { id: 26, name: '楼梯(下)', color: '#8b4513', passable: true, layer: 0, icon: '🪜' },
  ladder: { id: 27, name: '梯子', color: '#654321', passable: true, layer: 1, icon: '🪜' },
  
  // 植物
  tree: { id: 28, name: '树木', color: '#228b22', passable: false, layer: 1, icon: '🌲' },
  treeBig: { id: 29, name: '大树', color: '#2e8b2e', passable: false, layer: 1, icon: '🌳' },
  bush: { id: 30, name: '灌木', color: '#3cb371', passable: false, layer: 1, icon: '🌿' },
  flower: { id: 31, name: '花丛', color: '#ff69b4', passable: true, layer: 0, icon: '🌸' },
  bamboo: { id: 32, name: '竹林', color: '#32cd32', passable: false, layer: 1, icon: '🎋' },
  cactus: { id: 33, name: '仙人掌', color: '#228b22', passable: false, layer: 1, icon: '🌵' },
  
  // 道路
  roadDirt: { id: 34, name: '土路', color: '#a0522d', passable: true, layer: 0, icon: '🛤️' },
  roadStone: { id: 35, name: '石板路', color: '#708090', passable: true, layer: 0, icon: '🛤️' },
  roadBrick: { id: 36, name: '砖路', color: '#cd853f', passable: true, layer: 0, icon: '🛤️' },
  bridgeWood: { id: 37, name: '木桥', color: '#8b7765', passable: true, layer: 0, icon: '🌉' },
  bridgeStone: { id: 38, name: '石桥', color: '#708090', passable: true, layer: 0, icon: '🌉' },
  
  // 特殊地形
  cave: { id: 39, name: '洞穴', color: '#1a202c', passable: true, layer: 0, icon: '🕳️' },
  lava: { id: 40, name: '岩浆', color: '#ff4500', passable: false, layer: 0, icon: '🔥' },
  ice: { id: 41, name: '冰面', color: '#87ceeb', passable: true, layer: 0, icon: '🧊' },
  snow: { id: 42, name: '雪地', color: '#f5f5f5', passable: true, layer: 0, icon: '❄️' },
  snowTree: { id: 43, name: '雪树', color: '#4682b4', passable: false, layer: 1, icon: '🎄' },
  
  // 建筑内部
  floor1: { id: 44, name: '地板A', color: '#deb887', passable: true, layer: 0, icon: '⬜' },
  floor2: { id: 45, name: '地板B', color: '#d2b48c', passable: true, layer: 0, icon: '⬜' },
  carpet: { id: 46, name: '地毯', color: '#cd5c5c', passable: true, layer: 0, icon: '🟥' },
  altar: { id: 47, name: '祭坛', color: '#8b4513', passable: false, layer: 1, icon: '⛫' },
  table: { id: 48, name: '桌子', color: '#a0522d', passable: false, layer: 1, icon: '🪑' },
  chair: { id: 49, name: '椅子', color: '#8b7355', passable: false, layer: 1, icon: '🪑' },
  
  // 水域装饰
  boat: { id: 50, name: '小船', color: '#8b4513', passable: false, layer: 1, icon: '⛵' },
  dock: { id: 51, name: '码头', color: '#a0522d', passable: true, layer: 0, icon: '🏗️' },
  beach: { id: 52, name: '沙滩', color: '#f4a460', passable: true, layer: 0, icon: '🏖️' },
  
  // 传送点
  warp: { id: 53, name: '传送点', color: '#9932cc', passable: true, layer: 0, icon: '🌀' },
  portal: { id: 54, name: '传送门', color: '#8a2be2', passable: true, layer: 1, icon: '🌀' },
  
  // 危险区域
  trap: { id: 55, name: '陷阱', color: '#dc143c', passable: true, layer: 0, icon: '⚠️' },
  poison: { id: 56, name: '毒沼', color: '#228b22', passable: true, layer: 0, icon: '☠️' },
  
  // 任务相关
  questStart: { id: 57, name: '任务起点', color: '#ffd700', passable: true, layer: 0, icon: '🚩' },
  questEnd: { id: 58, name: '任务终点', color: '#e94560', passable: true, layer: 0, icon: '🏁' },
  
  // 城市装饰
  fountain: { id: 59, name: '喷泉', color: '#4169e1', passable: false, layer: 1, icon: '⛲' },
  statue: { id: 60, name: '雕像', color: '#c0c0c0', passable: false, layer: 1, icon: '🗿' },
  lamp: { id: 61, name: '路灯', color: '#ffd700', passable: false, layer: 1, icon: '🏮' },
  
  // 自然装饰
  mountain: { id: 62, name: '山脉', color: '#696969', passable: false, layer: 1, icon: '⛰️' },
  hill: { id: 63, name: '小山', color: '#556b2f', passable: false, layer: 1, icon: '⛰️' },
  cliff: { id: 64, name: '悬崖', color: '#708090', passable: false, layer: 1, icon: '🪨' },
  
  // 空瓦片
  empty: { id: 0, name: '空白', color: '#000000', passable: true, layer: 0, icon: '⬛' },
};

/**
 * 获取所有瓦片类型列表
 */
TileDefinitions.getAll = function() {
  return Object.values(this);
};

/**
 * 根据ID获取瓦片定义
 */
TileDefinitions.getById = function(id) {
  return Object.values(this).find(t => t.id === id);
};

/**
 * 根据名称获取瓦片定义
 */
TileDefinitions.getByName = function(name) {
  return Object.values(this).find(t => t.name === name);
};

/**
 * 获取可通行瓦片
 */
TileDefinitions.getPassable = function() {
  return Object.values(this).filter(t => t.passable);
};

/**
 * 获取阻挡瓦片
 */
TileDefinitions.getBlocking = function() {
  return Object.values(this).filter(t => !t.passable);
};

/**
 * 获取指定层级的瓦片
 */
TileDefinitions.getByLayer = function(layer) {
  return Object.values(this).filter(t => t.layer === layer);
};

// 按类别分组
TileDefinitions.categories = {
  ground: [1, 2, 3, 4, 5, 6, 14, 15, 16, 17, 34, 35, 36, 44, 45, 46],
  water: [7, 8, 9, 10, 50, 51, 52],
  plants: [28, 29, 30, 31, 32, 33],
  walls: [18, 19, 20, 21],
  doors: [22, 23, 24],
  stairs: [25, 26, 27],
  roads: [34, 35, 36, 37, 38],
  special: [39, 40, 41, 42, 43, 53, 54, 55, 56, 57, 58],
  buildings: [47, 48, 49, 59, 60, 61],
  nature: [62, 63, 64, 11, 12, 13],
};

/**
 * 根据类别获取瓦片
 */
TileDefinitions.getByCategory = function(category) {
  const ids = this.categories[category];
  if (!ids) return [];
  return ids.map(id => this.getById(id)).filter(Boolean);
};

window.TileDefinitions = TileDefinitions;
