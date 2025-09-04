package api

import (
	"net/http"
	"project/src/config"
	"project/src/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCharacterHandler 创建群组角色
func CreateCharacterHandler(c *gin.Context) {
	var req models.GroupCharacterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证群组是否存在
	var group models.LlmGroup
	if err := config.DB.First(&group, req.GID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
				Success: false,
				Message: "指定的群组不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "验证群组失败: " + err.Error(),
		})
		return
	}

	// 创建角色
	character := models.GroupCharacter{
		GID:          req.GID,
		Name:         req.Name,
		Personality:  req.Personality,
		Model:        req.Model,
		Avatar:       req.Avatar,
		CustomPrompt: req.CustomPrompt,
	}

	if err := config.DB.Create(&character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "创建角色失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GroupCharacterResponse{
		Success: true,
		Message: "创建角色成功",
		Data:    &character,
	})
}

// GetCharactersHandler 获取角色列表
func GetCharactersHandler(c *gin.Context) {
	var characters []models.GroupCharacter
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
	gid := c.Query("gid")
	name := c.Query("name")
	model := c.Query("model")

	// 构建查询
	query := config.DB.Model(&models.GroupCharacter{})
	if gid != "" {
		query = query.Where("gid = ?", gid)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if model != "" {
		query = query.Where("model = ?", model)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharacterListResponse{
			Success: false,
			Message: "获取角色总数失败: " + err.Error(),
		})
		return
	}

	// 获取分页数据，预加载群组信息
	if err := query.Preload("Group").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&characters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharacterListResponse{
			Success: false,
			Message: "获取角色列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GroupCharacterListResponse{
		Success: true,
		Message: "获取角色列表成功",
		Data:    characters,
		Total:   total,
	})
}

// GetCharactersByGroupHandler 根据群组获取角色列表
func GetCharactersByGroupHandler(c *gin.Context) {
	gid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.GroupCharactersByGroupResponse{
			Success: false,
			Message: "无效的群组ID",
		})
		return
	}

	// 验证群组是否存在
	var group models.LlmGroup
	if err := config.DB.First(&group, gid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.GroupCharactersByGroupResponse{
				Success: false,
				Message: "群组不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.GroupCharactersByGroupResponse{
			Success: false,
			Message: "查询群组失败: " + err.Error(),
		})
		return
	}

	var characters []models.GroupCharacter
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

	// 获取该群组的角色总数
	if err := config.DB.Model(&models.GroupCharacter{}).Where("gid = ?", gid).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharactersByGroupResponse{
			Success: false,
			Message: "获取角色总数失败: " + err.Error(),
		})
		return
	}

	// 获取分页数据
	if err := config.DB.Where("gid = ?", gid).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&characters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharactersByGroupResponse{
			Success: false,
			Message: "获取角色列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GroupCharactersByGroupResponse{
		Success: true,
		Message: "获取群组角色列表成功",
		Group:   &group,
		Data:    characters,
		Total:   total,
	})
}

// GetCharacterHandler 获取单个角色详情
func GetCharacterHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
			Success: false,
			Message: "无效的角色ID",
		})
		return
	}

	var character models.GroupCharacter
	if err := config.DB.Preload("Group").First(&character, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.GroupCharacterResponse{
				Success: false,
				Message: "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "获取角色详情失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GroupCharacterResponse{
		Success: true,
		Message: "获取角色详情成功",
		Data:    &character,
	})
}

// UpdateCharacterHandler 更新角色
func UpdateCharacterHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
			Success: false,
			Message: "无效的角色ID",
		})
		return
	}

	var req models.GroupCharacterUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查角色是否存在
	var character models.GroupCharacter
	if err := config.DB.First(&character, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.GroupCharacterResponse{
				Success: false,
				Message: "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "查询角色失败: " + err.Error(),
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Personality != "" {
		updates["personality"] = req.Personality
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.CustomPrompt != "" {
		updates["custom_prompt"] = req.CustomPrompt
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
			Success: false,
			Message: "至少需要提供一个更新字段",
		})
		return
	}

	// 执行更新
	if err := config.DB.Model(&character).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "更新角色失败: " + err.Error(),
		})
		return
	}

	// 重新获取更新后的数据
	if err := config.DB.Preload("Group").First(&character, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "获取更新后的角色信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GroupCharacterResponse{
		Success: true,
		Message: "更新角色成功",
		Data:    &character,
	})
}

// DeleteCharacterHandler 删除角色
func DeleteCharacterHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.GroupCharacterResponse{
			Success: false,
			Message: "无效的角色ID",
		})
		return
	}

	// 检查角色是否存在
	var character models.GroupCharacter
	if err := config.DB.First(&character, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.GroupCharacterResponse{
				Success: false,
				Message: "角色不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "查询角色失败: " + err.Error(),
		})
		return
	}

	// 删除角色
	if err := config.DB.Delete(&character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.GroupCharacterResponse{
			Success: false,
			Message: "删除角色失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GroupCharacterResponse{
		Success: true,
		Message: "删除角色成功",
	})
}
