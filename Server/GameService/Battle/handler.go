package battle

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	common "game-server/Common"
	buff "game-server/GameService/Buff"
	monster "game-server/GameService/Monster"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理
type Handler struct {
	service      *Service
	gatewayURL   string // Gateway服务地址（用于推送WebSocket消息）
	gatewayMutex sync.RWMutex
	client       *http.Client // HTTP客户端（复用连接）
}

// NewHandler 创建Handler实例
func NewHandler(gatewayURL string) *Handler {
	return &Handler{
		service:    NewService(),
		gatewayURL: gatewayURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// SetGatewayURL 设置Gateway服务地址
func (h *Handler) SetGatewayURL(url string) {
	h.gatewayMutex.Lock()
	defer h.gatewayMutex.Unlock()
	h.gatewayURL = url
}

// GetGatewayURL 获取Gateway服务地址
func (h *Handler) GetGatewayURL() string {
	h.gatewayMutex.RLock()
	defer h.gatewayMutex.RUnlock()
	return h.gatewayURL
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	battleGroup := r.Group("/api/battle")
	{
		// 战斗相关
		battleGroup.POST("/attack", h.HandleAttack)   // 玩家攻击怪物
		battleGroup.POST("/pvp", h.HandlePVPAttack)   // 玩家攻击玩家（PVP）
		battleGroup.POST("/damage", h.HandleDamage)   // 处理伤害
		battleGroup.POST("/death", h.HandleDeath)     // 处理死亡
		battleGroup.POST("/respawn", h.HandleRespawn) // 处理复活
		battleGroup.POST("/levelup", h.HandleLevelUp) // 处理升级
		battleGroup.POST("/buff", h.HandleBuff)       // 处理增益
		battleGroup.POST("/debuff", h.HandleDeBuff)   // 处理减益
		battleGroup.POST("/event", h.HandleMapEvent)  // 处理地图事件
	}
}

// HandleAttack 处理玩家攻击怪物请求
func (h *Handler) HandleAttack(c *gin.Context) {
	var req AttackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	log.Printf("⚔️ 收到攻击请求: roleID=%d, monsterID=%d, skillID=%d",
		req.AttackerID, req.TargetID, req.SkillID)

	// 调用战斗服务处理攻击
	result := h.service.PlayerAttackMonster(req.AttackerID, req.TargetID, req.SkillID)

	if result == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "战斗计算失败",
			"data": nil,
		})
		return
	}

	if !result.Success && result.ErrorCode > 0 {
		// 攻击失败（距离过远、冷却中等）：通过WebSocket推送错误给攻击者
		// 前端使用WebSocket发送攻击请求，不读取HTTP响应，必须通过push通道下发错误
		errorMsg := map[string]interface{}{
			"attacker_id": req.AttackerID,
			"target_id":   req.TargetID,
			"error_code":  result.ErrorCode,
			"error_msg":   result.ErrorMsg,
		}
		h.pushToClient(req.AttackerID, "attack_failed", errorMsg)

		c.JSON(http.StatusOK, gin.H{
			"code": result.ErrorCode,
			"msg":  result.ErrorMsg,
			"data": result,
		})
		return
	}

	// 攻击成功，返回完整结果
	log.Printf("✅ 攻击结果: 怪物%d 受到%d点伤害 (暴击=%v, 闪避=%v, 死亡=%v)",
		req.TargetID, result.Damage, result.IsCrit, result.IsMiss, result.IsDead)

	// 广播战斗结果给同地图所有玩家（包括攻击者）
	// 其他玩家可以看到打怪伤害飘字，attacker_id 用于客户端区分是否是自己造成的伤害
	h.broadcastToMap(getPlayerMapID(req.AttackerID), "attack_result", result)

	// 如果怪物死亡，广播给同地图所有玩家
	if result.IsDead {
		deathNotice := map[string]interface{}{
			"monster_id": req.TargetID,
			"killer_id":  req.AttackerID,
			"exp_gain":   result.ExpGain,
			"gold_gain":  result.GoldGain,
			"drops":      result.Drops,
		}
		h.broadcastToMap(getPlayerMapID(req.AttackerID), "monster_death", deathNotice)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
}

// HandleDamage 处理伤害
func (h *Handler) HandleDamage(c *gin.Context) {
	var req DamageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理伤害逻辑
	result := h.service.ProcessDamage(req)

	// 推送给目标玩家
	h.pushToClient(req.TargetID, "damage", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandlePVPAttack 处理玩家攻击玩家请求（PVP）
func (h *Handler) HandlePVPAttack(c *gin.Context) {
	var req PVPAttackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	log.Printf("⚔️ PVP攻击请求: 攻击者=%d, 目标=%d, skillID=%d",
		req.AttackerID, req.TargetID, req.SkillID)

	// 调用PVP战斗服务
	result := h.service.PlayerAttackPlayer(req.AttackerID, req.TargetID, req.SkillID)

	// 攻击失败：推送错误给攻击者
	if !result.Success && result.ErrorCode > 0 {
		errorMsg := map[string]interface{}{
			"attacker_id": req.AttackerID,
			"target_id":   req.TargetID,
			"error_code":  result.ErrorCode,
			"error_msg":   result.ErrorMsg,
		}
		h.pushToClient(req.AttackerID, "attack_failed", errorMsg)
		c.JSON(http.StatusOK, gin.H{"code": result.ErrorCode, "msg": result.ErrorMsg, "data": result})
		return
	}

	// 推送PVP伤害结果给攻击者和目标
	// 构造前端可识别的PVP伤害消息（与CMD_DAMAGE协议一致）
	// 注：底层CalculateFinalDamage的isMiss语义为"攻击未命中（含闪避）"，
	// 系统未独立区分"未命中"与"闪避"。前端Game.js读is_dodged、BattleSystem.js读is_miss，
	// 因此两个字段都赋值result.IsMiss，确保前端两个模块都能正确识别未命中状态。
	pvpResultMsg := map[string]interface{}{
		"target_id":       req.TargetID,
		"attacker_id":     req.AttackerID,
		"attacker_type":   1, // 1=玩家
		"damage":          result.Damage,
		"is_critical":     result.IsCrit,
		"is_miss":         result.IsMiss, // 未命中（前端BattleSystem使用）
		"is_blocked":      result.IsBlocked,
		"is_dodged":       result.IsMiss, // 闪避（前端Game.js使用，语义同is_miss）
		"current_hp":      result.TargetHP,
		"max_hp":          result.TargetMaxHP,
		"is_dead":         result.IsDead,
		"is_skill_attack": result.IsSkillAttack,
		"skill_name":      result.SkillName,
		"buff_applied":    result.BuffApplied,
		"current_mp":      result.AttackerMP,
		"max_mp":          result.AttackerMaxMP,
		"exp_gain":        result.ExpGain,
		"pk_value_gain":   result.PkValueGain,
		"is_pvp":          true,
	}

	// 推送给攻击者（作为攻击结果）
	h.pushToClient(req.AttackerID, "attack_result", pvpResultMsg)
	// 推送给目标（作为受击结果）
	h.pushToClient(req.TargetID, "damage", pvpResultMsg)

	// 如果目标死亡，广播给同地图所有玩家
	if result.IsDead {
		deathNotice := map[string]interface{}{
			"victim_id":     req.TargetID,
			"killer_id":     req.AttackerID,
			"exp_gain":      result.ExpGain,
			"pk_value_gain": result.PkValueGain,
			"is_pvp":        true,
		}
		h.broadcastToMap(getPlayerMapID(req.AttackerID), "player_death", deathNotice)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleDeath 处理死亡
func (h *Handler) HandleDeath(c *gin.Context) {
	var req DeathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理死亡逻辑
	result := h.service.ProcessDeath(req)

	// 推送给目标玩家
	h.pushToClient(req.TargetID, "death", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleRespawn 处理复活
func (h *Handler) HandleRespawn(c *gin.Context) {
	var req RespawnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理复活逻辑
	result := h.service.ProcessRespawn(req)

	// 推送给玩家
	h.pushToClient(req.RoleID, "respawn", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleLevelUp 处理升级
func (h *Handler) HandleLevelUp(c *gin.Context) {
	var req LevelUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理升级逻辑
	result := h.service.ProcessLevelUp(req)

	// 推送给玩家
	h.pushToClient(req.RoleID, "levelup", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleBuff 处理增益
func (h *Handler) HandleBuff(c *gin.Context) {
	var req BuffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理增益逻辑
	result := h.service.ProcessBuff(req)

	// 推送给玩家
	h.pushToClient(req.TargetID, "buff", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleDeBuff 处理减益
func (h *Handler) HandleDeBuff(c *gin.Context) {
	var req DeBuffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理减益逻辑
	result := h.service.ProcessDeBuff(req)

	// 推送给玩家
	h.pushToClient(req.TargetID, "debuff", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleMapEvent 处理地图事件
func (h *Handler) HandleMapEvent(c *gin.Context) {
	var req MapEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理地图事件逻辑
	result := h.service.ProcessMapEvent(req)

	// 广播给地图所有玩家
	h.broadcastToMap(req.MapID, "mapevent", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// pushToClient 推送消息给客户端（通过Gateway）
func (h *Handler) pushToClient(roleID uint64, msgType string, data interface{}) {
	gatewayURL := h.GetGatewayURL()
	if gatewayURL == "" {
		log.Printf("Gateway URL未配置，跳过推送")
		return
	}

	// 构建推送请求
	pushReq := map[string]interface{}{
		"role_id":  roleID,
		"msg_type": msgType,
		"data":     data,
	}

	jsonData, err := json.Marshal(pushReq)
	if err != nil {
		log.Printf("序列化推送数据失败: %v", err)
		return
	}

	// 异步发送HTTP请求到Gateway（复用client）
	go func() {
		resp, err := h.client.Post(gatewayURL+"/internal/push", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("推送消息到Gateway失败: %v", err)
			return
		}
		defer resp.Body.Close()
	}()
}

// PushMonsterAttackResult 推送怪物攻击玩家的结果给同地图所有客户端
// 供main.go注入到monster包使用（避免monster包依赖battle包）
// 广播给同地图所有玩家，让其他玩家也能看到受击飘字和血条
func (h *Handler) PushMonsterAttackResult(mapID uint32, targetRoleID uint64, monsterName string, result *common.MonsterAttackResult) {
	if result == nil {
		return
	}

	// 构造前端可识别的伤害消息（与CMD_DAMAGE协议一致）
	// 注：底层CalculateFinalDamage的isMiss语义为"攻击未命中（含闪避）"，
	// 系统未独立区分"未命中"与"闪避"。前端Game.js读is_dodged、BattleSystem.js读is_miss，
	// 因此两个字段都赋值result.IsMiss，确保前端两个模块都能正确识别未命中状态。
	damageMsg := map[string]interface{}{
		"target_id":     targetRoleID,
		"attacker_id":   result.MonsterID,
		"attacker_name": monsterName,
		"attacker_type": 2, // 2=怪物
		"damage":        result.Damage,
		"is_critical":   result.IsCrit,
		"is_miss":       result.IsMiss, // 未命中（前端BattleSystem使用）
		"is_blocked":    false,
		"is_dodged":     result.IsMiss, // 闪避（前端Game.js使用，语义同is_miss）
		"current_hp":    result.PlayerHP,
		"max_hp":        result.PlayerMaxHP,
		"is_dead":       result.IsDead,
	}

	// 广播给同地图所有玩家（包括被攻击者），让所有人看到受击飘字和血条
	h.broadcastToMap(mapID, "damage", damageMsg)
}

// broadcastToMap 广播消息给地图所有玩家（通过Gateway）
func (h *Handler) broadcastToMap(mapID uint32, msgType string, data interface{}) {
	gatewayURL := h.GetGatewayURL()
	if gatewayURL == "" {
		log.Printf("Gateway URL未配置，跳过广播")
		return
	}

	// 构建广播请求
	broadcastReq := map[string]interface{}{
		"map_id":   mapID,
		"msg_type": msgType,
		"data":     data,
	}

	jsonData, err := json.Marshal(broadcastReq)
	if err != nil {
		log.Printf("序列化广播数据失败: %v", err)
		return
	}

	// 异步发送HTTP请求到Gateway（复用client）
	go func() {
		resp, err := h.client.Post(gatewayURL+"/internal/broadcast", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("广播消息到Gateway失败: %v", err)
			return
		}
		defer resp.Body.Close()
	}()
}

// BroadcastMonsterPositions 广播怪物位置更新给地图所有玩家
func (h *Handler) BroadcastMonsterPositions(mapID uint32, data interface{}) {
	log.Printf("📡 广播怪物位置: 地图%d, %d个怪物", mapID, len(data.(gin.H)["monsters"].([]monster.MonsterPositionInfo)))
	h.broadcastToMap(mapID, "monster_position_update", data)
}

// BroadcastMonsterSpawn 广播怪物生成（复活）消息给地图所有玩家
// 客户端收到后会重新创建怪物对象
func (h *Handler) BroadcastMonsterSpawn(mapID uint32, spawnInfo monster.MonsterSpawnInfo) {
	h.broadcastToMap(mapID, "monster_spawn", spawnInfo)
}

// BroadcastMonsterPositionsBinary 广播怪物位置（二进制协议，紧凑格式）
// 二进制格式: [map_id:4B][count:2B][timestamp:8B] + 每个怪物[instance_id:4B][x:2B][y:2B][state:1B][hp:4B]
func (h *Handler) BroadcastMonsterPositionsBinary(mapID uint32, positions []monster.MonsterPositionInfo) {
	if len(positions) == 0 {
		return
	}

	// 编码二进制数据
	body := EncodeMonsterPositions(mapID, positions)

	// 优先通过消息总线发送（RabbitMQ，自动降级到HTTP）
	// 消息总线不可用时，回退到直连Gateway的HTTP二进制接口
	if common.GlobalMessageBus != nil && common.GlobalMessageBus.IsAvailable() {
		if err := common.PublishMonsterPositionBinary(mapID, 3101, body); err != nil {
			log.Printf("消息总线发送怪物位置失败，回退直连HTTP: %v", err)
			h.broadcastToMapBinary(mapID, 3101, body)
		}
	} else {
		// 消息总线不可用，直接走HTTP二进制接口
		h.broadcastToMapBinary(mapID, 3101, body)
	}
}

// EncodeMonsterPositions 编码怪物位置为二进制格式
// 格式: [map_id:4B uint32][count:2B uint16][timestamp:8B int64] + 每个怪物13字节
func EncodeMonsterPositions(mapID uint32, positions []monster.MonsterPositionInfo) []byte {
	const headerSize = 14 // 4 + 2 + 8
	const perMonster = 13 // 4 + 2 + 2 + 1 + 4
	buf := make([]byte, headerSize+len(positions)*perMonster)

	offset := 0
	// map_id (uint32 小端)
	binary.LittleEndian.PutUint32(buf[offset:], mapID)
	offset += 4
	// count (uint16 小端)
	binary.LittleEndian.PutUint16(buf[offset:], uint16(len(positions)))
	offset += 2
	// timestamp (int64 小端)
	binary.LittleEndian.PutUint64(buf[offset:], uint64(time.Now().UnixMilli()))
	offset += 8

	// 每个怪物
	for _, p := range positions {
		// instance_id (uint32 小端，怪物实例ID不会超过42亿)
		binary.LittleEndian.PutUint32(buf[offset:], uint32(p.InstanceID))
		offset += 4
		// x (uint16 小端)
		binary.LittleEndian.PutUint16(buf[offset:], uint16(p.X))
		offset += 2
		// y (uint16 小端)
		binary.LittleEndian.PutUint16(buf[offset:], uint16(p.Y))
		offset += 2
		// state (uint8)
		buf[offset] = byte(p.State)
		offset += 1
		// hp (int32 小端)
		binary.LittleEndian.PutUint32(buf[offset:], uint32(p.HP))
		offset += 4
	}

	return buf
}

// broadcastToMapBinary 通过二进制接口广播消息给地图所有玩家
// Gateway收到后直接透传 [cmd][body] 给客户端，零拷贝
func (h *Handler) broadcastToMapBinary(mapID uint32, cmd uint16, body []byte) {
	gatewayURL := h.GetGatewayURL()
	if gatewayURL == "" {
		return
	}

	// 异步发送二进制数据到Gateway的专用接口
	go func() {
		url := gatewayURL + "/internal/broadcast_binary"
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("构建二进制广播请求失败: %v", err)
			return
		}
		// 通过Header传递元数据，body是纯二进制
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Map-Id", strconv.FormatUint(uint64(mapID), 10))
		req.Header.Set("X-Cmd", strconv.FormatUint(uint64(cmd), 10))

		resp, err := h.client.Do(req)
		if err != nil {
			log.Printf("二进制广播到Gateway失败: %v", err)
			return
		}
		defer resp.Body.Close()
	}()
}

// getPlayerMapID 获取玩家当前地图ID
func getPlayerMapID(roleID uint64) uint32 {
	pos, err := common.DBGetRolePosition(roleID)
	if err != nil || pos == nil {
		return 1 // 默认返回新手村
	}
	return pos.MapID
}

// ========== BuffTickApplier 接口实现 ==========
// 让 Handler 实现 buff.BuffTickApplier 接口，用于处理BUFF持续效果

// ApplyPlayerHPChange 修改玩家HP（正=恢复, 负=伤害）
func (h *Handler) ApplyPlayerHPChange(playerID uint64, hpChange int) {
	newHP, err := common.DBRoleChangeHP(playerID, hpChange)
	if err != nil {
		log.Printf("[BUFF-TICK] 修改玩家[%d]HP失败: %v", playerID, err)
		return
	}
	log.Printf("[BUFF-TICK] 玩家[%d] HP变化: %d (当前HP: %d)", playerID, hpChange, newHP)
}

// ApplyPlayerMPChange 修改玩家MP（正=恢复, 负=消耗）
func (h *Handler) ApplyPlayerMPChange(playerID uint64, mpChange int) {
	newMP, err := common.DBRoleChangeMP(playerID, mpChange)
	if err != nil {
		log.Printf("[BUFF-TICK] 修改玩家[%d]MP失败: %v", playerID, err)
		return
	}
	log.Printf("[BUFF-TICK] 玩家[%d] MP变化: %d (当前MP: %d)", playerID, mpChange, newMP)
}

// ApplyMonsterHPChange 修改怪物HP（正=恢复, 负=伤害）
// 返回当前HP和是否死亡
func (h *Handler) ApplyMonsterHPChange(monsterID uint64, hpChange int) (currentHP int, isDead bool) {
	// 通过怪物服务修改HP
	if globalMonsterSvc == nil {
		log.Printf("[BUFF-TICK] 怪物服务未注入，无法修改怪物[%d]HP", monsterID)
		return 0, false
	}

	currentHP, isDead = globalMonsterSvc.MonsterTakeDamage(monsterID, -hpChange) // MonsterTakeDamage接收正数伤害
	if isDead {
		log.Printf("[BUFF-TICK] 怪物[%d]被BUFF效果击杀 (HP变化: %d)", monsterID, hpChange)
	} else {
		log.Printf("[BUFF-TICK] 怪物[%d] HP变化: %d (当前HP: %d)", monsterID, hpChange, currentHP)
	}
	return
}

// BroadcastBuffTick 推送BUFF Tick结果到客户端（飘字显示）
func (h *Handler) BroadcastBuffTick(targetID uint64, targetType uint8, hpChange int, mpChange int) {
	if hpChange == 0 && mpChange == 0 {
		return
	}

	// 构造BUFF tick消息
	tickMsg := map[string]interface{}{
		"target_id":   targetID,
		"target_type": targetType,
		"hp_change":   hpChange,
		"mp_change":   mpChange,
		"tick_type":   "buff", // 标识为BUFF tick
	}

	// 根据目标类型获取地图ID进行广播
	var mapID uint32
	if targetType == 1 { // 玩家
		mapID = getPlayerMapID(targetID)
	} else if targetType == 2 { // 怪物
		// 尝试从怪物服务获取地图ID
		if globalMonsterSvc != nil {
			if info, ok := globalMonsterSvc.GetMonsterInfo(targetID); ok {
				mapID = info.MapID
			}
		}
	}

	if mapID > 0 {
		h.broadcastToMap(mapID, "buff_tick", tickMsg)
		log.Printf("[BUFF-TICK] 广播: 地图%d, 目标[%d](type=%d), HP:%d, MP:%d", mapID, targetID, targetType, hpChange, mpChange)
	}
}

// BroadcastBuffTickBatch 批量推送BUFF Tick结果（性能优化版本）
// 按地图ID分组后批量广播，减少HTTP请求次数
func (h *Handler) BroadcastBuffTickBatch(results []buff.BuffTickResult) {
	if len(results) == 0 {
		return
	}

	// 按地图ID分组
	type MapBatch struct {
		MapID uint32
		Ticks []map[string]interface{}
	}

	mapBatches := make(map[uint32]*MapBatch)

	for _, result := range results {
		var mapID uint32

		// 获取目标所在地图
		if result.TargetType == 1 { // 玩家
			mapID = getPlayerMapID(result.TargetID)
		} else if result.TargetType == 2 { // 怪物
			if globalMonsterSvc != nil {
				if info, ok := globalMonsterSvc.GetMonsterInfo(result.TargetID); ok {
					mapID = info.MapID
				}
			}
		}

		if mapID > 0 {
			if _, exists := mapBatches[mapID]; !exists {
				mapBatches[mapID] = &MapBatch{
					MapID: mapID,
					Ticks: []map[string]interface{}{},
				}
			}

			tickMsg := map[string]interface{}{
				"target_id":   result.TargetID,
				"target_type": result.TargetType,
				"hp_change":   result.HpChange,
				"mp_change":   result.MpChange,
				"tick_type":   "buff_batch",
			}

			mapBatches[mapID].Ticks = append(mapBatches[mapID].Ticks, tickMsg)
		}
	}

	// 批量广播每个地图的tick结果
	for mapID, batch := range mapBatches {
		if len(batch.Ticks) == 0 {
			continue
		}

		batchMsg := map[string]interface{}{
			"map_id":     mapID,
			"tick_count": len(batch.Ticks),
			"ticks":      batch.Ticks,
			"timestamp":  time.Now().Unix(),
		}

		h.broadcastToMap(mapID, "buff_tick_batch", batchMsg)
		log.Printf("[BUFF-TICK-BATCH] 地图%d, 批量广播%d个目标", mapID, len(batch.Ticks))
	}
}
