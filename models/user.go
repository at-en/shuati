package models

import (
	"time"
	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Username      string    `json:"username" gorm:"unique;not null"`
	PasswordHash  string    `json:"-" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	LoginAttempts int       `json:"-" gorm:"default:0"`
	LockedUntil   *time.Time `json:"-"`
}

// UserRegister 用户注册请求
type UserRegister struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// UserLogin 用户登录请求
type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户响应数据
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// SetPassword 设置密码（加密存储）
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// IncrementLoginAttempts 增加登录失败次数
func (u *User) IncrementLoginAttempts() {
	u.LoginAttempts++
	if u.LoginAttempts >= 3 {
		lockTime := time.Now().Add(5 * time.Minute)
		u.LockedUntil = &lockTime
	}
}

// ResetLoginAttempts 重置登录失败次数
func (u *User) ResetLoginAttempts() {
	u.LoginAttempts = 0
	u.LockedUntil = nil
}