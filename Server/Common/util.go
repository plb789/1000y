package common

import (
	"encoding/binary"
)

// EncodePacket 组装二进制封包: [cmd(2)][len(2)][body][check(1)]
func EncodePacket(cmd uint16, body []byte) []byte {
	bodyLen := len(body)
	totalLen := 4 + bodyLen + 1
	pkg := make([]byte, totalLen)

	binary.LittleEndian.PutUint16(pkg[0:2], cmd)
	binary.LittleEndian.PutUint16(pkg[2:4], uint16(bodyLen))
	copy(pkg[4:], body)

	// 计算校验码
	var check byte = 0
	for i := 0; i < totalLen-1; i++ {
		check += pkg[i]
	}
	pkg[totalLen-1] = check
	return pkg
}

// CheckPacket 校验封包
func CheckPacket(data []byte) bool {
	if len(data) < 5 {
		return false
	}
	var sum int = 0
	for i := 0; i < len(data)-1; i++ {
		sum += int(data[i])
	}
	return byte(sum&0xFF) == data[len(data)-1]
}

// DecodePacket 解析封包，返回 cmd, body
func DecodePacket(data []byte) (uint16, []byte) {
	cmd := binary.LittleEndian.Uint16(data[0:2])
	bodyLen := binary.LittleEndian.Uint16(data[2:4])
	body := data[4 : 4+bodyLen]
	return cmd, body
}