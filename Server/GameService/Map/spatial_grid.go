package gamemap

// SpatialGrid 空间网格索引
// 将地图划分为固定大小的格子，每个格子维护该区域内的实体ID
// 用于加速碰撞检测：从 O(N) 遍历降为 O(1) 查询附近格子
//
// 使用场景：
//   - 玩家移动时检查碰撞（CheckEntityCollision）
//   - 怪物 AI 寻找最近目标
//   - 视野查询（GetPlayersInView）
//
// 线程安全说明：
//   - 所有方法都不是线程安全的，调用方必须持有 LoadedMap.mu 锁
type SpatialGrid struct {
	cellSize int               // 每个格子的边长（瓦片数）
	cols     int               // 网格列数
	rows     int               // 网格行数
	cells    []map[uint64]bool // 每个格子内的实体ID集合（key=roleID）
}

// NewSpatialGrid 创建空间网格
// cellSize 建议 8-16，太小会浪费内存，太大失去分区意义
func NewSpatialGrid(width, height, cellSize int) *SpatialGrid {
	if cellSize <= 0 {
		cellSize = 10
	}
	cols := (width + cellSize - 1) / cellSize
	rows := (height + cellSize - 1) / cellSize
	if cols <= 0 {
		cols = 1
	}
	if rows <= 0 {
		rows = 1
	}
	return &SpatialGrid{
		cellSize: cellSize,
		cols:     cols,
		rows:     rows,
		cells:    make([]map[uint64]bool, cols*rows),
	}
}

// cellIndex 计算坐标对应的格子索引
func (g *SpatialGrid) cellIndex(x, y int) int {
	cx := x / g.cellSize
	cy := y / g.cellSize
	if cx < 0 {
		cx = 0
	} else if cx >= g.cols {
		cx = g.cols - 1
	}
	if cy < 0 {
		cy = 0
	} else if cy >= g.rows {
		cy = g.rows - 1
	}
	return cy*g.cols + cx
}

// Add 添加实体到网格
func (g *SpatialGrid) Add(id uint64, x, y int) {
	idx := g.cellIndex(x, y)
	if g.cells[idx] == nil {
		g.cells[idx] = make(map[uint64]bool)
	}
	g.cells[idx][id] = true
}

// Remove 从网格移除实体（需要提供旧位置）
func (g *SpatialGrid) Remove(id uint64, x, y int) {
	idx := g.cellIndex(x, y)
	if g.cells[idx] == nil {
		return
	}
	delete(g.cells[idx], id)
	// 不主动清空空 map，避免频繁分配
}

// Move 更新实体位置（先从旧位置移除，再添加到新位置）
func (g *SpatialGrid) Move(id uint64, oldX, oldY, newX, newY int) {
	oldIdx := g.cellIndex(oldX, oldY)
	newIdx := g.cellIndex(newX, newY)
	if oldIdx == newIdx {
		return // 同一格子，无需操作
	}
	if g.cells[oldIdx] != nil {
		delete(g.cells[oldIdx], id)
	}
	if g.cells[newIdx] == nil {
		g.cells[newIdx] = make(map[uint64]bool)
	}
	g.cells[newIdx][id] = true
}

// QueryNearby 查询指定坐标周围格子内的所有实体ID
// radius=0 表示只查当前格子，1 表示查 3×3 范围，2 表示 5×5 范围
// 返回的 map 用于去重，调用方应遍历 keys
func (g *SpatialGrid) QueryNearby(x, y, radius int) map[uint64]bool {
	cx := x / g.cellSize
	cy := y / g.cellSize
	if cx < 0 {
		cx = 0
	} else if cx >= g.cols {
		cx = g.cols - 1
	}
	if cy < 0 {
		cy = 0
	} else if cy >= g.rows {
		cy = g.rows - 1
	}

	result := make(map[uint64]bool)
	for dy := -radius; dy <= radius; dy++ {
		ny := cy + dy
		if ny < 0 || ny >= g.rows {
			continue
		}
		for dx := -radius; dx <= radius; dx++ {
			nx := cx + dx
			if nx < 0 || nx >= g.cols {
				continue
			}
			cell := g.cells[ny*g.cols+nx]
			for id := range cell {
				result[id] = true
			}
		}
	}
	return result
}

// QueryNearbySlice 查询指定坐标周围格子内的实体ID（返回切片）
// 比 QueryNearby 少一次 map 分配，适合只遍历一次的场景
func (g *SpatialGrid) QueryNearbySlice(x, y, radius int) []uint64 {
	ids := g.QueryNearby(x, y, radius)
	result := make([]uint64, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result
}

// Reset 清空网格（重置所有格子）
func (g *SpatialGrid) Reset() {
	for i := range g.cells {
		g.cells[i] = nil
	}
}

// CellSize 返回格子大小
func (g *SpatialGrid) CellSize() int { return g.cellSize }

// Cols 返回网格列数
func (g *SpatialGrid) Cols() int { return g.cols }

// Rows 返回网格行数
func (g *SpatialGrid) Rows() int { return g.rows }
