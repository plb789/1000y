package common

import (
	"errors"
	"fmt"
	"game-server/DBService/redis"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UID      uint64 `json:"uid"`
	Username string `json:"username"`
	RoleID   uint64 `json:"role_id"` // 当前选中的角色ID
	jwt.RegisteredClaims
}

// SecretKey JWT密钥(从环境变量读取，默认值仅用于开发环境)
var SecretKey = getSecretKey()

func getSecretKey() []byte {
	if key := os.Getenv("JWT_SECRET_KEY"); key != "" {
		return []byte(key)
	}
	return []byte("千年江湖_JWT_Secret_Key_2024")
}

// GenerateToken 生成JWT Token
func GenerateToken(uid uint64, username string) (string, error) {
	claims := JWTClaims{
		UID:      uid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7天有效期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "千年江湖",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

// GenerateTokenWithRole 生成带角色ID的Token
func GenerateTokenWithRole(uid uint64, username string, roleID uint64) (string, error) {
	claims := JWTClaims{
		UID:      uid,
		Username: username,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "千年江湖",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

// ValidateToken 验证Token
func ValidateToken(tokenStr string) (*JWTClaims, error) {
	// 去掉Bearer前缀
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token解析失败: %v", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// ValidateTokenSimple 简单Token验证(仅验证格式和签名)
func ValidateTokenSimple(tokenStr string) (*JWTClaims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// SetTokenToRedis 将Token存储到Redis
func SetTokenToRedis(uid uint64, token string, expire time.Duration) error {
	key := fmt.Sprintf("token:%d", uid)
	return redis.Set(key, token, expire)
}

// GetTokenFromRedis 从Redis获取Token
func GetTokenFromRedis(uid uint64) (string, error) {
	key := fmt.Sprintf("token:%d", uid)
	return redis.Get(key)
}

// DeleteTokenFromRedis 删除Redis中的Token
func DeleteTokenFromRedis(uid uint64) error {
	key := fmt.Sprintf("token:%d", uid)
	return redis.Del(key)
}

// IsTokenValid 检查Token是否有效(Redis中是否存在)
func IsTokenValid(uid uint64, token string) bool {
	storedToken, err := GetTokenFromRedis(uid)
	if err != nil {
		return false
	}
	return storedToken == token
}

// ServiceToken 服务间通信的Token(用于微服务之间的认证)
var ServiceToken = getServiceToken()

func getServiceToken() string {
	if token := os.Getenv("SERVICE_TOKEN"); token != "" {
		return token
	}
	return "game-service-internal-token-2024"
}

// ValidateServiceToken 验证服务间通信Token
func ValidateServiceToken(token string) bool {
	return token == ServiceToken
}
