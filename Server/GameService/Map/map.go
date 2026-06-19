package gamemap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// Tile 瓦片结构（支持16位瓦片ID）
type Tile struct {
	Low  uint16 // 底层瓦片索引（16位，支持0-65535）
	High uint16 // 高层瓦片索引（16位，支持0-65535）
	Attr uint8  // 属性: 1=阻挡
}

type GameMap struct {
	Width     uint16
	Height    uint16
	Collision [][]bool
}

var GlobalMap *GameMap

// LoadMapFile 加载千年 .map 文件
// 支持两种格式：
//   旧格式：每瓦片3字节(Low/u8 + High/u8 + Attr/u8)
//   新格式：每瓦片5字节(Low/u16 + High/u16 + Attr/u8)
func LoadMapFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 跳过128字节文件头
	if len(data) < 132 {
		return fmt.Errorf("map file too small: %d bytes", len(data))
	}
	reader := bytes.NewReader(data[128:])

	var w, h uint16
	_ = binary.Read(reader, binary.LittleEndian, &w)
	_ = binary.Read(reader, binary.LittleEndian, &h)

	total := int(w) * int(h)

	// 自动检测文件格式：通过计算期望文件大小来判断
	// 新格式: 128字节头 + 4字节宽高 + 每瓦片5字节
	// 旧格式: 128字节头 + 4字节宽高 + 每瓦片3字节
	expectedSizeNewFormat := 128 + 4 + total*5
	expectedSizeOldFormat := 128 + 4 + total*3
	isNewFormat := abs(len(data)-expectedSizeNewFormat) <= abs(len(data)-expectedSizeOldFormat)

	fmt.Printf("📦 服务端加载地图: %s\n", path)
	fmt.Printf("   尺寸: %d x %d = %d 瓦片\n", w, h, total)
	fmt.Printf("   格式: %s\n", mapBool(isNewFormat, "新格式(16位ID)", "旧格式(8位ID)"))

	tiles := make([]Tile, total)

	for i := 0; i < total; i++ {
		var t Tile
		if isNewFormat {
			// 新格式：16位瓦片ID
			_ = binary.Read(reader, binary.LittleEndian, &t.Low)
			_ = binary.Read(reader, binary.LittleEndian, &t.High)
		} else {
			// 旧格式：8位瓦片ID（读取后转为16位）
			var low, high uint8
			_ = binary.Read(reader, binary.LittleEndian, &low)
			_ = binary.Read(reader, binary.LittleEndian, &high)
			t.Low = uint16(low)
			t.High = uint16(high)
		}
		_ = binary.Read(reader, binary.LittleEndian, &t.Attr)
		tiles[i] = t
	}

	// 构建碰撞表
	coll := make([][]bool, h)
	for y := 0; y < int(h); y++ {
		coll[y] = make([]bool, w)
		for x := 0; x < int(w); x++ {
			idx := y*int(w) + x
			coll[y][x] = (tiles[idx].Attr == 1)
		}
	}

	GlobalMap = &GameMap{
		Width:     w,
		Height:    h,
		Collision: coll,
	}
	return nil
}

// abs 返回绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// mapBool 根据条件返回对应字符串
func mapBool(cond bool, trueStr, falseStr string) string {
	if cond {
		return trueStr
	}
	return falseStr
}

// CanMove 校验是否可移动（防穿墙）
func CanMove(x, y int) bool {
	m := GlobalMap
	if x < 0 || y < 0 || x >= int(m.Width) || y >= int(m.Height) {
		return false
	}
	return !m.Collision[y][x]
}