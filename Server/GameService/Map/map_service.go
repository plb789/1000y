package gamemap

import (
	"errors"
	common "game-server/Common"
	"log"
	"sync"
)

// 全局地图服务实例
var globalService *Service

// GetService 获取全局地图服务实例
func GetService() *Service {
	if globalService == nil {
		globalService = NewService()
	}
	return globalService
}

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
	config := common.GetMapConfig(mapID)
	if config == nil {
		return nil, errors.New("地图不存在")
	}
	mapBase := MapBase{
		ID:          config.ID,
		Name:        config.Name,
		Width:       config.Width,
		Height:      config.Height,
		TileWidth:   config.TileWidth,
		TileHeight:  config.TileHeight,
		MapFile:     config.MapFile,
		TilesetFile: config.TilesetFile,
		Music:       config.Music,
		PkAllowed:   config.PkAllowed,
		ReviveMapID: config.ReviveMapID,
		ReviveX:     config.ReviveX,
		ReviveY:     config.ReviveY,
		LevelReq:    config.LevelReq,
		MinimapFile: config.MinimapFile,
	}
	return &mapBase, nil
}

// GetAllMaps 获取所有地图
func (s *Service) GetAllMaps() ([]MapBrief, error) {
	var maps []MapBrief
	for _, m := range common.GameConfig.Maps {
		maps = append(maps, MapBrief{
			ID:        m.ID,
			Name:      m.Name,
			Width:     m.Width,
			Height:    m.Height,
			LevelReq:  m.LevelReq,
			PkAllowed: m.PkAllowed,
		})
	}
	return maps, nil
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

	// 从全局地图数据获取碰撞数据
	var collision [][]bool
	gameMap := GetGameMap(mapBase.MapFile)
	if gameMap != nil && gameMap.Collision != nil {
		collision = gameMap.Collision
		log.Printf("LoadMap: 使用地图文件 %s 的碰撞数据，尺寸=%dx%d",
			mapBase.MapFile, len(collision[0]), len(collision))
	} else {
		// 如果地图文件未加载，创建默认碰撞表
		tileWidth := mapBase.TileWidth
		tileHeight := mapBase.TileHeight
		if tileWidth <= 0 {
			tileWidth = 48
		}
		if tileHeight <= 0 {
			tileHeight = 48
		}

		tileCountX := mapBase.Width / tileWidth
		tileCountY := mapBase.Height / tileHeight
		if tileCountX <= 0 {
			tileCountX = 100
		}
		if tileCountY <= 0 {
			tileCountY = 100
		}

		collision = make([][]bool, tileCountY)
		for y := 0; y < tileCountY; y++ {
			collision[y] = make([]bool, tileCountX)
			for x := 0; x < tileCountX; x++ {
				collision[y][x] = false
			}
		}
		log.Printf("LoadMap: 地图文件 %s 未加载，创建默认碰撞数据，尺寸=%dx%d",
			mapBase.MapFile, tileCountX, tileCountY)
	}

	// 创建地图实例
	lm := &LoadedMap{
		MapData:    mapBase,
		TileWidth:  mapBase.TileWidth,
		TileHeight: mapBase.TileHeight,
		Collision:  collision,
		Monsters:   make(map[uint64]*MonsterInstance),
		NPCs:       make(map[uint64]*NPCInstance),
		Players:    make(map[uint64]*PlayerInstance),
	}

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
	configNPCs := common.GetNPCsByMap(mapID)
	var npcs []NPCBase
	for _, n := range configNPCs {
		npcs = append(npcs, NPCBase{
			ID:         n.ID,
			Name:       n.Name,
			Type:       n.Type,
			MapID:      n.MapID,
			X:          n.X,
			Y:          n.Y,
			Face:       n.Face,
			SpriteID:   n.SpriteID,
			DialogText: n.DialogText,
			ShopID:     n.ShopID,
		})
	}
	return npcs, nil
}

// GetMonstersByMap 获取地图上的怪物列表
func (s *Service) GetMonstersByMap(mapID uint32) ([]MonsterBase, error) {
	var monsters []MonsterBase
	for _, m := range common.GameConfig.Monsters {
		if m.MapID == mapID {
			monsters = append(monsters, MonsterBase{
				ID:          m.ID,
				Name:        m.Name,
				Level:       m.Level,
				Type:        m.Type,
				MapID:       m.MapID,
				Hp:          m.Hp,
				Attack:      m.Attack,
				Defense:     m.Defense,
				Speed:       m.Speed,
				Hit:         m.Hit,
				Dodge:       m.Dodge,
				Crit:        m.Crit,
				AIType:      m.AIType,
				AttackRange: m.AttackRange,
				ChaseRange:  m.ChaseRange,
				GoldMin:     m.GoldMin,
				GoldMax:     m.GoldMax,
				Exp:         m.Exp,
				DropGroupID: m.DropGroupID,
				SpriteID:    m.SpriteID,
				RespawnTime: m.RespawnTime,
			})
		}
	}
	return monsters, nil
}

// EnterMap 角色进入地图
// 注意：客户端发送的是瓦片坐标，直接使用即可
func (s *Service) EnterMap(roleID uint64, mapID uint32, tileX, tileY int) error {
	lm, err := s.LoadMap(mapID)
	if err != nil {
		return err
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 优先使用数据库中的位置
	actualX, actualY := tileX, tileY
	position, err := common.DBGetRolePosition(roleID)
	if err == nil && position != nil {
		log.Printf("EnterMap: 从数据库读取角色 %d 位置: mapID=%d, x=%d, y=%d", roleID, position.MapID, position.X, position.Y)

		if position.MapID == mapID {
			if !s.isPositionBlockedLocked(lm, position.X, position.Y) {
				// 数据库位置可通行
				actualX, actualY = position.X, position.Y
			} else {
				// 数据库位置被阻挡，找附近可通行位置
				found := false
				for dx := -3; dx <= 3 && !found; dx++ {
					for dy := -3; dy <= 3 && !found; dy++ {
						if !s.isPositionBlockedLocked(lm, position.X+dx, position.Y+dy) {
							actualX, actualY = position.X+dx, position.Y+dy
							found = true
						}
					}
				}
				if !found {
					actualX, actualY = tileX, tileY
				}
				log.Printf("EnterMap: 数据库位置被阻挡，调整到 (%d, %d)", actualX, actualY)
			}
		} else {
			log.Printf("EnterMap: 数据库地图ID(%d)与请求地图ID(%d)不匹配，使用客户端坐标", position.MapID, mapID)
		}
	} else {
		log.Printf("EnterMap: 无法获取角色 %d 数据库位置，使用客户端坐标", roleID)
	}

	// 安全检查：确保 Collision 数组已初始化
	if len(lm.Collision) == 0 || len(lm.Collision[0]) == 0 {
		log.Printf("EnterMap: 地图 %d 的 Collision 数组未初始化，使用默认碰撞", mapID)
	} else {
		collisionWidth := len(lm.Collision[0])
		collisionHeight := len(lm.Collision)

		// 检查位置是否可通过
		if actualY >= 0 && actualY < collisionHeight && actualX >= 0 && actualX < collisionWidth {
			if lm.Collision[actualY][actualX] {
				return errors.New("该位置不可通过")
			}
		}
	}

	// 添加玩家到地图
	lm.Players[roleID] = &PlayerInstance{
		RoleID: roleID,
		X:      actualX,
		Y:      actualY,
	}

	log.Printf("EnterMap: 玩家 %d 进入地图 %d (tileX=%d, tileY=%d)", roleID, mapID, actualX, actualY)
	return nil
}

// isPositionBlockedLocked 在已持有锁的情况下检查位置是否阻挡
func (s *Service) isPositionBlockedLocked(lm *LoadedMap, x, y int) bool {
	if len(lm.Collision) == 0 || len(lm.Collision[0]) == 0 {
		return false // 未初始化时默认允许通过
	}

	collisionWidth := len(lm.Collision[0])
	collisionHeight := len(lm.Collision)

	if y < 0 || y >= collisionHeight || x < 0 || x >= collisionWidth {
		return true // 超出边界视为阻挡
	}

	return lm.Collision[y][x]
}

// LeaveMap 角色离开地图
func (s *Service) LeaveMap(roleID uint64, mapID uint32) error {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		return nil // 地图未加载,无需处理
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 获取玩家离开前的位置
	player, exists := lm.Players[roleID]
	if exists {
		// 离开地图时保存玩家位置到数据库
		common.DBRoleChangePosition(roleID, mapID, player.X, player.Y)
		log.Printf("LeaveMap: 玩家 %d 离开地图 %d，保存位置 (%d, %d)", roleID, mapID, player.X, player.Y)
	}

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

// MovePlayerResult 移动结果
type MovePlayerResult struct {
	Success bool
	X       int
	Y       int
}

// MovePlayer 移动玩家
func (s *Service) MovePlayer(roleID uint64, mapID uint32, x, y int) (*MovePlayerResult, error) {
	// 检查位置是否合法
	if s.IsPositionBlocked(mapID, x, y) {
		return &MovePlayerResult{Success: false, X: x, Y: y}, errors.New("位置被阻挡")
	}

	// 更新玩家位置
	if err := s.UpdatePlayerPosition(roleID, mapID, x, y); err != nil {
		// 如果玩家不在地图中，尝试先进入地图
		if err.Error() == "玩家不在此地图" {
			if enterErr := s.EnterMap(roleID, mapID, x, y); enterErr != nil {
				return &MovePlayerResult{Success: false, X: x, Y: y}, enterErr
			}
		} else {
			return &MovePlayerResult{Success: false, X: x, Y: y}, err
		}
	}

	// 注意：不再每次移动都写数据库，只在离开地图时保存

	return &MovePlayerResult{Success: true, X: x, Y: y}, nil
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
// 注意：客户端发送的是瓦片坐标，直接使用即可
func (s *Service) IsPositionBlocked(mapID uint32, tileX, tileY int) bool {
	lm, ok := s.GetLoadedMap(mapID)
	if !ok {
		log.Printf("IsPositionBlocked: 地图 %d 未加载", mapID)
		return true
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// 安全检查：确保 Collision 数组已初始化
	if len(lm.Collision) == 0 || len(lm.Collision[0]) == 0 {
		log.Printf("IsPositionBlocked: 地图 %d 的 Collision 数组未初始化", mapID)
		return false // 未初始化时默认允许通过
	}

	collisionWidth := len(lm.Collision[0])
	collisionHeight := len(lm.Collision)

	log.Printf("IsPositionBlocked: mapID=%d, tile=(%d,%d), collisionSize=%dx%d",
		mapID, tileX, tileY, collisionWidth, collisionHeight)

	if tileY < 0 || tileY >= collisionHeight || tileX < 0 || tileX >= collisionWidth {
		log.Printf("IsPositionBlocked: 坐标超出范围, tileX=%d, tileY=%d, collisionSize=%dx%d",
			tileX, tileY, collisionWidth, collisionHeight)
		return true
	}

	blocked := lm.Collision[tileY][tileX]
	if blocked {
		log.Printf("IsPositionBlocked: 位置(%d,%d)被阻挡", tileX, tileY)
	}

	return blocked
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
