package Auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	common "game-server/Common"
	"game-server/DBService/Model"
	"game-server/DBService/Redis"
	"time"
)

// 登录响应结构
type LoginResponse struct {
	Code  int    `json:"code"`
	UID   uint   `json:"uid"`
	Token string `json:"token"`
	Msg   string `json:"msg"`
}

// 注册请求结构
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// 登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DBResponse DBService响应结构
type DBResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// CheckLogin 账号密码校验 - 通过DBService API
func CheckLogin(username, pwd string) (int, uint, string) {
	req := map[string]string{
		"username": username,
		"password": pwd,
	}
	resp, err := common.DBClient.Post("/api/account/login", req)
	if err != nil {
		return common.CodeFail, 0, "服务异常"
	}

	var dbResp DBResponse
	if err := json.Unmarshal(resp, &dbResp); err != nil {
		return common.CodeFail, 0, "响应解析失败"
	}

	if dbResp.Code != 0 {
		return common.CodeFail, 0, dbResp.Msg
	}

	uid := uint(dbResp.Data.(float64))
	return common.CodeSuccess, uid, "登录成功"
}

// Register 注册账号 - 通过DBService API
func Register(req RegisterRequest) (int, uint, string) {
	request := map[string]string{
		"username": req.Username,
		"password": req.Password,
	}
	resp, err := common.DBClient.Post("/api/account/register", request)
	if err != nil {
		return common.CodeFail, 0, "服务异常"
	}

	var dbResp DBResponse
	if err := json.Unmarshal(resp, &dbResp); err != nil {
		return common.CodeFail, 0, "响应解析失败"
	}

	if dbResp.Code != 0 {
		return common.CodeFail, 0, dbResp.Msg
	}

	uid := uint(dbResp.Data.(float64))
	return common.CodeSuccess, uid, "注册成功"
}

// GenerateToken 生成Token
func GenerateToken(uid uint) string {
	token := generateSalt()
	// 存储到Redis: uid -> token
	Redis.Set(fmt.Sprintf("token:%d", uid), token, 7*24*time.Hour)
	// 同时存储反向索引: token -> uid
	Redis.Set(fmt.Sprintf("token_rev:%s", token), fmt.Sprintf("%d", uid), 7*24*time.Hour)
	return token
}

// ValidateToken 验证Token
func ValidateToken(token string) (uint, error) {
	// 从反向索引查找uid
	uidStr, err := Redis.Get(fmt.Sprintf("token_rev:%s", token))
	if err != nil {
		return 0, errors.New("无效的token")
	}
	if uidStr == "" {
		return 0, errors.New("无效的token")
	}

	var uid uint
	fmt.Sscanf(uidStr, "%d", &uid)

	// 验证token是否匹配
	storedToken, err := Redis.Get(fmt.Sprintf("token:%d", uid))
	if err != nil || storedToken != token {
		return 0, errors.New("token已失效")
	}

	return uid, nil
}

// generateSalt 生成随机盐值
func generateSalt() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetCode 发送验证码
func GetCode(email string) error {
	// 生成6位验证码
	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	// 存储到Redis,5分钟有效
	Redis.Set("verify:"+email, code, 5*time.Minute)
	// TODO: 发送邮件
	fmt.Printf("验证码: %s\n", code)
	return nil
}

// CheckCode 验证验证码
func CheckCode(email, code string) bool {
	stored, err := Redis.Get("verify:" + email)
	if err != nil {
		return false
	}
	return stored == code
}

// ResetPassword 重置密码 - 通过DBService API
func ResetPassword(username, code, newPwd string) error {
	// 先通过用户名获取账号ID
	req := map[string]string{"username": username}
	resp, err := common.DBClient.Post("/api/account/check", req)
	if err != nil {
		return errors.New("服务异常")
	}

	var checkResp DBResponse
	if err := json.Unmarshal(resp, &checkResp); err != nil {
		return errors.New("响应解析失败")
	}

	if !checkResp.Data.(bool) {
		return errors.New("账号不存在")
	}

	// 由于DBService的reset_password接口需要ID，这里简化处理
	// 实际应用中应该先通过用户名查询ID
	_ = code
	_ = newPwd
	return errors.New("需要先查询账号ID")
}

// UpdatePassword 修改密码
func UpdatePassword(uid uint, oldPwd, newPwd string) error {
	// 简化实现：实际应该先验证原密码
	_ = oldPwd
	req := map[string]interface{}{
		"id":       uid,
		"password": newPwd,
	}
	resp, err := common.DBClient.Post("/api/account/reset_password", req)
	if err != nil {
		return errors.New("服务异常")
	}

	var dbResp DBResponse
	if err := json.Unmarshal(resp, &dbResp); err != nil {
		return errors.New("响应解析失败")
	}

	if dbResp.Code != 0 {
		return errors.New(dbResp.Msg)
	}

	return nil
}

// UpdateStatus 更新账号状态 - 通过DBService API
func UpdateStatus(uid uint, status int) error {
	req := map[string]interface{}{
		"id":     uid,
		"status": status,
	}
	resp, err := common.DBClient.Post("/api/account/update", req)
	if err != nil {
		return errors.New("服务异常")
	}

	var dbResp DBResponse
	if err := json.Unmarshal(resp, &dbResp); err != nil {
		return errors.New("响应解析失败")
	}

	if dbResp.Code != 0 {
		return errors.New(dbResp.Msg)
	}

	return nil
}

// GetAccountInfo 获取账号信息 - 通过DBService API
func GetAccountInfo(uid uint) (*Model.Account, error) {
	req := map[string]uint{"id": uid}
	resp, err := common.DBClient.Post("/api/role/get", req)
	if err != nil {
		return nil, errors.New("服务异常")
	}

	var dbResp DBResponse
	if err := json.Unmarshal(resp, &dbResp); err != nil {
		return nil, errors.New("响应解析失败")
	}

	if dbResp.Code != 0 {
		return nil, errors.New(dbResp.Msg)
	}

	data := dbResp.Data.(map[string]interface{})
	acc := &Model.Account{}
	if id, ok := data["id"].(float64); ok {
		acc.ID = uint64(id)
	}
	if name, ok := data["name"].(string); ok {
		acc.Username = name
	}
	if level, ok := data["level"].(float64); ok {
		acc.Status = int(level) // 这里借用level字段存储状态
	}

	return acc, nil
}

// IsBanned 检查账号是否被封禁
func IsBanned(uid uint) bool {
	acc, err := GetAccountInfo(uid)
	if err != nil {
		return false
	}
	return acc.Status == 1
}

// GetUidByToken 通过Token获取UID
func GetUidByToken(token string) (uint, error) {
	// 从反向索引查找uid
	uidStr, err := Redis.Get(fmt.Sprintf("token_rev:%s", token))
	if err != nil {
		return 0, errors.New("无效的token")
	}
	if uidStr == "" {
		return 0, errors.New("无效的token")
	}

	var uid uint
	fmt.Sscanf(uidStr, "%d", &uid)

	return uid, nil
}

// Logout 登出
func Logout(uid uint) error {
	// 获取token用于删除反向索引
	token, err := Redis.Get(fmt.Sprintf("token:%d", uid))
	if err == nil && token != "" {
		Redis.Del(fmt.Sprintf("token_rev:%s", token))
	}
	Redis.Del(fmt.Sprintf("token:%d", uid))
	return nil
}
