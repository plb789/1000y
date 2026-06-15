package gamemap

import (
	"bytes"
	"encoding/binary"
	"os"
)

type Tile struct {
	Low  uint8
	High uint8
	Attr uint8 // 1=阻挡
}

type GameMap struct {
	Width     uint16
	Height    uint16
	Collision [][]bool
}

var GlobalMap *GameMap

// LoadMapFile 加载千年 .map 文件
func LoadMapFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data[128:])

	var w, h uint16
	_ = binary.Read(reader, binary.LittleEndian, &w)
	_ = binary.Read(reader, binary.LittleEndian, &h)

	total := int(w) * int(h)
	tiles := make([]Tile, total)
	for i := 0; i < total; i++ {
		var t Tile
		_ = binary.Read(reader, binary.LittleEndian, &t.Low)
		_ = binary.Read(reader, binary.LittleEndian, &t.High)
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

// CanMove 校验是否可移动（防穿墙）
func CanMove(x, y int) bool {
	m := GlobalMap
	if x < 0 || y < 0 || x >= int(m.Width) || y >= int(m.Height) {
		return false
	}
	return !m.Collision[y][x]
}