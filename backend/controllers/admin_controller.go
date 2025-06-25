package controllers

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type AdminController struct {
	Config *config.Config
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users []UserInfo `json:"users"`
	Total int64      `json:"total"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	UserID uint   `json:"userId" binding:"required"`
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// UpdateUserRoleRequest 更新用户角色请求
type UpdateUserRoleRequest struct {
	UserID uint   `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin user"`
}

// checkAdminRole 检查是否为管理员
func (ctrl *AdminController) checkAdminRole(c *gin.Context) bool {
	role, exists := c.Get("role")
	if !exists {
		return false
	}

	roleStr, ok := role.(string)
	if !ok {
		return false
	}

	return roleStr == "admin"
}

// GetUsers 获取用户列表
func (ctrl *AdminController) GetUsers(c *gin.Context) {
	// 检查是否为管理员
	if !ctrl.checkAdminRole(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 获取查询参数
	status := c.Query("status") // pending, active, disabled
	role := c.Query("role")     // admin, user

	// 构建查询
	query := ctrl.Config.DB.Model(&models.User{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("获取用户总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 获取用户列表
	var users []models.User
	if err := query.Order("created_at desc").Find(&users).Error; err != nil {
		log.Printf("获取用户列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 转换为响应格式
	userList := make([]UserInfo, 0, len(users))
	for _, user := range users {
		userList = append(userList, UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, UserListResponse{
		Users: userList,
		Total: total,
	})
}

// ApproveUser 审核通过用户
func (ctrl *AdminController) ApproveUser(c *gin.Context) {
	// 检查是否为管理员
	if !ctrl.checkAdminRole(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req struct {
		UserID uint `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 查找用户
	var user models.User
	if err := ctrl.Config.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 检查用户状态
	if user.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户状态不是待审核"})
		return
	}

	// 更新用户状态为active
	user.Status = "active"
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("审核用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}

	log.Printf("管理员审核通过用户: %s (ID: %d)", user.Username, user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "用户审核通过"})
}

// UpdateUserStatus 更新用户状态
func (ctrl *AdminController) UpdateUserStatus(c *gin.Context) {
	// 检查是否为管理员
	if !ctrl.checkAdminRole(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取当前用户ID，防止自己禁用自己
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(uint) == req.UserID && req.Status == "disabled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能禁用自己的账号"})
		return
	}

	// 查找用户
	var user models.User
	if err := ctrl.Config.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 更新用户状态
	user.Status = req.Status
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("更新用户状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	log.Printf("管理员更新用户状态: %s (ID: %d) -> %s", user.Username, user.ID, req.Status)
	c.JSON(http.StatusOK, gin.H{"message": "用户状态更新成功"})
}

// UpdateUserRole 更新用户角色
func (ctrl *AdminController) UpdateUserRole(c *gin.Context) {
	// 检查是否为管理员
	if !ctrl.checkAdminRole(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取当前用户ID，防止自己降级自己
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(uint) == req.UserID && req.Role == "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能降级自己的权限"})
		return
	}

	// 查找用户
	var user models.User
	if err := ctrl.Config.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 更新用户角色
	user.Role = req.Role
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("更新用户角色失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新角色失败"})
		return
	}

	log.Printf("管理员更新用户角色: %s (ID: %d) -> %s", user.Username, user.ID, req.Role)
	c.JSON(http.StatusOK, gin.H{"message": "用户角色更新成功"})
}

// GetUserStats 获取用户统计信息
func (ctrl *AdminController) GetUserStats(c *gin.Context) {
	// 检查是否为管理员
	if !ctrl.checkAdminRole(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var stats struct {
		TotalUsers    int64 `json:"totalUsers"`
		PendingUsers  int64 `json:"pendingUsers"`
		ActiveUsers   int64 `json:"activeUsers"`
		DisabledUsers int64 `json:"disabledUsers"`
		AdminUsers    int64 `json:"adminUsers"`
	}

	// 获取各种状态的用户数量
	ctrl.Config.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	ctrl.Config.DB.Model(&models.User{}).Where("status = ?", "pending").Count(&stats.PendingUsers)
	ctrl.Config.DB.Model(&models.User{}).Where("status = ?", "active").Count(&stats.ActiveUsers)
	ctrl.Config.DB.Model(&models.User{}).Where("status = ?", "disabled").Count(&stats.DisabledUsers)
	ctrl.Config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&stats.AdminUsers)

	c.JSON(http.StatusOK, stats)
}
