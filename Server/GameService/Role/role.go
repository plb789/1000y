package role

import (
	gameMap "game-server/GameService/Map"
	"sync"
)

type Player struct {
	UID   uint
	TileX int
	TileY int
}

var OnlinePlayers = sync.Map{}

// MoveCheck 角色移动校验
func MoveCheck(uid uint, targetX, targetY int) bool {
	if !gameMap.CanMove(targetX, targetY) {
		return false
	}
	val, ok := OnlinePlayers.Load(uid)
	if !ok {
		return false
	}
	p := val.(*Player)
	p.TileX = targetX
	p.TileY = targetY
	return true
}