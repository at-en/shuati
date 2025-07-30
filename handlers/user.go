package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"quiz-system/models"
	"quiz-system/services"
	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req models.UserRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := services.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}

	// 创建新用户
	user := models.User{
		Username: req.Username,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encrypt password",
		})
		return
	}

	if err := services.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// 查找用户
	var user models.User
	if err := services.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// 检查用户是否被锁定
	if user.IsLocked() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Account is temporarily locked due to too many failed login attempts",
		})
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		// 增加失败次数
		user.IncrementLoginAttempts()
		services.DB.Save(&user)
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// 登录成功，重置失败次数
	user.ResetLoginAttempts()
	services.DB.Save(&user)

	// 生成会话ID
	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate session",
		})
		return
	}

	// 创建用户会话
	session := &services.UserSession{
		UserID:    user.ID,
		Username:  user.Username,
		LoginTime: time.Now(),
		LastSeen:  time.Now(),
	}
	services.Cache.SetUserSession(sessionID, session)

	// 设置Cookie
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true) // 24小时有效

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
		"session_id": sessionID,
	})
}

// Logout 用户登出
func Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Already logged out",
		})
		return
	}

	// 删除会话
	services.Cache.DeleteUserSession(sessionID)
	
	// 清除Cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetProfile 获取用户信息
func GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	userSession := user.(*services.UserSession)
	
	// 获取用户统计信息
	stats, err := services.GetUserStats(userSession.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": models.UserResponse{
			ID:       userSession.UserID,
			Username: userSession.Username,
		},
		"stats": stats,
	})
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No session found",
			})
			c.Abort()
			return
		}

		session, exists := services.Cache.GetUserSession(sessionID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid session",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", session)
		c.Next()
	}
}

// generateSessionID 生成会话ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}