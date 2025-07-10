package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"project/src/config"
	"sort"
	"strings"
	"time"
)

// SMSService 短信服务接口
type SMSService interface {
	SendSMS(phone, code string) error
	SendSMSWithTemplate(phone, templateCode string, templateParam map[string]string) error
}

// smsService 短信服务实现
type smsService struct {
	config config.AliyunSMSConfig
}

// NewSMSService 创建短信服务实例
func NewSMSService(config config.AliyunSMSConfig) SMSService {
	return &smsService{
		config: config,
	}
}

// SendSMS 发送验证码短信
func (s *smsService) SendSMS(phone, code string) error {
	templateParam := map[string]string{
		"code": code,
	}
	return s.SendSMSWithTemplate(phone, s.config.TemplateCode, templateParam)
}

// SendSMSWithTemplate 发送自定义模板短信
func (s *smsService) SendSMSWithTemplate(phone, templateCode string, templateParam map[string]string) error {
	const (
		apiURL     = "https://dysmsapi.aliyuncs.com/"
		apiVersion = "2017-05-25"
	)
	fmt.Println(s.config)

	// 准备请求参数
	params := map[string]string{
		"AccessKeyId":      s.config.AccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"SignName":         s.config.SignName,
		"TemplateCode":     templateCode,
		"Version":          apiVersion,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   generateNonce(16),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}

	// 添加模板参数
	if len(templateParam) > 0 {
		templateParamJSON, err := json.Marshal(templateParam)
		if err != nil {
			return fmt.Errorf("marshal template param failed: %v", err)
		}
		params["TemplateParam"] = string(templateParamJSON)
	}

	// 参数排序
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var canonicalizedQueryString []string
	for _, key := range keys {
		canonicalizedQueryString = append(canonicalizedQueryString,
			fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(params[key])))
	}

	queryString := strings.Join(canonicalizedQueryString, "&")
	stringToSign := fmt.Sprintf("GET&%s&%s",
		url.QueryEscape("/"),
		url.QueryEscape(queryString))

	// 计算签名
	signature := calculateHmacSha1(stringToSign, s.config.AccessKeySecret+"&")

	// 构建最终的 URL
	finalURL := fmt.Sprintf("%s?%s&Signature=%s",
		apiURL, queryString, url.QueryEscape(signature))

	// 发送请求
	resp, err := http.Get(finalURL)
	if err != nil {
		return fmt.Errorf("send http request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %v", err)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("parse response failed: %v, body: %s", err, string(body))
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http request failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	// 检查业务状态
	if code, ok := response["Code"].(string); !ok || code != "OK" {
		message := "Unknown error"
		if msg, ok := response["Message"].(string); ok {
			message = msg
		}
		return fmt.Errorf("SMS send failed: %s (Code: %s)", message, code)
	}

	return nil
}

// generateNonce 生成随机字符串
func generateNonce(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)

	nonce := make([]string, length)
	for i, b := range bytes {
		nonce[i] = fmt.Sprintf("%02x", b)
	}

	return strings.Join(nonce, "")
}

// calculateHmacSha1 计算 HMAC-SHA1 签名
func calculateHmacSha1(message, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
