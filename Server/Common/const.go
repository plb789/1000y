package common

// WebSocket 协议号
const (
	CmdHeart    uint16 = 0x0001 // 心跳包
	CmdLogin    uint16 = 0x0002 // 登录请求
	CmdMove     uint16 = 0x0003 // 角色移动
	CmdChat     uint16 = 0x0004 // 聊天
)

// 通用错误码
const (
	CodeSuccess = 0
	CodeFail    = -1
)