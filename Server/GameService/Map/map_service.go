package gamemap

import (
	"errors"
	"game-server/DBService/mysql"
	"sync"
)

// LoadedMap 已加载的地图
type LoadedMap struct {
	MapData    *MapBase
	TileWidth  int
	TileHeight int
	Collision  [][]bool                    // 碰撞检测表
	Monsters   map[uint64]*MonsterInstance // key=怪物实例ID
	NPCs       map[uint64]*NPCInstance     // key=NPC实例ID
	Players    map[uint64]*PlayerInstance  // key=角色ID
	mu         sync.RWMutex
}

// MonsterInstance 怪物实例
type MonsterInstance struct {
	BaseID      uint64
	MonsterBase *MonsterBase
	X           int
	Y           int
	CurrentHP   int
	Status      uint8  // 0=存活, 1=战斗, 2=死亡
	TargetID    uint64 // 攻击目标
	RespawnAt   int64  // 复活时间戳
}

// NPCInstance NPC实例
type NPCInstance struct {
	BaseID  uint64
	NPCBase *NPCBase
	X       int
	Y       int
	Facing  uint8
}

// PlayerInstance 地图上的玩家实例
type PlayerInstance struct {
	RoleID uint64
	X      int
	Y      int
}

// Service 地图服务
type Service struct {
	loadedMaps map[uint32]*LoadedMap // key=mapID
	mu         sync.RWMutex
}

// NewService 创建地图服务实例
func NewService() *Service {
	return &Service{
		loadedMaps: make(map[uint32]*LoadedMap),
	}
}

// GetMapBase 获取地图基础信息
func (s *Service) GetMapBase(mapID uint32) (*MapBase, error) {
	var mapBase MapBase
	if err := mysql.DB.Where("id = ?", mapID).First(&mapBase).Error; err != nil {
		return nil, err
	}
	return &mapBase, nil
}

// GetAllMaps 获取所有地图
func (s *Service) GetAllMaps() ([]MapBrief, error) {
	var maps []MapBrief
	err := mysql.DB.Model(&MapBase{}).
		Select("id, name, width, height, level_req, pk_allowed").
		Order("id ASC").
		Find(&maps).Error
	return maps, err
}

// LoadMap 加载地图到内存
func (s *Service) LoadMap(mapID uint32) (*LoadedMap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已加载
	if lm, ok := s.loadedMaps[mapID]; ok {
		return lm, nil
	}

	// 获取地图基础信息
	mapBase, err := s.GetMapBase(mapID)
	if err != nil {
		return nil, errors.New("地图不存在")
	}

	// 创建地图实例
	lm := &LoadedMap{
		MapData:    mapBase,
		TileWidth:  mapBase.TileWidth,
		TileHeight: mapBase.TileHeight,
		Collision:  make([][]bool, mapBase.Height/mapBase.TileHeight),
		Monsters:   make(map[uint64]*MonsterInstance),
		NPCs:       make(map[uint64]*NPCInstance),
		Players:    make(map[uint64]*PlayerInstance),
	}

	// 初始化碰撞表(默认全部可通过)
	for y := 0; y < len(lm.Collision); y++ {
		lm.Collision[y] = make([]bool, mapBase.Width/mapBase.TileWidth)
		for x := 0; x < len(lm.Collision[y]); x++ {
			lm.Collision[y][x] = false // false=可通过
		}
	}

	// 加载地图文件(如果存在)
	// 实际项目中应该调用gameMap.LoadMapFile

	s.loadedMaps[mapID] = lm
	return lm, nil
}

// GetLoadedMap 获取已加载的地图
func (s *Service) GetLoadedMap(mapID uint32) (*LoadedMap, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lm, ok := s.loadedMaps[mapID]
	return lm, ok
}

// CanMove 检查移动是否允许
func (s *Service) CanMove(mapID uint32, tileX, tileY int) bool {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return false
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// 检查边界
	if tileY < 0 || tileY >= len(lm.Collision) || tileX < 0 || tileX >= len(lm.Collision[0]) {
		return false
	}

	return !lm.Collision[tileY][tileX]
}

// SetCollision 设置碰撞
func (s *Service) SetCollision(mapID uint32, tileX, tileY int, blocked bool) error {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return errors.New("地图未加载")
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	if tileY < 0 || tileY >= len(lm.Collision) || tileX < 0 || tileX >= len(lm.Collision[0]) {
		return errors.New("坐标超出范围")
	}

	lm.Collision[tileY][tileX] = blocked
	return nil
}

// GetNPCsByMap 获取地图上的NPC列表
func (s *Service) GetNPCsByMap(mapID uint32) ([]NPCBase, error) {
	var npcs []NPCBase
	err := mysql.DB.Where("map_id = ?", mapID).Find(&npcs).Error
	return npcs, err
}

// GetMonstersByMap 获取地图上的怪物列表
func (s *Service) GetMonstersByMap(mapID uint32) ([]MonsterBase, error) {
	var monsters []MonsterBase
	err := mysql.DB.Where("map_id = ?", mapID).Find(&monsters).Error
	return monsters, err
}

// EnterMap 角色进入地图
func (s *Service) EnterMap(roleID uint64, mapID uint32, x, y int) error {
	lm, err := s.LoadMap(mapID)
	if err != nil {
		return err
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 检查位置是否可通过
	tileX := x / lm.TileWidth
	tileY := y / lm.TileHeight
	if tileY >= 0 && tileY < len(lm.Collision) && tileX >= 0 && tileX < len(lm.Collision[0]) {
		if lm.Collision[tileY][tileX] {
			return errors.New("该位置不可通过")
		}
	}

	// 添加玩家到地图
	lm.Players[roleID] = &PlayerInstance{
		RoleID: roleID,
		X:      x,
		Y:      y,
	}

	return nil
}

// LeaveMap 角色离开地图
func (s *Service) LeaveMap(roleID uint64, mapID uint32) error {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return nil // 地图未加载,无需处理
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	delete(lm.Players, roleID)
	return nil
}

// UpdatePlayerPosition 更新玩家位置
func (s *Service) UpdatePlayerPosition(roleID uint64, mapID uint32, x, y int) error {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return errors.New("地图未加载")
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	if player, exists := lm.Players[roleID]; exists {
		player.X = x
		player.Y = y
		return nil
	}

	return errors.New("玩家不在此地图")
}

// GetPlayersInView 获取视野范围内的玩家
func (s *Service) GetPlayersInView(mapID uint32, centerX, centerY, viewRange int) []PlayerInstance {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return nil
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var result []PlayerInstance
	for _, player := range lm.Players {
		dx := player.X - centerX
		dy := player.Y - centerY
		if dx*dx+dy*dy <= viewRange*viewRange {
			result = append(result, *player)
		}
	}
	return result
}

// GetMapInfo 获取地图信息(包含在线玩家数)
func (s *Service) GetMapInfo(mapID uint32) (*MapBase, int, error) {
	mapBase, err := s.GetMapBase(mapID)
	if err != nil {
		return nil, 0, err
	}

	lm, ok := s.GetLoadedMap(mapID)
	playerCount := 0
	if ok {
		lm.mu.RLock()
		playerCount = len(lm.Players)
		lm.mu.RUnlock()
	}

	return mapBase, playerCount, nil
}

// TeleportPlayer 传送玩家
func (s *Service) TeleportPlayer(roleID uint64, fromMapID, toMapID uint32, x, y int) error {
	// 先离开原地图
	if err := s.LeaveMap(roleID, fromMapID); err != nil {
		return err
	}
	// 进入新地图
	return s.EnterMap(roleID, toMapID, x, y)
}

// IsPositionBlocked 检查位置是否阻挡
func (s *Service) IsPositionBlocked(mapID uint32, pixelX, pixelY int) bool {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return true
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	tileX := pixelX / lm.TileWidth
	tileY := pixelY / lm.TileHeight

	if tileY < 0 || tileY >= len(lm.Collision) || tileX < 0 || tileX >= len(lm.Collision[0]) {
		return true
	}

	return lm.Collision[tileY][tileX]
}

// GetPlayerCountOnMap 获取地图上的玩家数量
func (s *Service) GetPlayerCountOnMap(mapID uint32) int {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return 0
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return len(lm.Players)
}
