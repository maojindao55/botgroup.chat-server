package api

import (
	"net/http"
	"project/src/config"
	"project/src/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateGroupHandler 创建群组
func CreateGroupHandler(c *gin.Context) {
	var req models.LlmGroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.LlmGroupResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 创建群组
	group := models.LlmGroup{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := config.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "创建群组失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.LlmGroupResponse{
		Success: true,
		Message: "创建群组成功",
		Data:    &group,
	})
}

// GetGroupsHandler 获取群组列表
func GetGroupsHandler(c *gin.Context) {
	var groups []models.LlmGroup
	var total int64

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 搜索参数
	name := c.Query("name")

	// 构建查询
	query := config.DB.Model(&models.LlmGroup{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.LlmGroupListResponse{
			Success: false,
			Message: "获取群组总数失败: " + err.Error(),
		})
		return
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.LlmGroupListResponse{
			Success: false,
			Message: "获取群组列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.LlmGroupListResponse{
		Success: true,
		Message: "获取群组列表成功",
		Data:    groups,
		Total:   total,
	})
}

// GetGroupHandler 获取单个群组详情
func GetGroupHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.LlmGroupResponse{
			Success: false,
			Message: "无效的群组ID",
		})
		return
	}

	var group models.LlmGroup
	if err := config.DB.Preload("Characters").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.LlmGroupResponse{
				Success: false,
				Message: "群组不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "获取群组详情失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.LlmGroupResponse{
		Success: true,
		Message: "获取群组详情成功",
		Data:    &group,
	})
}

// UpdateGroupHandler 更新群组
func UpdateGroupHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.LlmGroupResponse{
			Success: false,
			Message: "无效的群组ID",
		})
		return
	}

	var req models.LlmGroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.LlmGroupResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查群组是否存在
	var group models.LlmGroup
	if err := config.DB.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.LlmGroupResponse{
				Success: false,
				Message: "群组不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "查询群组失败: " + err.Error(),
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.LlmGroupResponse{
			Success: false,
			Message: "至少需要提供一个更新字段",
		})
		return
	}

	// 执行更新
	if err := config.DB.Model(&group).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "更新群组失败: " + err.Error(),
		})
		return
	}

	// 重新获取更新后的数据
	if err := config.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "获取更新后的群组信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.LlmGroupResponse{
		Success: true,
		Message: "更新群组成功",
		Data:    &group,
	})
}

// DeleteGroupHandler 删除群组
func DeleteGroupHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.LlmGroupResponse{
			Success: false,
			Message: "无效的群组ID",
		})
		return
	}

	// 检查群组是否存在
	var group models.LlmGroup
	if err := config.DB.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.LlmGroupResponse{
				Success: false,
				Message: "群组不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "查询群组失败: " + err.Error(),
		})
		return
	}

	// 删除群组（会级联删除相关角色）
	if err := config.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.LlmGroupResponse{
			Success: false,
			Message: "删除群组失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.LlmGroupResponse{
		Success: true,
		Message: "删除群组成功",
	})
}

