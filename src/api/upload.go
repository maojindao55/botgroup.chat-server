package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"project/src/config"
	"project/src/models"

	"github.com/gin-gonic/gin"
)

// UploadResponse 上传响应结构
type UploadResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CloudflareUploadResponse Cloudflare API响应结构
type CloudflareUploadResponse struct {
	Success bool        `json:"success"`
	Errors  []string    `json:"errors"`
	Result  interface{} `json:"result"`
}

// UploadHandler 处理文件上传请求
func UploadHandler(c *gin.Context) {
	// 从认证中间件中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, UploadResponse{
			Success: false,
			Message: "用户认证失败",
		})
		return
	}

	// 类型断言
	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "用户信息类型错误",
		})
		return
	}

	// 检查 Cloudflare 配置
	if config.AppConfig.Cloudflare.AccountID == "" || config.AppConfig.Cloudflare.APIToken == "" {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Cloudflare配置缺失",
		})
		return
	}

	// 创建 multipart/form-data 请求
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加 requireSignedURLs 字段
	if err := writer.WriteField("requireSignedURLs", "false"); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "创建表单字段失败: " + err.Error(),
		})
		return
	}

	// 添加 metadata 字段
	metadata := map[string]interface{}{
		"user": user.ID,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "序列化metadata失败: " + err.Error(),
		})
		return
	}
	if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "创建metadata字段失败: " + err.Error(),
		})
		return
	}

	// 关闭 multipart writer
	writer.Close()

	// 构建 Cloudflare API URL
	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/images/v2/direct_upload", config.AppConfig.Cloudflare.AccountID)

	// 创建请求
	req, err := http.NewRequest("POST", apiURL, &buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "创建请求失败: " + err.Error(),
		})
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.Cloudflare.APIToken)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "请求Cloudflare API失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "读取响应失败: " + err.Error(),
		})
		return
	}

	// 解析响应
	var cloudflareResp CloudflareUploadResponse
	if err := json.Unmarshal(body, &cloudflareResp); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "解析响应失败: " + err.Error() + " " + string(body),
		})
		return
	}

	// 检查 Cloudflare API 响应状态
	if !cloudflareResp.Success {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "Cloudflare API错误: " + fmt.Sprintf("%v", cloudflareResp.Errors),
		})
		return
	}

	// 记录日志
	fmt.Printf("上传头像成功: %v\n", cloudflareResp.Result)

	// 返回成功响应
	c.JSON(http.StatusOK, UploadResponse{
		Success: true,
		Message: "获取上传URL成功",
		Data:    cloudflareResp.Result,
	})
}
