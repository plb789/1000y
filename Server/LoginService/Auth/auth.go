package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	common "game-server/Common"
	"game-server/DBService/model"
	"game-server/DBService/mysql"
	"game-server/DBService/redis"
	"time"

	"golang.org/x/crypto/bcrypt"
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

// CheckLogin 账号密码校验
func CheckLogin(username, pwd string) (int, uint, string) {
	var acc model.Account
	err := mysql.DB.Where("username = ?", username).First(&acc).Error
	if err != nil {
		return common.CodeFail, 0, "账号不存在"
	}

	// 使用bcrypt验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(pwd)); err != nil {
		return common.CodeFail, 0, "密码错误"
	}

	// status: 0=正常, 1=封禁
	if acc.Status == 1 {
		return common.CodeFail, 0, "账号已被封禁"
	}

	// 更新登录信息
	now := time.Now()
	mysql.DB.Model(&acc).Updates(map[string]interface{}{
		"last_login_time": now,
		"last_login_ip":   "127.0.0.1",
	})

	return common.CodeSuccess, acc.ID, "登录成功"
}

// Register 注册账号
func Register(req RegisterRequest) (int, uint, string) {
	// 检查用户名是否已存在
	var count int64
	mysql.DB.Model(&model.Account{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return common.CodeFail, 0, "用户名已存在"
	}

	// 使用bcrypt加密密码
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return common.CodeFail, 0, "密码加密失败"
	}

	// 创建账号 (status: 0=正常, 1=封禁)
	account := model.Account{
		Username: req.Username,
		Password: string(pwdHash),
		Status:   0, // 新账号默认正常
	}

	if err := mysql.DB.Create(&account).Error; err != nil {
		return common.CodeFail, 0, "注册失败"
	}

	return common.CodeSuccess, account.ID, "注册成功"
}

// GenerateToken 生成Token
func GenerateToken(uid uint) string {
	token := generateSalt()
	// 存储到Redis: uid -> token
	redis.Set(fmt.Sprintf("token:%d", uid), token, 7*24*time.Hour)
	// 同时存储反向索引: token -> uid
	redis.Set(fmt.Sprintf("token_rev:%s", token), fmt.Sprintf("%d", uid), 7*24*time.Hour)
	return token
}

// ValidateToken 验证Token
func ValidateToken(token string) (uint, error) {
	// 从反向索引查找uid
	uidStr, err := redis.Get(fmt.Sprintf("token_rev:%s", token))
	if err != nil {
		return 0, errors.New("无效的token")
	}
	if uidStr == "" {
		return 0, errors.New("无效的token")
	}

	var uid uint
	fmt.Sscanf(uidStr, "%d", &uid)

	// 验证token是否匹配
	storedToken, err := redis.Get(fmt.Sprintf("token:%d", uid))
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
	redis.Set("verify:"+email, code, 5*time.Minute)
	// TODO: 发送邮件
	fmt.Printf("验证码: %s\n", code)
	return nil
}

// CheckCode 验证验证码
func CheckCode(email, code string) bool {
	stored, err := redis.Get("verify:" + email)
	if err != nil {
		return false
	}
	return stored == code
}

// ResetPassword 重置密码
func ResetPassword(username, code, newPwd string) error {
	var acc model.Account
	if err := mysql.DB.Where("username = ?", username).First(&acc).Error; err != nil {
		return errors.New("账号不存在")
	}

	// 验证验证码(简化处理,暂不使用)
	_ = code

	// 使用bcrypt加密新密码
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := mysql.DB.Model(&acc).Update("password", string(pwdHash)).Error; err != nil {
		return errors.New("密码更新失败")
	}

	return nil
}

// UpdatePassword 修改密码
func UpdatePassword(uid uint, oldPwd, newPwd string) error {
	var acc model.Account
	if err := mysql.DB.First(&acc, uid).Error; err != nil {
		return errors.New("账号不存在")
	}

	// 使用bcrypt验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(oldPwd)); err != nil {
		return errors.New("原密码错误")
	}

	// 使用bcrypt加密新密码
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	mysql.DB.Model(&acc).Update("password", string(newHash))

	return nil
}

// UpdateStatus 更新账号状态
func UpdateStatus(uid uint, status int) error {
	return mysql.DB.Model(&model.Account{}).Where("id = ?", uid).Update("status", status).Error
}

// GetAccountInfo 获取账号信息
func GetAccountInfo(uid uint) (*model.Account, error) {
	var acc model.Account
	if err := mysql.DB.First(&acc, uid).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

// IsBanned 检查账号是否被封禁
func IsBanned(uid uint) bool {
	acc, err := GetAccountInfo(uid)
	if err != nil {
		return false
	}
	// status: 0=正常, 1=封禁
	return acc.Status == 1
}

// GetUidByToken 通过Token获取UID
func GetUidByToken(token string) (uint, error) {
	// 从反向索引查找uid
	uidStr, err := redis.Get(fmt.Sprintf("token_rev:%s", token))
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
	token, err := redis.Get(fmt.Sprintf("token:%d", uid))
	if err == nil && token != "" {
		redis.Del(fmt.Sprintf("token_rev:%s", token))
	}
	redis.Del(fmt.Sprintf("token:%d", uid))
	return nil
}
