package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JWTHeader JWT 头部
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTPayload JWT 负载
type JWTPayload struct {
	UserID string `json:"userId"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// GenerateToken 生成 JWT token
func GenerateToken(userID string, jwtSecret string) (string, error) {
	// 创建头部
	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	// 创建负载
	now := time.Now().Unix()
	payload := JWTPayload{
		UserID: userID,
		Exp:    now + (7 * 24 * 60 * 60), // 7天过期
		Iat:    now,
	}

	// Base64URL 编码
	headerEncoded, err := base64URLEncode(header)
	if err != nil {
		return "", err
	}

	payloadEncoded, err := base64URLEncode(payload)
	if err != nil {
		return "", err
	}

	// 生成签名
	signature, err := generateSignature(headerEncoded+"."+payloadEncoded, jwtSecret)
	if err != nil {
		return "", err
	}

	// 组合最终的 token
	token := fmt.Sprintf("%s.%s.%s", headerEncoded, payloadEncoded, signature)
	return token, nil
}

// ValidateToken 验证 JWT token
func ValidateToken(token, jwtSecret string) (*JWTPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signature := parts[2]

	// 验证签名
	expectedSignature, err := generateSignature(headerEncoded+"."+payloadEncoded, jwtSecret)
	if err != nil {
		return nil, err
	}

	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	// 解码负载
	payload, err := base64URLDecode(payloadEncoded)
	if err != nil {
		return nil, err
	}

	var jwtPayload JWTPayload
	if err := json.Unmarshal(payload, &jwtPayload); err != nil {
		return nil, err
	}

	// 检查过期时间
	if time.Now().Unix() > jwtPayload.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &jwtPayload, nil
}

// base64URLEncode Base64URL 编码
func base64URLEncode(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.EncodeToString(jsonData)
	// 移除填充字符
	encoded = strings.TrimRight(encoded, "=")
	return encoded, nil
}

// base64URLDecode Base64URL 解码
func base64URLDecode(data string) ([]byte, error) {
	// 添加必要的填充字符
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	return base64.URLEncoding.DecodeString(data)
}

// generateSignature 生成 HMAC-SHA256 签名
func generateSignature(data, secret string) (string, error) {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	// 移除填充字符
	signature = strings.TrimRight(signature, "=")
	return signature, nil
}
